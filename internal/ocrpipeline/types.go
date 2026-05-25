package ocrpipeline

// StructuredPageOCR is the canonical page-level OCR representation used by
// the structured pipeline before deterministic Markdown rendering.
type StructuredPageOCR struct {
	SchemaVersion string     `json:"schema_version"`
	BookID        string     `json:"book_id"`
	PageNumber    int        `json:"page_number"`
	PageType      PageType   `json:"page_type"`
	Blocks        []OCRBlock `json:"blocks"`
	Warnings      []Warning  `json:"warnings,omitempty"`
}

type PageType string

const (
	PageTypeBlank        PageType = "blank"
	PageTypeTitle        PageType = "title"
	PageTypeFrontMatter  PageType = "front_matter"
	PageTypeTOC          PageType = "table_of_contents"
	PageTypeTOF          PageType = "table_of_figures"
	PageTypeBody         PageType = "body"
	PageTypeFigure       PageType = "figure"
	PageTypeTable        PageType = "table"
	PageTypeBibliography PageType = "bibliography"
)

type BlockType string

const (
	BlockHeading    BlockType = "heading"
	BlockParagraph  BlockType = "paragraph"
	BlockList       BlockType = "list"
	BlockTable      BlockType = "table"
	BlockFigure     BlockType = "figure"
	BlockFootnote   BlockType = "footnote"
	BlockPageFooter BlockType = "page_footer"
	BlockBlank      BlockType = "blank"
)

type OCRBlock struct {
	ID          string      `json:"id"`
	Type        BlockType   `json:"type"`
	Text        string      `json:"text,omitempty"`
	Level       int         `json:"level,omitempty"`
	Items       []ListItem  `json:"items,omitempty"`
	Table       *TableBlock `json:"table,omitempty"`
	Caption     string      `json:"caption,omitempty"`
	Description string      `json:"description,omitempty"`
	DiagramText []string    `json:"diagram_text,omitempty"`
	Confidence  string      `json:"confidence,omitempty"`
	Warnings    []Warning   `json:"warnings,omitempty"`
}

type ListItem struct {
	Text     string     `json:"text"`
	Children []ListItem `json:"children,omitempty"`
}

type TableBlock struct {
	Headers []string   `json:"headers,omitempty"`
	Rows    [][]string `json:"rows,omitempty"`
}

type Warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	BlockID string `json:"block_id,omitempty"`
}

type FigureRef struct {
	Path string `json:"path"`
	Alt  string `json:"alt,omitempty"`
}

type FigureResolver interface {
	RefFor(pageNumber int, blockID string) (FigureRef, bool)
}

type FigureMap map[int]map[string]FigureRef

func (m FigureMap) RefFor(pageNumber int, blockID string) (FigureRef, bool) {
	byBlock, ok := m[pageNumber]
	if !ok {
		return FigureRef{}, false
	}
	ref, ok := byBlock[blockID]
	return ref, ok
}
