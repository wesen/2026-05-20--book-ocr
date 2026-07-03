package ocrpipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-go-golems/book-ocr/internal/ocrvalidation"
	"github.com/go-go-golems/geppetto/pkg/turns"
	"github.com/go-go-golems/geppetto/pkg/turns/serde"
	jsonsanitize "github.com/go-go-golems/sanitize/pkg/json"
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
	jsonText := normalizeStructuredJSONText(raw)
	if jsonText == "" {
		return StructuredPageOCR{}, fmt.Errorf("empty structured OCR response")
	}
	page, err := decodeStructuredPage(jsonText)
	if err != nil {
		sanitized := jsonsanitize.Sanitize(jsonText)
		if sanitized.StrictParseClean && strings.TrimSpace(sanitized.Sanitized) != "" {
			page, err = decodeStructuredPage(sanitized.Sanitized)
		}
	}
	if err != nil {
		return StructuredPageOCR{}, fmt.Errorf("parse structured OCR JSON: %w", err)
	}
	if page.SchemaVersion == "" {
		return StructuredPageOCR{}, fmt.Errorf("structured OCR response missing schema_version")
	}
	if page.PageNumber <= 0 {
		return StructuredPageOCR{}, fmt.Errorf("structured OCR response missing positive page_number")
	}
	repairStructuredPage(&page)
	return page, nil
}

func decodeStructuredPage(jsonText string) (StructuredPageOCR, error) {
	jsonText = fixLeadingZeroJSONNumbers(jsonText)
	var page StructuredPageOCR
	if err := json.Unmarshal([]byte(jsonText), &page); err != nil {
		return StructuredPageOCR{}, err
	}
	return page, nil
}

func normalizeStructuredJSONText(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```JSON")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)
	if start := strings.IndexByte(trimmed, '{'); start >= 0 {
		depth := 0
		inString := false
		escaped := false
		for i := start; i < len(trimmed); i++ {
			ch := trimmed[i]
			if inString {
				if escaped {
					escaped = false
					continue
				}
				if ch == '\\' {
					escaped = true
					continue
				}
				if ch == '"' {
					inString = false
				}
				continue
			}
			switch ch {
			case '"':
				inString = true
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					return strings.TrimSpace(trimmed[start : i+1])
				}
			}
		}
	}
	return trimmed
}

var leadingZeroJSONNumberRE = regexp.MustCompile(`(:\s*)0+(\d+)`)
var figureCaptionLineRE = regexp.MustCompile(`(?i)^\s*Figure\s+\d+(?:[-.]\d+)?\s*:`)

func fixLeadingZeroJSONNumbers(jsonText string) string {
	return leadingZeroJSONNumberRE.ReplaceAllString(jsonText, `${1}${2}`)
}

func repairStructuredPage(page *StructuredPageOCR) {
	for i := range page.Blocks {
		if strings.TrimSpace(page.Blocks[i].ID) == "" {
			page.Blocks[i].ID = fmt.Sprintf("p%03d-b%03d", page.PageNumber, i+1)
		}
		if page.Blocks[i].Type == BlockFigure && strings.TrimSpace(page.Blocks[i].Caption) == "" {
			if caption, ok := nearbyFigureCaption(page.Blocks, i); ok {
				page.Blocks[i].Caption = caption
			}
		}
	}
	page.Blocks = dropDuplicateFigureCaptionHeadings(page.Blocks)
}

func nearbyFigureCaption(blocks []OCRBlock, figureIndex int) (string, bool) {
	for _, idx := range []int{figureIndex - 1, figureIndex + 1} {
		if idx < 0 || idx >= len(blocks) {
			continue
		}
		candidate := strings.TrimSpace(blocks[idx].Text)
		if figureCaptionLineRE.MatchString(candidate) {
			return candidate, true
		}
	}
	return "", false
}

func dropDuplicateFigureCaptionHeadings(blocks []OCRBlock) []OCRBlock {
	out := make([]OCRBlock, 0, len(blocks))
	for i, block := range blocks {
		if block.Type == BlockHeading && figureCaptionLineRE.MatchString(strings.TrimSpace(block.Text)) {
			if i+1 < len(blocks) && blocks[i+1].Type == BlockFigure && NormalizeSpace(blocks[i+1].Caption) == NormalizeSpace(block.Text) {
				continue
			}
		}
		out = append(out, block)
	}
	return out
}

func NormalizeSpace(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
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
		"error":      filepath.Join(pageDir, "07-error.txt"),
	}

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
	if err := writeTurnYAML(paths["input"], result.InputTurn); err != nil {
		return StructuredPageRunResult{}, err
	}
	if err := writeTurnYAML(paths["final"], result.FinalTurn); err != nil {
		return StructuredPageRunResult{}, err
	}
	if err := os.WriteFile(paths["raw"], []byte(result.RawResponse+"\n"), 0o644); err != nil {
		return StructuredPageRunResult{}, err
	}

	parsed, err := ParseStructuredOCRResponse(result.RawResponse)
	if err != nil {
		_ = os.WriteFile(paths["error"], []byte(err.Error()+"\n"), 0o644)
		return StructuredPageRunResult{PageNumber: input.PageNumber, PageDir: pageDir, RawResponse: paths["raw"], TurnsDSN: turnStore.DSN()}, err
	}
	if parsed.PageNumber != input.PageNumber {
		err := fmt.Errorf("structured OCR page mismatch: input page %03d response page %03d", input.PageNumber, parsed.PageNumber)
		_ = os.WriteFile(paths["error"], []byte(err.Error()+"\n"), 0o644)
		return StructuredPageRunResult{PageNumber: input.PageNumber, PageDir: pageDir, RawResponse: paths["raw"], StructuredJSON: paths["structured"], TurnsDSN: turnStore.DSN()}, err
	}
	renderOpts := DefaultRenderOptions()
	if input.Render != nil {
		renderOpts = *input.Render
	}
	rendered := RenderPageMarkdown(parsed, nil, renderOpts)
	validation := ValidateStructuredPage(parsed, rendered)
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
	validation := StructuredValidation{PageNumber: page.PageNumber, FigureCaptions: ocrvalidation.ExtractFigureCaptions(rendered), RenderedBytes: len([]byte(rendered)), StructuredBlocks: len(page.Blocks)}
	for _, block := range page.Blocks {
		if block.Type == BlockFigure && strings.TrimSpace(block.Caption) == "" {
			validation.Warnings = append(validation.Warnings, ocrvalidation.Warning{Code: "figure_missing_caption", Page: page.PageNumber, Value: block.ID, Message: fmt.Sprintf("figure block %q has no caption", block.ID)})
		}
		if block.Type == BlockTable {
			if block.Table == nil {
				validation.Warnings = append(validation.Warnings, ocrvalidation.Warning{Code: "table_missing_payload", Page: page.PageNumber, Value: block.ID, Message: fmt.Sprintf("table block %q has no table payload", block.ID)})
			} else if len(block.Table.Headers) == 0 && len(block.Table.Rows) == 0 {
				validation.Warnings = append(validation.Warnings, ocrvalidation.Warning{Code: "table_empty", Page: page.PageNumber, Value: block.ID, Message: fmt.Sprintf("table block %q has no headers or rows", block.ID)})
			}
		}
		if block.Type == BlockCode && strings.TrimSpace(block.Text) == "" {
			validation.Warnings = append(validation.Warnings, ocrvalidation.Warning{Code: "code_empty", Page: page.PageNumber, Value: block.ID, Message: fmt.Sprintf("code block %q has no text", block.ID)})
		}
		if block.Type == BlockList && len(block.Items) == 0 {
			validation.Warnings = append(validation.Warnings, ocrvalidation.Warning{Code: "list_empty", Page: page.PageNumber, Value: block.ID, Message: fmt.Sprintf("list block %q has no items", block.ID)})
		}
	}
	return validation
}

func writeTurnYAML(path string, turn *turns.Turn) error {
	payload, err := serde.ToYAML(turn, serde.Options{})
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}
