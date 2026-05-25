package ocrvalidation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvaluateAnchorsFlagsForbiddenCaptionBleed(t *testing.T) {
	oracle := AnchorOracle{TargetPage: 12, ExpectedPhrases: []string{"A Rudimentary User Interface"}, ForbiddenCaptions: []string{"Figure 1-1: A Rudimentary User Interface"}}
	result := EvaluateAnchors("A Rudimentary User Interface\nFigure 1-1: A Rudimentary User Interface", oracle)
	require.Equal(t, 1, result.ExpectedHits)
	require.Equal(t, 1, result.ForbiddenHits)
	require.Equal(t, 1, result.ForbiddenCaptionHits)
	require.True(t, result.SuspectedBleed)
}

func TestEvaluateAnchorsNormalizesLineBreaks(t *testing.T) {
	oracle := AnchorOracle{TargetPage: 59, ExpectedPhrases: []string{"Application Data Base"}}
	result := EvaluateAnchors("Application\nData\nBase", oracle)
	require.Equal(t, 1, result.ExpectedHits)
	require.Empty(t, result.MissingExpected)
}

func TestDetectAdjacentDuplicateFigureCaptions(t *testing.T) {
	warnings := DetectAdjacentDuplicateFigureCaptions([]PageText{
		{Number: 12, Markdown: "<!-- page:012 -->\n\nA prose reference.\n\nFigure 1-1: A Rudimentary User Interface\n"},
		{Number: 13, Markdown: "<!-- page:013 -->\n\nFigure 1-1: A Rudimentary User Interface\n\n![figure](figures/page_013_figure_01.png)\n"},
	})
	require.Len(t, warnings, 1)
	require.Equal(t, "adjacent_duplicate_figure_caption", warnings[0].Code)
	require.Equal(t, 12, warnings[0].Page)
	require.Equal(t, 13, warnings[0].Other)
}

func TestFigureReferenceWithoutCaptionDoesNotTriggerDuplicate(t *testing.T) {
	warnings := DetectAdjacentDuplicateFigureCaptions([]PageText{
		{Number: 12, Markdown: "A prose sentence mentions Figure 1-1 but does not render a caption."},
		{Number: 13, Markdown: "Figure 1-1: A Rudimentary User Interface\n\n![figure](figures/page_013_figure_01.png)"},
	})
	require.Empty(t, warnings)
}

func TestPage116MayMentionFigure57ButMustNotRenderCaption(t *testing.T) {
	oracle := AnchorOracle{TargetPage: 116, ExpectedPhrases: []string{"Graphics Redisplay", "PSBase", "presentation data base"}, ForbiddenCaptions: []string{"Figure 5-7: Reference Resolution"}}
	text := "The third kind of resolution walks up the presentation hierarchy.\n\n5.2 Graphics Redisplay\n\nThis section discusses the next PSBase module and the presentation data base."
	result := EvaluateAnchors(text, oracle)
	require.Equal(t, 3, result.ExpectedHits)
	require.False(t, result.SuspectedBleed)
}
