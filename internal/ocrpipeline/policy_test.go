package ocrpipeline

import (
	"testing"

	"github.com/go-go-golems/book-ocr/internal/bookprofile"
	"github.com/stretchr/testify/require"
)

func mixedFixturePage() StructuredPageOCR {
	return StructuredPageOCR{SchemaVersion: StructuredOCRSchemaVersion, BookID: "report-794", PageNumber: 48, PageType: PageTypeTable, Blocks: []OCRBlock{
		{ID: "p048-h001", Type: BlockHeading, Level: 2, Text: "2.1 PPSCalc"},
		{ID: "p048-c001", Type: BlockCode, Text: "(defmethod present ((self box)))"},
		{ID: "p048-f001", Type: BlockFigure, Caption: "Figure 2-12: PPSCalc -- Formula Moved", Description: "Spreadsheet grid with columns A B C."},
		{ID: "p048-t001", Type: BlockTable, Table: &TableBlock{Headers: []string{"", "A"}, Rows: [][]string{{"1", "100"}}}},
	}}
}

// The compiled Report-794 profile must reproduce the built-in defaults
// exactly: same prompt bytes, same rendered Markdown bytes. This is the
// contract that lets profiles replace hardcoded policy without a quality
// regression on the book everything was tuned on.
func TestReport794ProfilePolicyMatchesBuiltinDefaults(t *testing.T) {
	prompt, render := PolicyFromProfile(bookprofile.Report794())

	input := StructuredOCRInput{BookID: "report-794", PageNumber: 32}
	require.Equal(t, RenderStructuredOCRPromptSpec(DefaultPromptSpec(), input), RenderStructuredOCRPromptSpec(*prompt, input))

	page := mixedFixturePage()
	figures := FigureMap{48: {"p048-f001": {Path: "figures/page_048_figure_01.png"}}}
	require.Equal(t, RenderPageMarkdown(page, figures, DefaultRenderOptions()), RenderPageMarkdown(page, figures, *render))
}

// A generic profile produces generic behavior: plain code fences, no
// Lisp/PPSCalc lexicon in the prompt, and no figure suppression — the F2/F3
// hardcodings must not leak into profile-driven runs for other books.
func TestGenericProfilePolicyHasNoReport794Behavior(t *testing.T) {
	generic := bookprofile.Profile{ID: "generic-technical-book", PageImages: bookprofile.PageImagePolicy{Glob: "page_*.png"}}
	prompt, render := PolicyFromProfile(generic)

	input := StructuredOCRInput{BookID: "some-other-book", PageNumber: 7}
	promptText := RenderStructuredOCRPromptSpec(*prompt, input)
	require.NotContains(t, promptText, "PPSCalc")
	require.NotContains(t, promptText, "Common Lisp")
	require.NotContains(t, promptText, "Preserve historical terminology")

	page := mixedFixturePage()
	figures := FigureMap{48: {"p048-f001": {Path: "figures/page_048_figure_01.png"}}}
	rendered := RenderPageMarkdown(page, figures, *render)
	require.Contains(t, rendered, "```\n(defmethod present ((self box)))\n```", "generic profile renders plain fences")
	require.NotContains(t, rendered, "```common-lisp")
	require.Contains(t, rendered, "![", "generic profile does not suppress the spreadsheet figure image")
}
