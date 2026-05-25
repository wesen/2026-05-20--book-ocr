package vlmseparation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePages(t *testing.T) {
	pages, err := ParsePages("13,12,31-32", "")
	require.NoError(t, err)
	require.Equal(t, []int{12, 13, 31, 32}, pages)
}

func TestScoreTrialFlagsForbiddenCaptionBleed(t *testing.T) {
	raw := `{"target_page":12,"transcribed_page_identity":{"title_or_caption_lines":["Figure 1-1: A Rudimentary User Interface"]},"content_markers":{"figure_captions":["Figure 1-1: A Rudimentary User Interface"]},"transcription":"A Rudimentary User Interface\nFigure 1-1: A Rudimentary User Interface","suspected_context_copy":false}`
	parsed, err := ParseBenchmarkResponse(raw)
	require.NoError(t, err)
	metrics := ScoreTrial(raw, parsed, PageOracle{TargetPage: 12, ExpectedPhrases: []string{"A Rudimentary User Interface"}, ForbiddenCaptions: []string{"Figure 1-1: A Rudimentary User Interface"}}, nil)
	require.True(t, metrics.JSONParseOK)
	require.True(t, metrics.SuspectedBleed)
	require.Equal(t, 1, metrics.ForbiddenCaptionHits)
}
