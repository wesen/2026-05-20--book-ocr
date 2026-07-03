package ocrpipeline

import (
	"fmt"
	"regexp"
	"strings"
)

type RenderOptions struct {
	IncludeDiagramText bool
	IncludeFooters     bool
	WrapWidth          int
	// CodeLanguage is the fence language for code blocks; empty renders a
	// plain ``` fence.
	CodeLanguage string
	// SuppressTextualFigureCues: a figure whose caption/description contains
	// one of these is really textual content already transcribed elsewhere,
	// so its image/marker is suppressed (the caption is kept).
	SuppressTextualFigureCues []string
	// SuppressTableFigureCues: same, but only when the figure is immediately
	// followed by a table block that carries the content.
	SuppressTableFigureCues []string
	// EnableBoxedSetFallback turns "items A, B and C" figure descriptions
	// into a rendered {A, B, C} code block when the image is suppressed — a
	// Report-794-specific transform, off for generic books.
	EnableBoxedSetFallback bool
}

// DefaultRenderOptions is the Report-794 rendering policy the pipeline was
// built on; profile-driven runs override it (see RenderOptionsFromProfile).
func DefaultRenderOptions() RenderOptions {
	return RenderOptions{
		WrapWidth:                 88,
		CodeLanguage:              "common-lisp",
		SuppressTextualFigureCues: []string{"code listing", "code sample", "code block", "lisp-like definition", "boxed example", "boxed presentation"},
		SuppressTableFigureCues:   []string{"ppscalc", "spreadsheet", "formula display", "value display", "formula moved", "preparing to copy formula", "grid", "table"},
		EnableBoxedSetFallback:    true,
	}
}

func RenderPageMarkdown(page StructuredPageOCR, figures FigureResolver, opts RenderOptions) string {
	if opts.WrapWidth <= 0 {
		opts.WrapWidth = 88
	}
	var out strings.Builder
	fmt.Fprintf(&out, "<!-- page:%03d -->\n\n", page.PageNumber)
	if page.PageType == PageTypeBlank || len(page.Blocks) == 0 {
		out.WriteString("[BLANK PAGE]\n")
		return out.String()
	}
	for i, block := range page.Blocks {
		suppressFigureImage := false
		if block.Type == BlockFigure {
			suppressFigureImage = figureMatchesCues(block, opts.SuppressTextualFigureCues)
			if !suppressFigureImage && i+1 < len(page.Blocks) && page.Blocks[i+1].Type == BlockTable {
				suppressFigureImage = figureMatchesCues(block, opts.SuppressTableFigureCues)
			}
		}
		renderBlock(&out, page.PageNumber, block, figures, opts, suppressFigureImage)
	}
	return strings.TrimRight(out.String(), "\n") + "\n"
}

func renderBlock(out *strings.Builder, pageNumber int, block OCRBlock, figures FigureResolver, opts RenderOptions, suppressFigureImage bool) {
	switch block.Type {
	case BlockHeading:
		level := block.Level
		if level < 1 {
			level = 1
		}
		if level > 6 {
			level = 6
		}
		fmt.Fprintf(out, "%s %s\n\n", strings.Repeat("#", level), strings.TrimSpace(block.Text))
	case BlockParagraph:
		out.WriteString(WrapParagraph(block.Text, opts.WrapWidth))
		out.WriteString("\n\n")
	case BlockList:
		renderList(out, block.Items, 0)
		out.WriteString("\n")
	case BlockTable:
		renderTable(out, block.Table)
		out.WriteString("\n")
	case BlockCode:
		renderCode(out, block.Text, opts.CodeLanguage)
	case BlockFigure:
		renderFigure(out, pageNumber, block, figures, opts, suppressFigureImage)
	case BlockFootnote:
		fmt.Fprintf(out, "[^%s]: %s\n\n", block.ID, strings.TrimSpace(block.Text))
	case BlockPageFooter:
		if opts.IncludeFooters && strings.TrimSpace(block.Text) != "" {
			fmt.Fprintf(out, "<!-- footer: %s -->\n\n", strings.TrimSpace(block.Text))
		}
	case BlockBlank:
		out.WriteString("[BLANK PAGE]\n\n")
	default:
		if strings.TrimSpace(block.Text) != "" {
			out.WriteString(WrapParagraph(block.Text, opts.WrapWidth))
			out.WriteString("\n\n")
		}
	}
}

func figureMatchesCues(block OCRBlock, cues []string) bool {
	text := strings.ToLower(strings.Join([]string{block.Caption, block.Description}, " "))
	for _, cue := range cues {
		if cue != "" && strings.Contains(text, strings.ToLower(cue)) {
			return true
		}
	}
	return false
}

var boxedSetItemsRE = regexp.MustCompile(`(?i)items?\s+([A-Z0-9_ -]+(?:,\s*[A-Z0-9_ -]+)*(?:,?\s+and\s+[A-Z0-9_ -]+)?)`)

func textualFigureFallback(block OCRBlock) string {
	desc := strings.TrimSpace(block.Description)
	match := boxedSetItemsRE.FindStringSubmatch(desc)
	if len(match) == 2 {
		items := strings.TrimSpace(match[1])
		items = strings.ReplaceAll(items, " and ", ", ")
		items = strings.TrimSuffix(items, ".")
		return "{" + items + "}"
	}
	return ""
}

func renderCode(out *strings.Builder, text string, language string) {
	text = strings.TrimRight(text, "\n")
	if strings.TrimSpace(text) == "" {
		return
	}
	out.WriteString("```")
	out.WriteString(language)
	out.WriteString("\n")
	out.WriteString(text)
	out.WriteString("\n```\n\n")
}

func renderFigure(out *strings.Builder, pageNumber int, block OCRBlock, figures FigureResolver, opts RenderOptions, suppressImage bool) {
	caption := strings.TrimSpace(block.Caption)
	if caption != "" {
		out.WriteString(caption)
		out.WriteString("\n\n")
	}
	if suppressImage {
		// A neighboring block, or a simple textual fallback below, carries the readable
		// content of this figure. Keep the caption but avoid emitting a marker/image
		// that would duplicate searchable text/code/table content in review PDFs.
		if opts.EnableBoxedSetFallback {
			if fallback := textualFigureFallback(block); fallback != "" {
				renderCode(out, fallback, opts.CodeLanguage)
			}
		}
	} else if figures != nil {
		if ref, ok := figures.RefFor(pageNumber, block.ID); ok {
			alt := strings.TrimSpace(ref.Alt)
			if alt == "" {
				alt = strings.TrimSpace(block.Description)
			}
			if alt == "" {
				alt = caption
			}
			fmt.Fprintf(out, "![%s](%s)\n\n", alt, ref.Path)
		} else if strings.TrimSpace(block.Description) != "" {
			fmt.Fprintf(out, "[FIGURE: %s]\n\n", strings.TrimSpace(block.Description))
		}
	} else if strings.TrimSpace(block.Description) != "" {
		fmt.Fprintf(out, "[FIGURE: %s]\n\n", strings.TrimSpace(block.Description))
	}
	if opts.IncludeDiagramText && len(block.DiagramText) > 0 {
		out.WriteString("```text\n")
		for _, line := range block.DiagramText {
			out.WriteString(line)
			out.WriteByte('\n')
		}
		out.WriteString("```\n\n")
	}
}

func renderList(out *strings.Builder, items []ListItem, depth int) {
	indent := strings.Repeat("  ", depth)
	for _, item := range items {
		fmt.Fprintf(out, "%s- %s\n", indent, strings.TrimSpace(item.Text))
		if len(item.Children) > 0 {
			renderList(out, item.Children, depth+1)
		}
	}
}

func renderTable(out *strings.Builder, table *TableBlock) {
	if table == nil {
		return
	}
	cols := len(table.Headers)
	for _, row := range table.Rows {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if cols == 0 {
		return
	}
	headers := padCells(table.Headers, cols)
	if len(table.Headers) == 0 {
		headers = make([]string, cols)
		for i := range headers {
			headers[i] = fmt.Sprintf("Column %d", i+1)
		}
	}
	fmt.Fprintf(out, "| %s |\n", strings.Join(headers, " | "))
	sep := make([]string, cols)
	for i := range sep {
		sep[i] = "---"
	}
	fmt.Fprintf(out, "| %s |\n", strings.Join(sep, " | "))
	for _, row := range table.Rows {
		fmt.Fprintf(out, "| %s |\n", strings.Join(padCells(row, cols), " | "))
	}
}

func padCells(in []string, n int) []string {
	out := make([]string, n)
	for i := 0; i < n && i < len(in); i++ {
		out[i] = strings.TrimSpace(in[i])
	}
	return out
}

func WrapParagraph(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	if width <= 0 {
		return strings.Join(words, " ")
	}
	var out strings.Builder
	lineLen := 0
	for _, word := range words {
		if lineLen == 0 {
			out.WriteString(word)
			lineLen = len(word)
			continue
		}
		if lineLen+1+len(word) > width {
			out.WriteByte('\n')
			out.WriteString(word)
			lineLen = len(word)
			continue
		}
		out.WriteByte(' ')
		out.WriteString(word)
		lineLen += 1 + len(word)
	}
	return out.String()
}
