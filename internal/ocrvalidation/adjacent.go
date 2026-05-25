package ocrvalidation

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var figureCaptionRE = regexp.MustCompile(`(?im)^\s*(Figure\s+\d+(?:[-.]\d+)?\s*:\s*.+?)\s*$`)

func ExtractFigureCaptions(markdown string) []string {
	matches := figureCaptionRE.FindAllStringSubmatch(markdown, -1)
	out := make([]string, 0, len(matches))
	seen := map[string]bool{}
	for _, match := range matches {
		caption := strings.TrimSpace(match[1])
		key := NormalizeCaption(caption)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, caption)
	}
	return out
}

func DetectAdjacentDuplicateFigureCaptions(pages []PageText) []Warning {
	captionPages := map[string][]int{}
	captionDisplay := map[string]string{}
	for _, page := range pages {
		for _, caption := range ExtractFigureCaptions(page.Markdown) {
			key := NormalizeCaption(caption)
			captionPages[key] = append(captionPages[key], page.Number)
			if captionDisplay[key] == "" {
				captionDisplay[key] = caption
			}
		}
	}
	var warnings []Warning
	for key, nums := range captionPages {
		sort.Ints(nums)
		for i := 1; i < len(nums); i++ {
			if nums[i] == nums[i-1]+1 {
				warnings = append(warnings, Warning{Code: "adjacent_duplicate_figure_caption", Page: nums[i-1], Other: nums[i], Value: captionDisplay[key], Message: fmt.Sprintf("%q appears on adjacent pages %03d and %03d", captionDisplay[key], nums[i-1], nums[i])})
			}
		}
	}
	sort.Slice(warnings, func(i, j int) bool {
		if warnings[i].Page != warnings[j].Page {
			return warnings[i].Page < warnings[j].Page
		}
		return warnings[i].Value < warnings[j].Value
	})
	return warnings
}

func NormalizeCaption(caption string) string {
	return strings.Join(strings.Fields(strings.ToLower(caption)), " ")
}
