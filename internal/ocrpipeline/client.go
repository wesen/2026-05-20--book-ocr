package ocrpipeline

import (
	"context"
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/geppetto/pkg/turns"
	"github.com/go-go-golems/pinocchio/pkg/cmds/profilebootstrap"
)

type StructuredOCRInput struct {
	BookID            string
	RunID             string
	PageNumber        int
	ImagePath         string
	WorkDir           string
	TurnsDB           string
	TurnsDSN          string
	Profile           string
	ProfileRegistries []string
	DryRun            bool
}

type StructuredOCRResult struct {
	Page        StructuredPageOCR
	RawResponse string
	InputTurn   *turns.Turn
	FinalTurn   *turns.Turn
}

type StructuredOCRClient interface {
	OCRPage(ctx context.Context, input StructuredOCRInput, imageBytes []byte) (StructuredOCRResult, error)
}

type GeppettoStructuredOCRClient struct{}

func NewGeppettoStructuredOCRClient() *GeppettoStructuredOCRClient {
	return &GeppettoStructuredOCRClient{}
}

func (c *GeppettoStructuredOCRClient) OCRPage(ctx context.Context, input StructuredOCRInput, imageBytes []byte) (StructuredOCRResult, error) {
	inputTurn, err := BuildStructuredOCRInputTurn(input, imageBytes)
	if err != nil {
		return StructuredOCRResult{}, err
	}
	parsed, err := profilebootstrap.NewCLISelectionValues(profilebootstrap.CLISelectionInput{
		Profile:           input.Profile,
		ProfileRegistries: append([]string(nil), input.ProfileRegistries...),
	})
	if err != nil {
		return StructuredOCRResult{}, fmt.Errorf("build pinocchio profile selection: %w", err)
	}
	resolved, err := profilebootstrap.ResolveCLIEngineSettings(ctx, parsed)
	if err != nil {
		return StructuredOCRResult{}, fmt.Errorf("resolve pinocchio profile-backed engine settings: %w", err)
	}
	if resolved.Close != nil {
		defer resolved.Close()
	}
	eng, err := profilebootstrap.NewEngineFromResolvedCLIEngineSettings(resolved)
	if err != nil {
		return StructuredOCRResult{}, fmt.Errorf("build geppetto engine from pinocchio profile: %w", err)
	}
	finalTurn, err := eng.RunInference(ctx, inputTurn)
	if err != nil {
		return StructuredOCRResult{}, fmt.Errorf("run structured OCR inference: %w", err)
	}
	raw, err := lastLLMText(finalTurn)
	if err != nil {
		return StructuredOCRResult{}, err
	}
	return StructuredOCRResult{RawResponse: raw, InputTurn: inputTurn, FinalTurn: finalTurn}, nil
}

func BuildStructuredOCRInputTurn(input StructuredOCRInput, imageBytes []byte) (*turns.Turn, error) {
	if strings.TrimSpace(input.BookID) == "" {
		return nil, fmt.Errorf("bookID is required")
	}
	if input.PageNumber <= 0 {
		return nil, fmt.Errorf("page number must be positive")
	}
	if strings.TrimSpace(input.ImagePath) == "" {
		return nil, fmt.Errorf("image path is required")
	}
	if len(imageBytes) == 0 {
		return nil, fmt.Errorf("image bytes are required")
	}
	turnID := PageTurnID(input.PageNumber, 1, "structured-ocr")
	turn := &turns.Turn{ID: turnID}
	turns.AppendBlock(turn, turns.NewSystemTextBlock(StructuredOCRSystemPrompt))
	images := []map[string]any{{
		"media_type": mediaTypeFromImagePath(input.ImagePath),
		"content":    append([]byte(nil), imageBytes...),
		"detail":     "high",
		"role":       "target",
		"page":       input.PageNumber,
	}}
	turns.AppendBlock(turn, turns.NewUserMultimodalBlock(RenderStructuredOCRPrompt(input), images))
	return turn, nil
}

func mediaTypeFromImagePath(path string) string {
	if mediaType := mime.TypeByExtension(strings.ToLower(filepath.Ext(path))); mediaType != "" {
		return mediaType
	}
	return "application/octet-stream"
}

func CountTurnImages(turn *turns.Turn) int {
	if turn == nil {
		return 0
	}
	count := 0
	for _, block := range turn.Blocks {
		for _, value := range block.Payload {
			count += countImagesInValue(value)
		}
	}
	return count
}

func countImagesInValue(value any) int {
	switch v := value.(type) {
	case []map[string]any:
		return len(v)
	case []any:
		count := 0
		for _, item := range v {
			count += countImagesInValue(item)
		}
		return count
	case map[string]any:
		if _, ok := v["media_type"]; ok {
			if _, ok := v["content"]; ok {
				return 1
			}
		}
		count := 0
		for _, item := range v {
			count += countImagesInValue(item)
		}
		return count
	default:
		return 0
	}
}

func lastLLMText(turn *turns.Turn) (string, error) {
	if turn == nil {
		return "", fmt.Errorf("nil geppetto turn")
	}
	blocks := turns.FindLastBlocksByKind(*turn, turns.BlockKindLLMText)
	if len(blocks) == 0 {
		return "", fmt.Errorf("structured OCR response did not include an LLM text block")
	}
	text, _ := blocks[len(blocks)-1].Payload[turns.PayloadKeyText].(string)
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("structured OCR response was empty")
	}
	return text, nil
}
