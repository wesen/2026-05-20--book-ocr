package ocrpipeline

import (
	"fmt"
	"strings"
)

type RenderOptions struct {
	IncludeDiagramText bool
	IncludeFooters     bool
	WrapWidth          int
}

func DefaultRenderOptions() RenderOptions {
	return RenderOptions{WrapWidth: 88}
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
	for _, block := range page.Blocks {
		renderBlock(&out, page.PageNumber, block, figures, opts)
	}
	return strings.TrimRight(out.String(), "\n") + "\n"
}

func renderBlock(out *strings.Builder, pageNumber int, block OCRBlock, figures FigureResolver, opts RenderOptions) {
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
	case BlockFigure:
		renderFigure(out, pageNumber, block, figures, opts)
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

func renderFigure(out *strings.Builder, pageNumber int, block OCRBlock, figures FigureResolver, opts RenderOptions) {
	caption := strings.TrimSpace(block.Caption)
	if caption != "" {
		out.WriteString(caption)
		out.WriteString("\n\n")
	}
	if figures != nil {
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
