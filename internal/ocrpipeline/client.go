package ocrpipeline

import (
	"context"
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/geppetto/pkg/turns"
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
