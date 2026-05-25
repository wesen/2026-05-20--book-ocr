package vlmseparation

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/geppetto/pkg/turns"
)

func BuildTurnForScenario(input TrialInput) (*turns.Turn, error) {
	turn := &turns.Turn{ID: input.TurnID}
	turns.AppendBlock(turn, turns.NewSystemTextBlock(benchmarkSystemPrompt()))
	switch input.Scenario.Name {
	case ScenarioTargetOnly:
		img, err := imagePayload(input.TargetImagePath, "target", input.TargetPage)
		if err != nil {
			return nil, err
		}
		turns.AppendBlock(turn, turns.NewUserMultimodalBlock(targetOnlyPrompt(input), []map[string]any{img}))
	case ScenarioSingleBlockTargetFirst:
		images, err := contextImagePayloads(input, []string{"target", "previous", "next"})
		if err != nil {
			return nil, err
		}
		turns.AppendBlock(turn, turns.NewUserMultimodalBlock(singleBlockPrompt(input), images))
	case ScenarioSingleBlockLabeledImages:
		images, err := contextImagePayloads(input, []string{"target", "previous", "next"})
		if err != nil {
			return nil, err
		}
		turns.AppendBlock(turn, turns.NewUserMultimodalBlock(singleBlockLabeledPrompt(input), images))
	case ScenarioMultiBlockLabeled:
		if err := appendMultiBlockLabeled(turn, input); err != nil {
			return nil, err
		}
	case ScenarioContextFirstNegativeControl:
		images, err := contextImagePayloads(input, []string{"previous", "next", "target"})
		if err != nil {
			return nil, err
		}
		turns.AppendBlock(turn, turns.NewUserMultimodalBlock(contextFirstPrompt(input), images))
	case ScenarioTargetPlusTextContext:
		turns.AppendBlock(turn, turns.NewUserTextBlock("Text-only context from neighboring pages. Use for terminology/style only; do not copy unless visible on target page.\n\nPrevious context:\n"+input.PreviousText+"\n\nNext context:\n"+input.NextText))
		img, err := imagePayload(input.TargetImagePath, "target", input.TargetPage)
		if err != nil {
			return nil, err
		}
		turns.AppendBlock(turn, turns.NewUserMultimodalBlock(targetOnlyPrompt(input), []map[string]any{img}))
	default:
		return nil, fmt.Errorf("unsupported scenario %q", input.Scenario.Name)
	}
	_ = turns.KeyTurnMetaSessionID.Set(&turn.Metadata, input.SessionID)
	_ = turns.KeyTurnMetaRuntime.Set(&turn.Metadata, "book-ocr/vlm-separation/"+input.Scenario.Name)
	return turn, nil
}

func appendMultiBlockLabeled(turn *turns.Turn, input TrialInput) error {
	turns.AppendBlock(turn, turns.NewUserTextBlock("The next image block is the ONLY OCR target. Transcribe only that target page."))
	img, err := imagePayload(input.TargetImagePath, "target", input.TargetPage)
	if err != nil {
		return err
	}
	turns.AppendBlock(turn, turns.NewUserMultimodalBlock(fmt.Sprintf("TARGET PAGE %03d. OCR this image only.", input.TargetPage), []map[string]any{img}))
	turns.AppendBlock(turn, turns.NewUserTextBlock("The following image blocks are context only. Do not transcribe text, captions, labels, or figures from them."))
	if input.PreviousImagePath != "" {
		prev, err := imagePayload(input.PreviousImagePath, "context_previous", input.PreviousPage)
		if err != nil {
			return err
		}
		turns.AppendBlock(turn, turns.NewUserMultimodalBlock(fmt.Sprintf("PREVIOUS CONTEXT PAGE %03d. Do not transcribe this image.", input.PreviousPage), []map[string]any{prev}))
	}
	if input.NextImagePath != "" {
		next, err := imagePayload(input.NextImagePath, "context_next", input.NextPage)
		if err != nil {
			return err
		}
		turns.AppendBlock(turn, turns.NewUserMultimodalBlock(fmt.Sprintf("NEXT CONTEXT PAGE %03d. Do not transcribe this image.", input.NextPage), []map[string]any{next}))
	}
	turns.AppendBlock(turn, turns.NewUserTextBlock("Now return JSON for TARGET PAGE ONLY using the benchmark schema."))
	return nil
}

func contextImagePayloads(input TrialInput, order []string) ([]map[string]any, error) {
	var images []map[string]any
	for _, item := range order {
		switch item {
		case "target":
			img, err := imagePayload(input.TargetImagePath, "target", input.TargetPage)
			if err != nil {
				return nil, err
			}
			images = append(images, img)
		case "previous":
			if input.PreviousImagePath == "" {
				continue
			}
			img, err := imagePayload(input.PreviousImagePath, "context_previous", input.PreviousPage)
			if err != nil {
				return nil, err
			}
			images = append(images, img)
		case "next":
			if input.NextImagePath == "" {
				continue
			}
			img, err := imagePayload(input.NextImagePath, "context_next", input.NextPage)
			if err != nil {
				return nil, err
			}
			images = append(images, img)
		}
	}
	return images, nil
}

func imagePayload(path string, role string, page int) (map[string]any, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"media_type": mediaTypeFromPath(path),
		"content":    body,
		"detail":     "high",
		"role":       role,
		"page":       page,
	}, nil
}

func mediaTypeFromPath(path string) string {
	if mediaType := mime.TypeByExtension(strings.ToLower(filepath.Ext(path))); mediaType != "" {
		return mediaType
	}
	return "application/octet-stream"
}

func benchmarkSystemPrompt() string {
	return `You are a benchmarked vision-language OCR engine. Follow target/context boundaries exactly. Return strict JSON only.`
}

func targetOnlyPrompt(input TrialInput) string {
	return fmt.Sprintf(`OCR target page %03d only. Return JSON with fields target_page, transcribed_page_identity, content_markers, transcription, suspected_context_copy, and notes. Do not include markdown fences.`, input.TargetPage)
}

func singleBlockPrompt(input TrialInput) string {
	return fmt.Sprintf(`You will receive multiple images in one block. The first image is TARGET PAGE %03d. Other images are context only. Transcribe only the target page. Return strict JSON only.`, input.TargetPage)
}

func singleBlockLabeledPrompt(input TrialInput) string {
	return fmt.Sprintf(`Images include metadata roles: target, context_previous, context_next. Only role=target page=%03d may be transcribed. If any context-only content appears in your output, set suspected_context_copy=true. Return strict JSON only.`, input.TargetPage)
}

func contextFirstPrompt(input TrialInput) string {
	return fmt.Sprintf(`This is a negative-control layout. Context images may appear before the target. The ONLY target is page %03d. Transcribe only that page and return strict JSON.`, input.TargetPage)
}
