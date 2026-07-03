// Package plugin hosts NDJSON-stdio plugins (devctl plugin protocol v2) and
// adapts them onto book-ocr's strategy seams. Plugins replace strategies,
// never invariants: the host keeps enforcing the single-target-image rule,
// schema validation, artifact writing, and retry classification regardless of
// what a plugin returns.
package plugin

import "encoding/json"

// Seam op names. A plugin advertises the ops it implements in its handshake
// capabilities; a binding (profile plugins: entry or --plugin flag) may only
// claim seams the plugin actually advertises.
const (
	OpOCRPage        = "ocr.page"
	OpPromptRender   = "prompt.render"
	OpFiguresSegment = "figures.segment"
)

// KnownSeams lists every op the host can dispatch today.
var KnownSeams = []string{OpOCRPage, OpPromptRender, OpFiguresSegment}

// OCRPageInput is the ocr.page request payload (op schema ocr.page/v1). The
// image travels by path, not bytes: plugins are local processes and open the
// file themselves.
type OCRPageInput struct {
	OpSchema   string `json:"op_schema"`
	BookID     string `json:"book_id"`
	PageNumber int    `json:"page_number"`
	ImagePath  string `json:"image_path"`
	DryRun     bool   `json:"dry_run"`
}

// OCRPageOutput is the ocr.page response payload. Page must be a
// StructuredPageOCR JSON object (schema structured-ocr/v1); RawResponse
// optionally preserves the underlying model text for LLM-backed plugins so
// the 03-raw-response.json artifact stays meaningful; Engine is free-form
// provenance recorded with the page.
type OCRPageOutput struct {
	Page        json.RawMessage `json:"page"`
	RawResponse string          `json:"raw_response,omitempty"`
	Engine      map[string]any  `json:"engine,omitempty"`
}

// PromptRenderInput is the prompt.render request payload (prompt.render/v1).
type PromptRenderInput struct {
	OpSchema      string `json:"op_schema"`
	BookID        string `json:"book_id"`
	PageNumber    int    `json:"page_number"`
	SchemaVersion string `json:"schema_version"`
}

// PromptRenderOutput carries the prompts to use for the page. The host owns
// the non-negotiable schema contract and appends it to User itself, so a
// plugin cannot accidentally drop the JSON-only instruction.
type PromptRenderOutput struct {
	System string `json:"system"`
	User   string `json:"user"`
}

// FiguresSegmentInput is the figures.segment request payload
// (figures.segment/v1), one call per figure marker.
type FiguresSegmentInput struct {
	OpSchema    string `json:"op_schema"`
	ImagePath   string `json:"image_path"`
	PageNumber  int    `json:"page_number"`
	FigureIndex int    `json:"figure_index"`
	Description string `json:"description"`
	MarginPx    int    `json:"margin_px"`
}

// CropRect is a pixel rectangle in page-image coordinates.
type CropRect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// FiguresSegmentOutput is the figures.segment response payload.
type FiguresSegmentOutput struct {
	Crop       CropRect `json:"crop"`
	Method     string   `json:"method,omitempty"`
	Confidence float64  `json:"confidence,omitempty"`
}
