package vlmseparation

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func AllScenarios() []Scenario {
	return []Scenario{
		{Name: ScenarioTargetOnly, Description: "Baseline: target page image only."},
		{Name: ScenarioSingleBlockTargetFirst, Description: "Current-style layout: one multimodal block with target image first, then context images."},
		{Name: ScenarioSingleBlockLabeledImages, Description: "One multimodal block with richer labels/metadata for target and context images."},
		{Name: ScenarioMultiBlockLabeled, Description: "Multiple user blocks: target image separated from context images by explicit text blocks."},
		{Name: ScenarioContextFirstNegativeControl, Description: "Negative control: context images before target image."},
		{Name: ScenarioTargetPlusTextContext, Description: "Target image only plus neighboring page text summaries."},
	}
}

func NormalizeScenarios(values []string) ([]Scenario, error) {
	known := map[string]Scenario{}
	for _, s := range AllScenarios() {
		known[s.Name] = s
	}
	var names []string
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				names = append(names, part)
			}
		}
	}
	if len(names) == 0 {
		names = []string{ScenarioTargetOnly}
	}
	out := make([]Scenario, 0, len(names))
	for _, name := range names {
		s, ok := known[name]
		if !ok {
			return nil, fmt.Errorf("unknown scenario %q", name)
		}
		out = append(out, s)
	}
	return out, nil
}

func ParsePages(value string, preset string) ([]int, error) {
	if strings.TrimSpace(value) == "" {
		switch strings.TrimSpace(preset) {
		case "", "report794-bleed-smoke":
			return []int{12, 13, 31, 32, 59, 60}, nil
		case "report794-figure-adjacent":
			return []int{12, 13, 31, 32, 42, 43, 59, 60, 87, 88, 97, 98, 112, 113, 115, 116}, nil
		default:
			return nil, fmt.Errorf("unknown preset %q and --target-pages is empty", preset)
		}
	}
	seen := map[int]bool{}
	var pages []int
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			start, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if err != nil {
				return nil, err
			}
			end, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err != nil {
				return nil, err
			}
			if end < start {
				return nil, fmt.Errorf("invalid page range %q", part)
			}
			for p := start; p <= end; p++ {
				if !seen[p] {
					seen[p] = true
					pages = append(pages, p)
				}
			}
			continue
		}
		p, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		if !seen[p] {
			seen[p] = true
			pages = append(pages, p)
		}
	}
	sort.Ints(pages)
	return pages, nil
}

func OracleForPage(page int) PageOracle {
	oracles := map[int]PageOracle{
		12: {TargetPage: 12, ExpectedPhrases: []string{"A Rudimentary User Interface", "Representation Shift"}, ForbiddenCaptions: []string{"Figure 1-1: A Rudimentary User Interface"}},
		13: {TargetPage: 13, ExpectedCaptions: []string{"Figure 1-1: A Rudimentary User Interface"}, ExpectedPhrases: []string{"Application Data Base", "observables", "queries"}},
		31: {TargetPage: 31, ExpectedPhrases: []string{"PPSCalc", "formula display", "value display"}},
		32: {TargetPage: 32, ExpectedCaptions: []string{"Figure 2-2: PPSCalc -- Formula Display", "Figure 2-3: PPSCalc -- Value Display"}},
		59: {TargetPage: 59, ExpectedPhrases: []string{"planned changes", "immediate changes"}},
		60: {TargetPage: 60, ExpectedPhrases: []string{"Adding a Data Base of Commands", "command"}},
	}
	if o, ok := oracles[page]; ok {
		return o
	}
	return PageOracle{TargetPage: page, ExpectedPhrases: []string{fmt.Sprintf("page %03d", page)}}
}
