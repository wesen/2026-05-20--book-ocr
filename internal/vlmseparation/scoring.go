package vlmseparation

import (
	"encoding/json"
	"math"
	"strings"
)

func ParseBenchmarkResponse(raw string) (*BenchmarkResponse, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)
	var response BenchmarkResponse
	if err := json.Unmarshal([]byte(trimmed), &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func ScoreTrial(raw string, parsed *BenchmarkResponse, oracle PageOracle, parseErr error) TrialMetrics {
	metrics := TrialMetrics{
		JSONParseOK:         parseErr == nil,
		ExpectedPhraseTotal: len(oracle.ExpectedPhrases) + len(oracle.ExpectedCaptions),
	}
	text := strings.ToLower(raw)
	if parsed != nil {
		parts := []string{parsed.Transcription}
		parts = append(parts, parsed.ContentMarkers.FigureCaptions...)
		parts = append(parts, parsed.ContentMarkers.SectionHeadings...)
		parts = append(parts, parsed.ContentMarkers.UniquePhrases...)
		parts = append(parts, parsed.TranscribedPageIdentity.TitleOrCaptionLines...)
		text = strings.ToLower(strings.Join(parts, "\n"))
		metrics.SuspectedBleed = parsed.SuspectedContextCopy
	}
	expected := append(append([]string{}, oracle.ExpectedPhrases...), oracle.ExpectedCaptions...)
	forbidden := append(append([]string{}, oracle.ForbiddenPhrases...), oracle.ForbiddenCaptions...)
	metrics.ExpectedPhraseHits = countPhraseHits(text, expected)
	metrics.ForbiddenPhraseHits = countPhraseHits(text, forbidden)
	metrics.ForbiddenCaptionHits = countPhraseHits(text, oracle.ForbiddenCaptions)
	if metrics.ForbiddenPhraseHits > 0 || metrics.ForbiddenCaptionHits > 0 {
		metrics.SuspectedBleed = true
	}
	denom := math.Max(1, float64(metrics.ExpectedPhraseTotal))
	score := float64(metrics.ExpectedPhraseHits) / denom
	if metrics.ForbiddenPhraseHits > 0 {
		score -= 0.5 * float64(metrics.ForbiddenPhraseHits)
	}
	if !metrics.JSONParseOK {
		score -= 0.25
	}
	if score < 0 {
		score = 0
	}
	metrics.TargetOnlyScore = score
	return metrics
}

func countPhraseHits(text string, phrases []string) int {
	count := 0
	for _, phrase := range phrases {
		phrase = strings.TrimSpace(strings.ToLower(phrase))
		if phrase == "" {
			continue
		}
		if strings.Contains(text, phrase) {
			count++
		}
	}
	return count
}

func Summarize(runID string, trials []TrialResult) Summary {
	summary := Summary{RunID: runID, TrialCount: len(trials), ByScenario: map[string]Summary{}}
	var scoreTotal float64
	for _, trial := range trials {
		if trial.Status == "succeeded" || trial.Status == "parse_failed" {
			summary.Succeeded++
		} else {
			summary.Failed++
		}
		if trial.Metrics.SuspectedBleed {
			summary.SuspectedBleed++
		}
		scoreTotal += trial.Metrics.TargetOnlyScore
		by := summary.ByScenario[trial.Scenario]
		by.RunID = runID
		by.TrialCount++
		if trial.Status == "succeeded" || trial.Status == "parse_failed" {
			by.Succeeded++
		} else {
			by.Failed++
		}
		if trial.Metrics.SuspectedBleed {
			by.SuspectedBleed++
		}
		by.AverageTargetOnlyScore += trial.Metrics.TargetOnlyScore
		summary.ByScenario[trial.Scenario] = by
	}
	if len(trials) > 0 {
		summary.AverageTargetOnlyScore = scoreTotal / float64(len(trials))
	}
	for key, by := range summary.ByScenario {
		if by.TrialCount > 0 {
			by.AverageTargetOnlyScore = by.AverageTargetOnlyScore / float64(by.TrialCount)
		}
		summary.ByScenario[key] = by
	}
	return summary
}
