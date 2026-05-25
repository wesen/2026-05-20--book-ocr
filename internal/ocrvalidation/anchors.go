package ocrvalidation

import "strings"

func EvaluateAnchors(text string, oracle AnchorOracle) AnchorResult {
	result := AnchorResult{TargetPage: oracle.TargetPage}
	expected := append(append([]string{}, oracle.ExpectedPhrases...), oracle.ExpectedCaptions...)
	forbidden := append(append([]string{}, oracle.ForbiddenPhrases...), oracle.ForbiddenCaptions...)
	result.ExpectedTotal = len(expected)
	normalizedText := normalizeText(text)
	for _, phrase := range expected {
		if containsPhrase(normalizedText, phrase) {
			result.ExpectedHits++
		} else if strings.TrimSpace(phrase) != "" {
			result.MissingExpected = append(result.MissingExpected, phrase)
		}
	}
	for _, phrase := range forbidden {
		if containsPhrase(normalizedText, phrase) {
			result.ForbiddenHits++
			result.MatchedForbidden = append(result.MatchedForbidden, phrase)
		}
	}
	for _, phrase := range oracle.ForbiddenCaptions {
		if containsPhrase(normalizedText, phrase) {
			result.ForbiddenCaptionHits++
		}
	}
	result.SuspectedBleed = result.ForbiddenHits > 0 || result.ForbiddenCaptionHits > 0
	return result
}

func containsPhrase(normalizedText, phrase string) bool {
	phrase = normalizeText(phrase)
	return phrase != "" && strings.Contains(normalizedText, phrase)
}

func normalizeText(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}
