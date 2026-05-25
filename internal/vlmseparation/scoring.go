package vlmseparation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	jsonsanitize "github.com/go-go-golems/sanitize/pkg/json"
)

var figureCaptionPattern = regexp.MustCompile(`(?i)figure\s+\d+(?:[-.]\d+)?\s*:`)

type BenchmarkResponseParseResult struct {
	Response       *BenchmarkResponse
	Error          error
	Sanitized      bool
	SchemaRepaired bool
	Strategy       string
}

func ParseBenchmarkResponse(raw string) (*BenchmarkResponse, error) {
	result := ParseBenchmarkResponseDetailed(raw)
	return result.Response, result.Error
}

func ParseBenchmarkResponseDetailed(raw string) BenchmarkResponseParseResult {
	jsonText := normalizeJSONText(raw)

	if response, err := parseStrictBenchmarkResponse(jsonText); err == nil && benchmarkResponseHasRequiredShape(response) {
		return BenchmarkResponseParseResult{Response: response, Strategy: "strict-json"}
	}

	sanitizedText := jsonText
	sanitized := false
	sanitizeResult := jsonsanitize.Sanitize(jsonText)
	if sanitizeResult.StrictParseClean && strings.TrimSpace(sanitizeResult.Sanitized) != "" {
		sanitizedText = sanitizeResult.Sanitized
		sanitized = sanitizeResult.Sanitized != jsonText || len(sanitizeResult.Fixes) > 0
		if response, err := parseStrictBenchmarkResponse(sanitizedText); err == nil && benchmarkResponseHasRequiredShape(response) {
			strategy := "strict-json-sanitized"
			if !sanitized {
				strategy = "strict-json-after-sanitize-validation"
			}
			return BenchmarkResponseParseResult{Response: response, Sanitized: sanitized, Strategy: strategy}
		}
	}

	response, err := parseFlexibleBenchmarkResponse(sanitizedText)
	if err == nil && benchmarkResponseHasContent(response) {
		strategy := "schema-repair"
		if sanitized {
			strategy = "sanitize-and-schema-repair"
		}
		return BenchmarkResponseParseResult{Response: response, Sanitized: sanitized, SchemaRepaired: true, Strategy: strategy}
	}

	strictErr := strictDecodeError(jsonText)
	if err != nil {
		return BenchmarkResponseParseResult{Error: fmt.Errorf("parse benchmark response: %w", err), Sanitized: sanitized, Strategy: "failed"}
	}
	return BenchmarkResponseParseResult{Error: fmt.Errorf("parse benchmark response: %w", strictErr), Sanitized: sanitized, Strategy: "failed"}
}

func parseStrictBenchmarkResponse(jsonText string) (*BenchmarkResponse, error) {
	var response BenchmarkResponse
	if err := json.Unmarshal([]byte(jsonText), &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func strictDecodeError(jsonText string) error {
	var response BenchmarkResponse
	if err := json.Unmarshal([]byte(jsonText), &response); err != nil {
		return err
	}
	return fmt.Errorf("response matched JSON syntax but not benchmark response shape")
}

func normalizeJSONText(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```JSON")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)
	if extracted, ok := extractFirstJSONObject(trimmed); ok {
		return extracted
	}
	return trimmed
}

func extractFirstJSONObject(s string) (string, bool) {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return "", false
	}
	inString := false
	escaped := false
	depth := 0
	for i := start; i < len(s); i++ {
		ch := s[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return strings.TrimSpace(s[start : i+1]), true
			}
		}
	}
	return "", false
}

func parseFlexibleBenchmarkResponse(jsonText string) (*BenchmarkResponse, error) {
	var root map[string]any
	decoder := json.NewDecoder(bytes.NewReader([]byte(jsonText)))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		return nil, err
	}
	response := &BenchmarkResponse{}
	response.TargetPage = intFromAny(firstPresent(root, "target_page", "page_number", "page"))
	response.Transcription = stringFromAny(firstPresent(root, "transcription", "text", "ocr", "ocr_text", "markdown", "content"))
	response.TranscribedPageIdentity = pageIdentityFromAny(firstPresent(root, "transcribed_page_identity", "page_identity", "identity"))
	if response.TranscribedPageIdentity.PageNumberVisible == "" {
		response.TranscribedPageIdentity.PageNumberVisible = stringFromAny(firstPresent(root, "page_number_visible", "visible_page_number"))
	}
	response.ContentMarkers = contentMarkersFromAny(firstPresent(root, "content_markers", "markers"))
	if len(response.ContentMarkers.FigureCaptions) == 0 {
		response.ContentMarkers.FigureCaptions = stringsFromAny(firstPresent(root, "figure_captions", "captions"))
	}
	if len(response.ContentMarkers.SectionHeadings) == 0 {
		response.ContentMarkers.SectionHeadings = stringsFromAny(firstPresent(root, "section_headings", "headings"))
	}
	if len(response.ContentMarkers.UniquePhrases) == 0 {
		response.ContentMarkers.UniquePhrases = stringsFromAny(firstPresent(root, "unique_phrases", "phrases"))
	}
	if v, ok := root["suspected_context_copy"]; ok {
		if b, ok := boolFromAny(v); ok {
			response.SuspectedContextCopy = b
		} else if note := stringFromAny(v); note != "" {
			response.Notes = append(response.Notes, note)
		}
	}
	response.Notes = append(response.Notes, stringsFromAny(root["notes"])...)
	if response.Transcription == "" {
		response.Transcription = strings.Join(response.ContentMarkers.UniquePhrases, "\n")
	}
	return response, nil
}

func benchmarkResponseHasRequiredShape(response *BenchmarkResponse) bool {
	return response != nil && response.TargetPage > 0 && benchmarkResponseHasContent(response)
}

func benchmarkResponseHasContent(response *BenchmarkResponse) bool {
	if response == nil {
		return false
	}
	return strings.TrimSpace(response.Transcription) != "" ||
		len(response.ContentMarkers.FigureCaptions) > 0 ||
		len(response.ContentMarkers.SectionHeadings) > 0 ||
		len(response.ContentMarkers.UniquePhrases) > 0 ||
		len(response.TranscribedPageIdentity.TitleOrCaptionLines) > 0
}

func firstPresent(m map[string]any, keys ...string) any {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return nil
}

func pageIdentityFromAny(v any) PageIdentity {
	identity := PageIdentity{}
	switch x := v.(type) {
	case map[string]any:
		identity.PageNumberVisible = stringFromAny(firstPresent(x, "page_number_visible", "visible_page_number"))
		identity.TitleOrCaptionLines = stringsFromAny(firstPresent(x, "title_or_caption_lines", "title_lines", "caption_lines"))
	case string:
		if strings.TrimSpace(x) != "" {
			identity.TitleOrCaptionLines = []string{strings.TrimSpace(x)}
		}
	}
	return identity
}

func contentMarkersFromAny(v any) ContentMarkers {
	markers := ContentMarkers{}
	switch x := v.(type) {
	case map[string]any:
		markers.FigureCaptions = stringsFromAny(firstPresent(x, "figure_captions", "captions"))
		markers.SectionHeadings = stringsFromAny(firstPresent(x, "section_headings", "headings"))
		markers.UniquePhrases = stringsFromAny(firstPresent(x, "unique_phrases", "phrases"))
	case []any:
		for _, s := range stringsFromAny(x) {
			markers.UniquePhrases = append(markers.UniquePhrases, s)
			if figureCaptionPattern.MatchString(s) {
				markers.FigureCaptions = append(markers.FigureCaptions, s)
			}
		}
	case string:
		if strings.TrimSpace(x) != "" {
			markers.UniquePhrases = []string{strings.TrimSpace(x)}
			if figureCaptionPattern.MatchString(x) {
				markers.FigureCaptions = []string{strings.TrimSpace(x)}
			}
		}
	}
	return markers
}

func stringsFromAny(v any) []string {
	switch x := v.(type) {
	case nil:
		return nil
	case []string:
		return cleanStrings(x)
	case []any:
		out := make([]string, 0, len(x))
		for _, item := range x {
			if s := stringFromAny(item); s != "" {
				out = append(out, s)
			}
		}
		return out
	case string:
		if strings.TrimSpace(x) == "" {
			return nil
		}
		return []string{strings.TrimSpace(x)}
	default:
		if s := stringFromAny(v); s != "" {
			return []string{s}
		}
		return nil
	}
}

func cleanStrings(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if strings.TrimSpace(s) != "" {
			out = append(out, strings.TrimSpace(s))
		}
	}
	return out
}

func stringFromAny(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(x)
	case json.Number:
		return x.String()
	case float64:
		if math.Trunc(x) == x {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(x)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func intFromAny(v any) int {
	switch x := v.(type) {
	case nil:
		return 0
	case int:
		return x
	case json.Number:
		i, _ := strconv.Atoi(x.String())
		return i
	case float64:
		return int(x)
	case string:
		i, _ := strconv.Atoi(strings.TrimLeft(strings.TrimSpace(x), "0"))
		if i == 0 && strings.Trim(strings.TrimSpace(x), "0") == "" && strings.TrimSpace(x) != "" {
			return 0
		}
		return i
	default:
		return 0
	}
}

func boolFromAny(v any) (bool, bool) {
	switch x := v.(type) {
	case bool:
		return x, true
	case string:
		s := strings.TrimSpace(strings.ToLower(x))
		if s == "true" || s == "yes" || s == "1" {
			return true, true
		}
		if s == "false" || s == "no" || s == "0" {
			return false, true
		}
	}
	return false, false
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
	normalizedText := normalizePhraseText(text)
	for _, phrase := range phrases {
		phrase = normalizePhraseText(phrase)
		if phrase == "" {
			continue
		}
		if strings.Contains(normalizedText, phrase) {
			count++
		}
	}
	return count
}

func normalizePhraseText(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
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
