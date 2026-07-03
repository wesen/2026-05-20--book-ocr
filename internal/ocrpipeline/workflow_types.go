package ocrpipeline

import "github.com/go-go-golems/book-ocr/internal/ocrvalidation"

const (
	StructuredPackageName     = "book-ocr/structured"
	StructuredProjectionName  = "book_ocr_structured"
	KindStructuredDiscover    = "book-ocr/structured/discover-pages"
	KindStructuredPage        = "book-ocr/structured/ocr-page"
	KindStructuredAssemble    = "book-ocr/structured/assemble-markdown"
	KindStructuredValidate    = "book-ocr/structured/validate-run"
	QueueStructuredControl    = "structured-control"
	QueueStructuredVision     = "structured-vision"
	QueueStructuredAssemble   = "structured-assemble"
	QueueStructuredValidation = "structured-validation"
)

type StructuredRunInput struct {
	BookID            string   `json:"book_id"`
	ImageDir          string   `json:"image_dir"`
	PageGlob          string   `json:"page_glob,omitempty"`
	StartPage         int      `json:"start_page,omitempty"`
	EndPage           int      `json:"end_page,omitempty"`
	WorkDir           string   `json:"work_dir"`
	RunID             string   `json:"run_id,omitempty"`
	Profile           string   `json:"profile,omitempty"`
	ProfileRegistries []string `json:"profile_registries,omitempty"`
	DryRun            bool     `json:"dry_run,omitempty"`
	ExpectedPages     int      `json:"expected_pages,omitempty"`
	MinRenderedBytes  int      `json:"min_rendered_bytes,omitempty"`
	EmbedFigures      bool     `json:"embed_figures,omitempty"`
	FiguresDir        string   `json:"figures_dir,omitempty"`
	RenderPDF         bool     `json:"render_pdf,omitempty"`
	PDFPath           string   `json:"pdf_path,omitempty"`
	PandocPath        string   `json:"pandoc_path,omitempty"`
	// BookProfile is an optional path to a book profile YAML. The discover
	// step resolves it and stamps the compiled prompt/render policy into
	// every page input, so the policy is persisted with the run and reruns
	// keep behaving identically.
	BookProfile string `json:"book_profile,omitempty"`
}

type StructuredPageWorkflowInput struct {
	BookID            string   `json:"book_id"`
	RunID             string   `json:"run_id"`
	PageNumber        int      `json:"page_number"`
	ImagePath         string   `json:"image_path"`
	WorkDir           string   `json:"work_dir"`
	TurnsDB           string   `json:"turns_db,omitempty"`
	TurnsDSN          string   `json:"turns_dsn,omitempty"`
	Profile           string   `json:"profile,omitempty"`
	ProfileRegistries []string `json:"profile_registries,omitempty"`
	DryRun            bool     `json:"dry_run,omitempty"`
	// Prompt and Render carry the compiled book-profile policy; nil means
	// the built-in (Report-794) defaults.
	Prompt *PromptSpec    `json:"prompt,omitempty"`
	Render *RenderOptions `json:"render,omitempty"`
	// PageTypeHint and Strategy come from the page.classify seam and persist
	// with the op so reruns route identically.
	PageTypeHint string `json:"page_type_hint,omitempty"`
	Strategy     string `json:"strategy,omitempty"`
}

type StructuredPageSpec struct {
	BookID     string `json:"book_id"`
	PageNumber int    `json:"page_number"`
	ImagePath  string `json:"image_path"`
}

type StructuredDiscoverResult struct {
	BookID      string               `json:"book_id"`
	PageCount   int                  `json:"page_count"`
	PageStepIDs []string             `json:"page_step_ids"`
	Pages       []StructuredPageSpec `json:"pages"`
}

type StructuredPageWorkflowResult struct {
	BookID                string `json:"book_id"`
	PageNumber            int    `json:"page_number"`
	PageDir               string `json:"page_dir"`
	RawResponsePath       string `json:"raw_response_path"`
	StructuredJSONPath    string `json:"structured_json_path"`
	RenderedMarkdownPath  string `json:"rendered_markdown_path"`
	ValidationJSONPath    string `json:"validation_json_path"`
	StructuredArtifactID  string `json:"structured_artifact_id,omitempty"`
	StructuredArtifactURI string `json:"structured_artifact_uri,omitempty"`
	RenderedArtifactID    string `json:"rendered_artifact_id,omitempty"`
	RenderedArtifactURI   string `json:"rendered_artifact_uri,omitempty"`
	ValidationArtifactID  string `json:"validation_artifact_id,omitempty"`
	ValidationArtifactURI string `json:"validation_artifact_uri,omitempty"`
	WarningCount          int    `json:"warning_count"`
	TableCount            int    `json:"table_count"`
	FigureCount           int    `json:"figure_count"`
	RenderedBytes         int    `json:"rendered_bytes"`
}

type StructuredAssembleInput struct {
	BookID       string `json:"book_id"`
	WorkDir      string `json:"work_dir"`
	ImageDir     string `json:"image_dir,omitempty"`
	EmbedFigures bool   `json:"embed_figures,omitempty"`
	FiguresDir   string `json:"figures_dir,omitempty"`
	RenderPDF    bool   `json:"render_pdf,omitempty"`
	PDFPath      string `json:"pdf_path,omitempty"`
	PandocPath   string `json:"pandoc_path,omitempty"`
}

type StructuredAssembleResult struct {
	BookID               string `json:"book_id"`
	PageCount            int    `json:"page_count"`
	MarkdownPath         string `json:"markdown_path"`
	MarkdownRefID        string `json:"markdown_ref_id,omitempty"`
	MarkdownURI          string `json:"markdown_uri,omitempty"`
	EmbeddedMarkdownPath string `json:"embedded_markdown_path,omitempty"`
	EmbeddedRefID        string `json:"embedded_ref_id,omitempty"`
	EmbeddedURI          string `json:"embedded_uri,omitempty"`
	FiguresDir           string `json:"figures_dir,omitempty"`
	FigureCount          int    `json:"figure_count,omitempty"`
	PDFPath              string `json:"pdf_path,omitempty"`
	PDFRefID             string `json:"pdf_ref_id,omitempty"`
	PDFURI               string `json:"pdf_uri,omitempty"`
	CharCount            int    `json:"char_count"`
}

type StructuredValidateInput struct {
	BookID           string `json:"book_id"`
	WorkDir          string `json:"work_dir"`
	ExpectedPages    int    `json:"expected_pages,omitempty"`
	MinRenderedBytes int    `json:"min_rendered_bytes,omitempty"`
}

type StructuredShortPageWarning struct {
	PageNumber           int    `json:"page_number"`
	RenderedBytes        int    `json:"rendered_bytes"`
	RenderedMarkdownPath string `json:"rendered_markdown_path,omitempty"`
}

type StructuredValidateResult struct {
	BookID                    string                       `json:"book_id"`
	PageCount                 int                          `json:"page_count"`
	ExpectedPages             int                          `json:"expected_pages,omitempty"`
	WarningCount              int                          `json:"warning_count"`
	AdjacentDuplicateCaptions []string                     `json:"adjacent_duplicate_captions,omitempty"`
	ShortPages                []StructuredShortPageWarning `json:"short_pages,omitempty"`
	PluginWarnings            []ocrvalidation.Warning      `json:"plugin_warnings,omitempty"`
	MinRenderedBytes          int                          `json:"min_rendered_bytes,omitempty"`
	ReportPath                string                       `json:"report_path"`
	ReportRefID               string                       `json:"report_ref_id,omitempty"`
	ReportURI                 string                       `json:"report_uri,omitempty"`
}
