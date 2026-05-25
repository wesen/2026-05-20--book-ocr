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

func TestFigureAdjacentPresetHasSpecializedOracles(t *testing.T) {
	pages, err := ParsePages("", "report794-figure-adjacent")
	require.NoError(t, err)
	require.Equal(t, []int{12, 13, 31, 32, 42, 43, 59, 60, 87, 88, 97, 98, 112, 113, 115, 116}, pages)
	for _, page := range pages {
		oracle := OracleForPage(page)
		require.NotContains(t, oracle.ExpectedPhrases, "page "+formatPage(page), "page %d should not use fallback oracle", page)
		require.NotEmpty(t, append(append([]string{}, oracle.ExpectedPhrases...), oracle.ExpectedCaptions...), "page %d should have expected phrases or captions", page)
	}
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

func TestParseBenchmarkResponseRepairsLiveStringAndArrayVariants(t *testing.T) {
	raw := `{
		"target_page": "012",
		"transcribed_page_identity": "Page 10 (scanned page)",
		"content_markers": ["A Rudimentary User Interface.", "Representation Shift."],
		"transcription": "A Rudimentary User Interface. Representation Shift.",
		"suspected_context_copy": "model note, not a boolean"
	}`
	result := ParseBenchmarkResponseDetailed(raw)
	require.NoError(t, result.Error)
	require.True(t, result.SchemaRepaired)
	require.Equal(t, "schema-repair", result.Strategy)
	require.Equal(t, 12, result.Response.TargetPage)
	require.Contains(t, result.Response.TranscribedPageIdentity.TitleOrCaptionLines, "Page 10 (scanned page)")
	require.Contains(t, result.Response.ContentMarkers.UniquePhrases, "A Rudimentary User Interface.")
	require.False(t, result.Response.SuspectedContextCopy)
	require.NotEmpty(t, result.Response.Notes)
}

func TestParseBenchmarkResponseRepairsTextFieldVariant(t *testing.T) {
	raw := `{"page_number":13,"text":"Figure 1-1: A Rudimentary User Interface\nApplication Data Base"}`
	result := ParseBenchmarkResponseDetailed(raw)
	require.NoError(t, result.Error)
	require.True(t, result.SchemaRepaired)
	require.Equal(t, 13, result.Response.TargetPage)
	require.Contains(t, result.Response.Transcription, "Figure 1-1")
	metrics := ScoreTrial(raw, result.Response, PageOracle{TargetPage: 13, ExpectedCaptions: []string{"Figure 1-1: A Rudimentary User Interface"}}, result.Error)
	require.True(t, metrics.JSONParseOK)
	require.Equal(t, 1, metrics.ExpectedPhraseHits)
}

func TestParseBenchmarkResponseExtractsFencedJSONObject(t *testing.T) {
	raw := "Here is the requested JSON:\n```json\n{\"target_page\":12,\"transcription\":\"A Rudimentary User Interface\"}\n```\nDone."
	result := ParseBenchmarkResponseDetailed(raw)
	require.NoError(t, result.Error)
	require.Equal(t, "strict-json", result.Strategy)
	require.Equal(t, 12, result.Response.TargetPage)
}
