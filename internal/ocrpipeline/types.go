package ocrpipeline

import (
	"encoding/json"
	"fmt"
	"strings"
)

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

func (p *StructuredPageOCR) UnmarshalJSON(data []byte) error {
	type alias StructuredPageOCR
	var raw struct {
		alias
		PageNumber any `json:"page_number"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = StructuredPageOCR(raw.alias)
	switch v := raw.PageNumber.(type) {
	case nil:
	case float64:
		p.PageNumber = int(v)
	case string:
		var n int
		if _, err := fmt.Sscanf(strings.TrimSpace(v), "%d", &n); err != nil {
			return fmt.Errorf("parse page_number %q: %w", v, err)
		}
		p.PageNumber = n
	default:
		var n int
		if _, err := fmt.Sscanf(fmt.Sprint(v), "%d", &n); err != nil {
			return fmt.Errorf("parse page_number %v: %w", v, err)
		}
		p.PageNumber = n
	}
	return nil
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
	BlockCode       BlockType = "code"
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

func (b *OCRBlock) UnmarshalJSON(data []byte) error {
	type alias OCRBlock
	var raw struct {
		alias
		DiagramText any `json:"diagram_text,omitempty"`
		Figure      *struct {
			Caption     string `json:"caption,omitempty"`
			Description string `json:"description,omitempty"`
			DiagramText any    `json:"diagram_text,omitempty"`
		} `json:"figure,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*b = OCRBlock(raw.alias)
	if raw.Figure != nil {
		if strings.TrimSpace(b.Caption) == "" {
			b.Caption = raw.Figure.Caption
		}
		if strings.TrimSpace(b.Description) == "" {
			b.Description = raw.Figure.Description
		}
		if raw.DiagramText == nil {
			raw.DiagramText = raw.Figure.DiagramText
		}
	}
	switch v := raw.DiagramText.(type) {
	case nil:
	case string:
		for _, line := range strings.Split(v, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				b.DiagramText = append(b.DiagramText, line)
			}
		}
	case []any:
		for _, item := range v {
			if s := strings.TrimSpace(fmt.Sprint(item)); s != "" {
				b.DiagramText = append(b.DiagramText, s)
			}
		}
	default:
		if s := strings.TrimSpace(fmt.Sprint(v)); s != "" {
			b.DiagramText = []string{s}
		}
	}
	return nil
}

type ListItem struct {
	Text     string     `json:"text"`
	Children []ListItem `json:"children,omitempty"`
}

func (i *ListItem) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		i.Text = text
		return nil
	}
	type alias ListItem
	var raw alias
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*i = ListItem(raw)
	return nil
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
