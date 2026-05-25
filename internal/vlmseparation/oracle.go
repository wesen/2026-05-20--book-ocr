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
		12:  {TargetPage: 12, ExpectedPhrases: []string{"A Rudimentary User Interface", "Representation Shift"}, ForbiddenCaptions: []string{"Figure 1-1: A Rudimentary User Interface"}},
		13:  {TargetPage: 13, ExpectedCaptions: []string{"Figure 1-1: A Rudimentary User Interface"}, ExpectedPhrases: []string{"Application Data Base", "observables", "queries"}},
		31:  {TargetPage: 31, ExpectedCaptions: []string{"Figure 2-1: The Primitive Presentation System (PPS) Model"}, ExpectedPhrases: []string{"Presentation Data Base", "Presentation Editor", "Application Data Base"}, ForbiddenCaptions: []string{"Figure 2-2: PPSCalc -- Formula Display", "Figure 2-3: PPSCalc -- Value Display"}},
		32:  {TargetPage: 32, ExpectedCaptions: []string{"Figure 2-2: PPSCalc -- Formula Display", "Figure 2-3: PPSCalc -- Value Display"}, ExpectedPhrases: []string{"A1*B1", "C1+C2", "2375"}, ForbiddenCaptions: []string{"Figure 2-1: The Primitive Presentation System (PPS) Model"}},
		42:  {TargetPage: 42, ExpectedCaptions: []string{"Figure 2-9: Presenter Parts"}, ExpectedPhrases: []string{"Domain Collector", "Recognizer", "Presentation Editor"}},
		43:  {TargetPage: 43, ExpectedPhrases: []string{"domain collector", "semantic presenter", "application data base"}, ForbiddenCaptions: []string{"Figure 2-9: Presenter Parts"}},
		59:  {TargetPage: 59, ExpectedCaptions: []string{"Figure 3-2: Extension with Both Planning and Immediate Changes"}, ExpectedPhrases: []string{"Presentation Data Base", "Application Data Base", "Future Data Base"}, ForbiddenCaptions: []string{"Figure 3-3: Command Data Base Extension"}},
		60:  {TargetPage: 60, ExpectedPhrases: []string{"Adding a Data Base of Commands", "direct manipulation", "command"}, ForbiddenCaptions: []string{"Figure 3-2: Extension with Both Planning and Immediate Changes", "Figure 3-3: Command Data Base Extension"}},
		87:  {TargetPage: 87, ExpectedPhrases: []string{"property sheet", "Chapter 7", "delete command"}, ForbiddenCaptions: []string{"Figure 4-6: Xerox Star -- Property Sheet"}},
		88:  {TargetPage: 88, ExpectedCaptions: []string{"Figure 4-6: Xerox Star -- Property Sheet"}, ExpectedPhrases: []string{"DOCUMENT PROPERTIES", "Chapter 7", "COVER SHEET"}, ForbiddenCaptions: []string{"Figure 4-7: Xerox Star -- Delete Confirmation", "Figure 4-8: Xerox Star Model"}},
		97:  {TargetPage: 97, ExpectedCaptions: []string{"Figure 4-12: Sample of Steamer Icons"}, ExpectedPhrases: []string{"ICON SAMPLER", "centrifugal pump", "digital bar"}},
		98:  {TargetPage: 98, ExpectedPhrases: []string{"animation", "simulation", "pipe flow"}, ForbiddenCaptions: []string{"Figure 4-12: Sample of Steamer Icons"}},
		112: {TargetPage: 112, ExpectedCaptions: []string{"Figure 5-6: Command Description Support"}, ExpectedPhrases: []string{"Domain Object", "Command Application", "Parameter Type"}},
		113: {TargetPage: 113, ExpectedPhrases: []string{"Lisp functions", "command application", "parameter descriptions"}, ForbiddenCaptions: []string{"Figure 5-6: Command Description Support"}},
		115: {TargetPage: 115, ExpectedCaptions: []string{"Figure 5-7: Reference Resolution"}, ExpectedPhrases: []string{"Resolve to property's value", "Resolve to property's owner", "presented domain object"}},
		116: {TargetPage: 116, ExpectedPhrases: []string{"Graphics Redisplay", "PSBase", "presentation data base"}, ForbiddenCaptions: []string{"Figure 5-7: Reference Resolution"}},
	}
	if o, ok := oracles[page]; ok {
		return o
	}
	return PageOracle{TargetPage: page, ExpectedPhrases: []string{fmt.Sprintf("page %03d", page)}}
}
