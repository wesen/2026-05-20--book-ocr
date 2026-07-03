package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/book-ocr/internal/ocrpipeline"
	"github.com/go-go-golems/book-ocr/internal/ocrquality"
	"github.com/go-go-golems/geppetto/pkg/turns"
)

// StructuredOCRClient delegates whole-page OCR (seam S1, op ocr.page) to a
// plugin. The host still parses, repairs, validates, renders, and writes all
// artifacts — the plugin only supplies the StructuredPageOCR JSON.
type StructuredOCRClient struct {
	mgr      *Manager
	fallback ocrpipeline.StructuredOCRClient
}

var _ ocrpipeline.StructuredOCRClient = &StructuredOCRClient{}

func NewStructuredOCRClient(mgr *Manager, fallback ocrpipeline.StructuredOCRClient) *StructuredOCRClient {
	return &StructuredOCRClient{mgr: mgr, fallback: fallback}
}

func (c *StructuredOCRClient) OCRPage(ctx context.Context, input ocrpipeline.StructuredOCRInput, imageBytes []byte) (ocrpipeline.StructuredOCRResult, error) {
	if !c.mgr.Has(OpOCRPage) {
		if c.fallback == nil {
			return ocrpipeline.StructuredOCRResult{}, fmt.Errorf("no ocr.page plugin bound and no fallback client configured")
		}
		return c.fallback.OCRPage(ctx, input, imageBytes)
	}
	req := OCRPageInput{OpSchema: "ocr.page/v1", BookID: input.BookID, PageNumber: input.PageNumber, ImagePath: input.ImagePath, DryRun: input.DryRun}
	var out OCRPageOutput
	if err := c.mgr.Call(ctx, OpOCRPage, req, &out); err != nil {
		return ocrpipeline.StructuredOCRResult{}, fmt.Errorf("plugin ocr.page (%s): %w", c.mgr.PluginIDFor(OpOCRPage), err)
	}
	if len(out.Page) == 0 {
		return ocrpipeline.StructuredOCRResult{}, fmt.Errorf("plugin ocr.page (%s) returned no page object", c.mgr.PluginIDFor(OpOCRPage))
	}
	raw := strings.TrimSpace(out.RawResponse)
	if raw == "" {
		raw = string(out.Page)
	}

	// The audit turn records what the host actually handed the plugin: a
	// delegation note, the request payload, and the single target-page image
	// (which also keeps the exactly-one-image invariant checkable).
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return ocrpipeline.StructuredOCRResult{}, err
	}
	inputTurn := &turns.Turn{ID: ocrpipeline.PageTurnID(input.PageNumber, 1, "structured-ocr")}
	turns.AppendBlock(inputTurn, turns.NewSystemTextBlock(fmt.Sprintf("ocr.page delegated to plugin %q", c.mgr.PluginIDFor(OpOCRPage))))
	mediaType := mime.TypeByExtension(strings.ToLower(filepath.Ext(input.ImagePath)))
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	images := []map[string]any{{
		"media_type": mediaType,
		"content":    append([]byte(nil), imageBytes...),
		"detail":     "high",
		"role":       "target",
		"page":       input.PageNumber,
	}}
	turns.AppendBlock(inputTurn, turns.NewUserMultimodalBlock(string(reqJSON), images))

	finalTurn := &turns.Turn{ID: inputTurn.ID, Blocks: append([]turns.Block(nil), inputTurn.Blocks...)}
	turns.AppendBlock(finalTurn, turns.NewAssistantTextBlock(raw))

	return ocrpipeline.StructuredOCRResult{RawResponse: raw, InputTurn: inputTurn, FinalTurn: finalTurn}, nil
}

// PromptRenderer delegates prompt construction (seam S2, op prompt.render) to
// a plugin. The host appends the non-negotiable schema contract to the user
// prompt so a plugin cannot drop the JSON-only instruction.
type PromptRenderer struct {
	mgr *Manager
}

var _ ocrpipeline.PromptRenderer = &PromptRenderer{}

func NewPromptRenderer(mgr *Manager) *PromptRenderer {
	return &PromptRenderer{mgr: mgr}
}

func (p *PromptRenderer) RenderPrompts(input ocrpipeline.StructuredOCRInput) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req := PromptRenderInput{OpSchema: "prompt.render/v1", BookID: input.BookID, PageNumber: input.PageNumber, SchemaVersion: ocrpipeline.StructuredOCRSchemaVersion}
	var out PromptRenderOutput
	if err := p.mgr.Call(ctx, OpPromptRender, req, &out); err != nil {
		return "", "", fmt.Errorf("plugin prompt.render (%s): %w", p.mgr.PluginIDFor(OpPromptRender), err)
	}
	system := strings.TrimSpace(out.System)
	user := strings.TrimSpace(out.User)
	if system == "" || user == "" {
		return "", "", fmt.Errorf("plugin prompt.render (%s) returned empty system or user prompt", p.mgr.PluginIDFor(OpPromptRender))
	}
	user = user + "\n\n" + ocrpipeline.StructuredOCRSchemaContract(input)
	return system, user, nil
}

// FigureSegmenter delegates figure crop computation (seam S5, op
// figures.segment) to a plugin. The host performs the actual cropping and
// sidecar/debug writing so review artifacts stay uniform across strategies.
type FigureSegmenter struct {
	mgr *Manager
}

var _ ocrquality.FigureSegmenter = &FigureSegmenter{}

func NewFigureSegmenter(mgr *Manager) *FigureSegmenter {
	return &FigureSegmenter{mgr: mgr}
}

func (s *FigureSegmenter) SegmentFigure(req ocrquality.SegmentRequest) (image.Rectangle, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	in := FiguresSegmentInput{OpSchema: "figures.segment/v1", ImagePath: req.ImagePath, PageNumber: req.PageNumber, FigureIndex: req.FigureIndex, Description: req.Description, MarginPx: req.Margin}
	var out FiguresSegmentOutput
	if err := s.mgr.Call(ctx, OpFiguresSegment, in, &out); err != nil {
		return image.Rectangle{}, "", fmt.Errorf("plugin figures.segment (%s): %w", s.mgr.PluginIDFor(OpFiguresSegment), err)
	}
	if out.Crop.Width <= 0 || out.Crop.Height <= 0 {
		return image.Rectangle{}, "", fmt.Errorf("plugin figures.segment (%s) returned empty crop %+v", s.mgr.PluginIDFor(OpFiguresSegment), out.Crop)
	}
	method := strings.TrimSpace(out.Method)
	if method == "" {
		method = "plugin:" + s.mgr.PluginIDFor(OpFiguresSegment)
	}
	return image.Rect(out.Crop.X, out.Crop.Y, out.Crop.X+out.Crop.Width, out.Crop.Y+out.Crop.Height), method, nil
}
