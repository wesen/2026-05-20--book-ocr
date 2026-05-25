package ocrpipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/book-ocr/internal/ocrvalidation"
	"github.com/go-go-golems/geppetto/pkg/turns"
	"github.com/go-go-golems/geppetto/pkg/turns/serde"
)

type StructuredPageRunResult struct {
	PageNumber     int                  `json:"page_number"`
	PageDir        string               `json:"page_dir"`
	RawResponse    string               `json:"raw_response_path"`
	StructuredJSON string               `json:"structured_json_path"`
	RenderedMD     string               `json:"rendered_markdown_path"`
	ValidationJSON string               `json:"validation_json_path"`
	TurnsDSN       string               `json:"turns_dsn,omitempty"`
	Validation     StructuredValidation `json:"validation"`
}

type StructuredValidation struct {
	PageNumber       int                         `json:"page_number"`
	FigureCaptions   []string                    `json:"figure_captions,omitempty"`
	Warnings         []ocrvalidation.Warning     `json:"warnings,omitempty"`
	AnchorResult     *ocrvalidation.AnchorResult `json:"anchor_result,omitempty"`
	RenderedBytes    int                         `json:"rendered_bytes"`
	StructuredBlocks int                         `json:"structured_blocks"`
}

type DryRunStructuredOCRClient struct{}

func (DryRunStructuredOCRClient) OCRPage(ctx context.Context, input StructuredOCRInput, imageBytes []byte) (StructuredOCRResult, error) {
	_ = ctx
	inputTurn, err := BuildStructuredOCRInputTurn(input, imageBytes)
	if err != nil {
		return StructuredOCRResult{}, err
	}
	page := FakeStructuredPage(input.BookID, input.PageNumber)
	raw, err := json.MarshalIndent(page, "", "  ")
	if err != nil {
		return StructuredOCRResult{}, err
	}
	finalTurn, err := BuildStructuredOCRInputTurn(input, imageBytes)
	if err != nil {
		return StructuredOCRResult{}, err
	}
	turns.AppendBlock(finalTurn, turns.NewAssistantTextBlock(string(raw)))
	return StructuredOCRResult{Page: page, RawResponse: string(raw), InputTurn: inputTurn, FinalTurn: finalTurn}, nil
}

func FakeStructuredPage(bookID string, pageNumber int) StructuredPageOCR {
	page := StructuredPageOCR{SchemaVersion: StructuredOCRSchemaVersion, BookID: bookID, PageNumber: pageNumber, PageType: PageTypeBody}
	switch pageNumber {
	case 32:
		page.PageType = PageTypeFigure
		page.Blocks = []OCRBlock{
			{ID: "p032-f001", Type: BlockFigure, Caption: "Figure 2-2: PPSCalc -- Formula Display", Description: "PPSCalc formula display."},
			{ID: "p032-t001", Type: BlockTable, Table: &TableBlock{Headers: []string{"", "A", "B", "C"}, Rows: [][]string{{"1", "100", "20", "A1*B1"}, {"2", "75", "5", "A2*B2"}, {"3", "", "", "C1+C2"}}}},
			{ID: "p032-f002", Type: BlockFigure, Caption: "Figure 2-3: PPSCalc -- Value Display", Description: "PPSCalc value display."},
			{ID: "p032-t002", Type: BlockTable, Table: &TableBlock{Headers: []string{"", "A", "B", "C"}, Rows: [][]string{{"1", "100", "20", "2000"}, {"2", "75", "5", "375"}, {"3", "", "", "2375"}}}},
		}
	case 12:
		page.Blocks = []OCRBlock{{ID: "p012-p001", Type: BlockParagraph, Text: "The rudimentary user interface is introduced before the full figure is shown on the next page."}}
	case 13:
		page.PageType = PageTypeFigure
		page.Blocks = []OCRBlock{{ID: "p013-f001", Type: BlockFigure, Caption: "Figure 1-1: A Rudimentary User Interface", Description: "Diagram showing a user interface connected to an application data base."}}
	default:
		page.Blocks = []OCRBlock{{ID: fmt.Sprintf("p%03d-p001", pageNumber), Type: BlockParagraph, Text: fmt.Sprintf("Dry-run structured OCR placeholder for page %03d.", pageNumber)}}
	}
	return page
}

func ParseStructuredOCRResponse(raw string) (StructuredPageOCR, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return StructuredPageOCR{}, fmt.Errorf("empty structured OCR response")
	}
	var page StructuredPageOCR
	if err := json.Unmarshal([]byte(trimmed), &page); err != nil {
		return StructuredPageOCR{}, fmt.Errorf("parse structured OCR JSON: %w", err)
	}
	if page.SchemaVersion == "" {
		return StructuredPageOCR{}, fmt.Errorf("structured OCR response missing schema_version")
	}
	if page.PageNumber <= 0 {
		return StructuredPageOCR{}, fmt.Errorf("structured OCR response missing positive page_number")
	}
	return page, nil
}

func RunStructuredPage(ctx context.Context, input StructuredOCRInput, client StructuredOCRClient) (StructuredPageRunResult, error) {
	if client == nil {
		return StructuredPageRunResult{}, fmt.Errorf("structured OCR client is required")
	}
	if strings.TrimSpace(input.WorkDir) == "" {
		return StructuredPageRunResult{}, fmt.Errorf("workDir is required")
	}
	imageBytes, err := os.ReadFile(input.ImagePath)
	if err != nil {
		return StructuredPageRunResult{}, fmt.Errorf("read target image: %w", err)
	}
	if strings.TrimSpace(input.RunID) == "" {
		input.RunID = "structured-page"
	}
	turnStore, closeFn, err := OpenOCRTurnStore(TurnStoreConfig{TurnsDSN: input.TurnsDSN, TurnsDB: input.TurnsDB, WorkDir: input.WorkDir, ConvID: BookRunConvID(input.BookID, input.RunID)})
	if err != nil {
		return StructuredPageRunResult{}, err
	}
	defer closeFn()

	result, err := client.OCRPage(ctx, input, imageBytes)
	if err != nil {
		return StructuredPageRunResult{}, err
	}
	if CountTurnImages(result.InputTurn) != 1 {
		return StructuredPageRunResult{}, fmt.Errorf("structured OCR input turn must contain exactly one image, got %d", CountTurnImages(result.InputTurn))
	}
	sessionID := PageSessionID(input.PageNumber)
	turnID := PageTurnID(input.PageNumber, 1, "structured-ocr")
	if err := turnStore.Save(ctx, sessionID, turnID, "input", result.InputTurn); err != nil {
		return StructuredPageRunResult{}, err
	}
	if err := turnStore.Save(ctx, sessionID, turnID, "final", result.FinalTurn); err != nil {
		return StructuredPageRunResult{}, err
	}
	parsed, err := ParseStructuredOCRResponse(result.RawResponse)
	if err != nil {
		return StructuredPageRunResult{}, err
	}
	if parsed.PageNumber != input.PageNumber {
		return StructuredPageRunResult{}, fmt.Errorf("structured OCR page mismatch: input page %03d response page %03d", input.PageNumber, parsed.PageNumber)
	}
	rendered := RenderPageMarkdown(parsed, nil, DefaultRenderOptions())
	validation := ValidateStructuredPage(parsed, rendered)

	pageDir := filepath.Join(input.WorkDir, "pages", fmt.Sprintf("page_%03d", input.PageNumber))
	if err := os.MkdirAll(pageDir, 0o755); err != nil {
		return StructuredPageRunResult{}, err
	}
	paths := map[string]string{
		"input":      filepath.Join(pageDir, "01-turn-input.yaml"),
		"final":      filepath.Join(pageDir, "02-turn-final.yaml"),
		"raw":        filepath.Join(pageDir, "03-raw-response.json"),
		"structured": filepath.Join(pageDir, "04-structured.json"),
		"rendered":   filepath.Join(pageDir, "05-rendered.md"),
		"validation": filepath.Join(pageDir, "06-validation.json"),
	}
	if err := writeTurnYAML(paths["input"], result.InputTurn); err != nil {
		return StructuredPageRunResult{}, err
	}
	if err := writeTurnYAML(paths["final"], result.FinalTurn); err != nil {
		return StructuredPageRunResult{}, err
	}
	if err := os.WriteFile(paths["raw"], []byte(result.RawResponse+"\n"), 0o644); err != nil {
		return StructuredPageRunResult{}, err
	}
	structuredJSON, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return StructuredPageRunResult{}, err
	}
	if err := os.WriteFile(paths["structured"], append(structuredJSON, '\n'), 0o644); err != nil {
		return StructuredPageRunResult{}, err
	}
	if err := os.WriteFile(paths["rendered"], []byte(rendered), 0o644); err != nil {
		return StructuredPageRunResult{}, err
	}
	validationJSON, err := json.MarshalIndent(validation, "", "  ")
	if err != nil {
		return StructuredPageRunResult{}, err
	}
	if err := os.WriteFile(paths["validation"], append(validationJSON, '\n'), 0o644); err != nil {
		return StructuredPageRunResult{}, err
	}
	return StructuredPageRunResult{PageNumber: input.PageNumber, PageDir: pageDir, RawResponse: paths["raw"], StructuredJSON: paths["structured"], RenderedMD: paths["rendered"], ValidationJSON: paths["validation"], TurnsDSN: turnStore.DSN(), Validation: validation}, nil
}

func ValidateStructuredPage(page StructuredPageOCR, rendered string) StructuredValidation {
	return StructuredValidation{PageNumber: page.PageNumber, FigureCaptions: ocrvalidation.ExtractFigureCaptions(rendered), RenderedBytes: len([]byte(rendered)), StructuredBlocks: len(page.Blocks)}
}

func writeTurnYAML(path string, turn *turns.Turn) error {
	payload, err := serde.ToYAML(turn, serde.Options{})
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}
