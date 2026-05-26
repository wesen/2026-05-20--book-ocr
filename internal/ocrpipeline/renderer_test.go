package ocrpipeline

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderPageMarkdownSuppressesTextualBoxedFigureWithFallback(t *testing.T) {
	page := StructuredPageOCR{SchemaVersion: "structured-ocr/v1", BookID: "report-794", PageNumber: 121, PageType: PageTypeBody, Blocks: []OCRBlock{
		{ID: "p121-f001", Type: BlockFigure, Description: "Boxed presentation of a set with items ONE, TWO, THREE."},
	}}
	figures := FigureMap{121: {"p121-f001": {Path: "figures/page_121_figure_01.png"}}}
	got := RenderPageMarkdown(page, figures, DefaultRenderOptions())
	require.Contains(t, got, "```text\n{ONE, TWO, THREE}\n```")
	require.NotContains(t, got, "page_121_figure_01.png")
	require.NotContains(t, got, "[FIGURE:")
}

func TestRenderPageMarkdownSuppressesSpreadsheetFigureImageWhenTableFollows(t *testing.T) {
	page := StructuredPageOCR{SchemaVersion: "structured-ocr/v1", BookID: "report-794", PageNumber: 48, PageType: PageTypeTable, Blocks: []OCRBlock{
		{ID: "p048-f001", Type: BlockFigure, Caption: "Figure 2-12: PPSCalc -- Formula Moved", Description: "Spreadsheet showing columns A B C and rows 1-3 with values and formulas."},
		{ID: "p048-t001", Type: BlockTable, Table: &TableBlock{Headers: []string{"", "A", "B", "C"}, Rows: [][]string{{"1", "100", "20", "A1*B1"}}}},
	}}
	figures := FigureMap{48: {"p048-f001": {Path: "figures/page_048_figure_01.png"}}}
	got := RenderPageMarkdown(page, figures, DefaultRenderOptions())
	require.Contains(t, got, "Figure 2-12: PPSCalc -- Formula Moved")
	require.Contains(t, got, "| 1 | 100 | 20 | A1*B1 |")
	require.NotContains(t, got, "page_048_figure_01.png")
	require.NotContains(t, got, "[FIGURE:")
}

func TestRenderPageMarkdownBodyAndFigure(t *testing.T) {
	page := StructuredPageOCR{SchemaVersion: "structured-ocr/v1", BookID: "report-794", PageNumber: 13, PageType: PageTypeFigure, Blocks: []OCRBlock{
		{ID: "p013-h001", Type: BlockHeading, Level: 2, Text: "Chapter 1"},
		{ID: "p013-p001", Type: BlockParagraph, Text: "This paragraph is intentionally long enough to exercise deterministic wrapping by the renderer without requiring a model to decide where line breaks belong."},
		{ID: "p013-f001", Type: BlockFigure, Caption: "Figure 1-1: A Rudimentary User Interface", Description: "Diagram showing users and an application data base.", DiagramText: []string{"User", "Application Data Base"}},
	}}
	figures := FigureMap{13: {"p013-f001": {Path: "figures/page_013_figure_01.png"}}}
	got := RenderPageMarkdown(page, figures, DefaultRenderOptions())
	require.Contains(t, got, "<!-- page:013 -->")
	require.Contains(t, got, "## Chapter 1")
	require.Contains(t, got, "Figure 1-1: A Rudimentary User Interface")
	require.Contains(t, got, "![Diagram showing users and an application data base.](figures/page_013_figure_01.png)")
	require.NotContains(t, got, "Application Data Base\n```")
	for _, line := range strings.Split(got, "\n") {
		require.LessOrEqual(t, len(line), 110)
	}
}

func TestRenderPageMarkdownOmitsFooterByDefault(t *testing.T) {
	page := StructuredPageOCR{BookID: "report-794", PageNumber: 116, PageType: PageTypeBody, Blocks: []OCRBlock{
		{ID: "p116-p001", Type: BlockParagraph, Text: "5.2 Graphics Redisplay"},
		{ID: "p116-footer", Type: BlockPageFooter, Text: "114"},
	}}
	got := RenderPageMarkdown(page, nil, DefaultRenderOptions())
	require.Contains(t, got, "5.2 Graphics Redisplay")
	require.NotContains(t, got, "footer")
}

func TestRenderPageMarkdownTableCodeAndBlank(t *testing.T) {
	tablePage := StructuredPageOCR{PageNumber: 32, PageType: PageTypeTable, Blocks: []OCRBlock{
		{ID: "p032-t001", Type: BlockTable, Table: &TableBlock{Headers: []string{"A", "B", "C"}, Rows: [][]string{{"100", "20", "A1*B1"}, {"75", "5", "A2*B2"}}}},
		{ID: "p032-c001", Type: BlockCode, Text: "for each cell:\n  recompute(value)"},
	}}
	got := RenderPageMarkdown(tablePage, nil, DefaultRenderOptions())
	require.Contains(t, got, "| A | B | C |")
	require.Contains(t, got, "| 100 | 20 | A1*B1 |")
	require.Contains(t, got, "```text\nfor each cell:\n  recompute(value)\n```")

	blank := RenderPageMarkdown(StructuredPageOCR{PageNumber: 2, PageType: PageTypeBlank}, nil, DefaultRenderOptions())
	require.Equal(t, "<!-- page:002 -->\n\n[BLANK PAGE]\n", blank)
}
