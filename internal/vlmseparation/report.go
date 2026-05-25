package vlmseparation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type ReportConfig struct {
	OutDirs      []string
	ReportDir    string
	MarkdownPath string
	JSONPath     string
	Write        bool
}

type ReportResult struct {
	GeneratedAt       time.Time                 `json:"generated_at"`
	OutDirs           []string                  `json:"out_dirs"`
	TrialCount        int                       `json:"trial_count"`
	LogicalTrialCount int                       `json:"logical_trial_count"`
	DuplicateCount    int                       `json:"duplicate_count"`
	ReplacementCount  int                       `json:"replacement_count"`
	Succeeded         int                       `json:"succeeded"`
	ParseOK           int                       `json:"parse_ok"`
	SuspectedBleed    int                       `json:"suspected_bleed"`
	ForbiddenHits     int                       `json:"forbidden_hits"`
	AverageScore      float64                   `json:"average_score"`
	ByScenario        []ReportScenarioAggregate `json:"by_scenario"`
	ByPage            []ReportPageAggregate     `json:"by_page"`
	LowTrials         []ReportTrial             `json:"low_trials,omitempty"`
	Trials            []ReportTrial             `json:"trials"`
}

type ReportTrial struct {
	TrialResult
	SourceOutDir string `json:"source_out_dir"`
}

type ReportScenarioAggregate struct {
	Scenario       string  `json:"scenario"`
	Trials         int     `json:"trials"`
	Succeeded      int     `json:"succeeded"`
	ParseOK        int     `json:"parse_ok"`
	SchemaRepaired int     `json:"schema_repaired"`
	SuspectedBleed int     `json:"suspected_bleed"`
	ForbiddenHits  int     `json:"forbidden_hits"`
	AverageScore   float64 `json:"average_score"`
	MinimumScore   float64 `json:"minimum_score"`
}

type ReportPageAggregate struct {
	TargetPage     int     `json:"target_page"`
	Trials         int     `json:"trials"`
	Succeeded      int     `json:"succeeded"`
	ParseOK        int     `json:"parse_ok"`
	SuspectedBleed int     `json:"suspected_bleed"`
	ForbiddenHits  int     `json:"forbidden_hits"`
	AverageScore   float64 `json:"average_score"`
	MinimumScore   float64 `json:"minimum_score"`
}

func BuildReport(ctx context.Context, cfg ReportConfig) (*ReportResult, error) {
	_ = ctx
	outDirs := normalizeList(cfg.OutDirs)
	if len(outDirs) == 0 {
		return nil, fmt.Errorf("at least one --out-dir is required")
	}
	var all []ReportTrial
	for _, outDir := range outDirs {
		trials, err := loadReportTrials(outDir)
		if err != nil {
			return nil, err
		}
		all = append(all, trials...)
	}
	logical, duplicates, replacements := selectLogicalTrials(all)
	result := summarizeReport(outDirs, all, logical, duplicates, replacements)
	if cfg.Write {
		cfg.OutDirs = outDirs
		if err := writeReportArtifacts(cfg, result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func loadReportTrials(outDir string) ([]ReportTrial, error) {
	trialPaths, err := filepath.Glob(filepath.Join(outDir, "trials", "trial-*", "trial.json"))
	if err != nil {
		return nil, err
	}
	if len(trialPaths) == 0 {
		return nil, fmt.Errorf("no trial.json files found under %s", filepath.Join(outDir, "trials"))
	}
	sort.Strings(trialPaths)
	trials := make([]ReportTrial, 0, len(trialPaths))
	for _, path := range trialPaths {
		trial, err := readTrialResult(path)
		if err != nil {
			return nil, err
		}
		rescored, err := rescoreTrialResult(trial)
		if err != nil {
			return nil, fmt.Errorf("report rescore %s: %w", path, err)
		}
		trials = append(trials, ReportTrial{TrialResult: rescored, SourceOutDir: outDir})
	}
	return trials, nil
}

func selectLogicalTrials(all []ReportTrial) ([]ReportTrial, int, int) {
	byKey := map[string]ReportTrial{}
	var keys []string
	duplicates := 0
	replacements := 0
	for _, trial := range all {
		key := reportTrialKey(trial)
		current, exists := byKey[key]
		if !exists {
			byKey[key] = trial
			keys = append(keys, key)
			continue
		}
		duplicates++
		if reportTrialRank(trial) > reportTrialRank(current) {
			byKey[key] = trial
			replacements++
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := byKey[keys[i]], byKey[keys[j]]
		if a.TargetPage != b.TargetPage {
			return a.TargetPage < b.TargetPage
		}
		return a.Scenario < b.Scenario
	})
	out := make([]ReportTrial, 0, len(keys))
	for _, key := range keys {
		out = append(out, byKey[key])
	}
	return out, duplicates, replacements
}

func reportTrialKey(trial ReportTrial) string {
	return fmt.Sprintf("%03d/%s", trial.TargetPage, trial.Scenario)
}

func reportTrialRank(trial ReportTrial) float64 {
	rank := trial.Metrics.TargetOnlyScore
	if trial.Status == "succeeded" {
		rank += 1000
	}
	if trial.Metrics.JSONParseOK {
		rank += 100
	}
	if !trial.Metrics.SuspectedBleed {
		rank += 10
	}
	if trial.Error == "" {
		rank += 1
	}
	return rank
}

func summarizeReport(outDirs []string, all []ReportTrial, logical []ReportTrial, duplicates, replacements int) *ReportResult {
	result := &ReportResult{GeneratedAt: time.Now().UTC(), OutDirs: outDirs, TrialCount: len(all), LogicalTrialCount: len(logical), DuplicateCount: duplicates, ReplacementCount: replacements, Trials: logical}
	var scoreTotal float64
	byScenario := map[string][]ReportTrial{}
	byPage := map[int][]ReportTrial{}
	for _, trial := range logical {
		if trial.Status == "succeeded" {
			result.Succeeded++
		}
		if trial.Metrics.JSONParseOK {
			result.ParseOK++
		}
		if trial.Metrics.SuspectedBleed {
			result.SuspectedBleed++
		}
		forbidden := trial.Metrics.ForbiddenPhraseHits + trial.Metrics.ForbiddenCaptionHits
		result.ForbiddenHits += forbidden
		scoreTotal += trial.Metrics.TargetOnlyScore
		byScenario[trial.Scenario] = append(byScenario[trial.Scenario], trial)
		byPage[trial.TargetPage] = append(byPage[trial.TargetPage], trial)
		if trial.Metrics.TargetOnlyScore < 1 || forbidden > 0 || trial.Metrics.SuspectedBleed || trial.Status != "succeeded" {
			result.LowTrials = append(result.LowTrials, trial)
		}
	}
	if len(logical) > 0 {
		result.AverageScore = scoreTotal / float64(len(logical))
	}
	for scenario, trials := range byScenario {
		result.ByScenario = append(result.ByScenario, scenarioAggregate(scenario, trials))
	}
	sort.Slice(result.ByScenario, func(i, j int) bool { return result.ByScenario[i].Scenario < result.ByScenario[j].Scenario })
	for page, trials := range byPage {
		result.ByPage = append(result.ByPage, pageAggregate(page, trials))
	}
	sort.Slice(result.ByPage, func(i, j int) bool { return result.ByPage[i].TargetPage < result.ByPage[j].TargetPage })
	sort.Slice(result.LowTrials, func(i, j int) bool {
		if result.LowTrials[i].TargetPage != result.LowTrials[j].TargetPage {
			return result.LowTrials[i].TargetPage < result.LowTrials[j].TargetPage
		}
		return result.LowTrials[i].Scenario < result.LowTrials[j].Scenario
	})
	return result
}

func scenarioAggregate(scenario string, trials []ReportTrial) ReportScenarioAggregate {
	agg := ReportScenarioAggregate{Scenario: scenario, Trials: len(trials), MinimumScore: 1}
	var scoreTotal float64
	if len(trials) == 0 {
		agg.MinimumScore = 0
		return agg
	}
	for _, trial := range trials {
		if trial.Status == "succeeded" {
			agg.Succeeded++
		}
		if trial.Metrics.JSONParseOK {
			agg.ParseOK++
		}
		if trial.Metrics.SchemaRepaired {
			agg.SchemaRepaired++
		}
		if trial.Metrics.SuspectedBleed {
			agg.SuspectedBleed++
		}
		agg.ForbiddenHits += trial.Metrics.ForbiddenPhraseHits + trial.Metrics.ForbiddenCaptionHits
		scoreTotal += trial.Metrics.TargetOnlyScore
		if trial.Metrics.TargetOnlyScore < agg.MinimumScore {
			agg.MinimumScore = trial.Metrics.TargetOnlyScore
		}
	}
	agg.AverageScore = scoreTotal / float64(len(trials))
	return agg
}

func pageAggregate(page int, trials []ReportTrial) ReportPageAggregate {
	agg := ReportPageAggregate{TargetPage: page, Trials: len(trials), MinimumScore: 1}
	var scoreTotal float64
	if len(trials) == 0 {
		agg.MinimumScore = 0
		return agg
	}
	for _, trial := range trials {
		if trial.Status == "succeeded" {
			agg.Succeeded++
		}
		if trial.Metrics.JSONParseOK {
			agg.ParseOK++
		}
		if trial.Metrics.SuspectedBleed {
			agg.SuspectedBleed++
		}
		agg.ForbiddenHits += trial.Metrics.ForbiddenPhraseHits + trial.Metrics.ForbiddenCaptionHits
		scoreTotal += trial.Metrics.TargetOnlyScore
		if trial.Metrics.TargetOnlyScore < agg.MinimumScore {
			agg.MinimumScore = trial.Metrics.TargetOnlyScore
		}
	}
	agg.AverageScore = scoreTotal / float64(len(trials))
	return agg
}

func writeReportArtifacts(cfg ReportConfig, result *ReportResult) error {
	reportDir := strings.TrimSpace(cfg.ReportDir)
	if reportDir == "" {
		reportDir = cfg.OutDirs[0]
	}
	jsonPath := strings.TrimSpace(cfg.JSONPath)
	if jsonPath == "" {
		jsonPath = filepath.Join(reportDir, "report.json")
	}
	markdownPath := strings.TrimSpace(cfg.MarkdownPath)
	if markdownPath == "" {
		markdownPath = filepath.Join(reportDir, "report.md")
	}
	if err := writeJSON(jsonPath, result); err != nil {
		return err
	}
	return os.WriteFile(markdownPath, []byte(RenderReportMarkdown(result)), 0o644)
}

func RenderReportMarkdown(result *ReportResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# VLM Separation Benchmark Report\n\n")
	fmt.Fprintf(&b, "- Generated at: %s\n", result.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "- Input runs: %s\n", strings.Join(result.OutDirs, ", "))
	fmt.Fprintf(&b, "- Raw trials: %d\n", result.TrialCount)
	fmt.Fprintf(&b, "- Logical trials: %d\n", result.LogicalTrialCount)
	fmt.Fprintf(&b, "- Duplicate logical cells: %d\n", result.DuplicateCount)
	fmt.Fprintf(&b, "- Retry replacements selected: %d\n", result.ReplacementCount)
	fmt.Fprintf(&b, "- Successful logical trials: %d\n", result.Succeeded)
	fmt.Fprintf(&b, "- Parseable logical trials: %d\n", result.ParseOK)
	fmt.Fprintf(&b, "- Suspected bleed: %d\n", result.SuspectedBleed)
	fmt.Fprintf(&b, "- Forbidden hits: %d\n", result.ForbiddenHits)
	fmt.Fprintf(&b, "- Average target-only score: %.3f\n\n", result.AverageScore)

	fmt.Fprintf(&b, "## By Scenario\n\n")
	fmt.Fprintf(&b, "| Scenario | Trials | Succeeded | Parse OK | Schema repaired | Suspected bleed | Forbidden hits | Average score | Minimum score |\n")
	fmt.Fprintf(&b, "|---|---:|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, agg := range result.ByScenario {
		fmt.Fprintf(&b, "| `%s` | %d | %d | %d | %d | %d | %d | %.3f | %.3f |\n", agg.Scenario, agg.Trials, agg.Succeeded, agg.ParseOK, agg.SchemaRepaired, agg.SuspectedBleed, agg.ForbiddenHits, agg.AverageScore, agg.MinimumScore)
	}

	fmt.Fprintf(&b, "\n## By Page\n\n")
	fmt.Fprintf(&b, "| Page | Trials | Succeeded | Parse OK | Suspected bleed | Forbidden hits | Average score | Minimum score |\n")
	fmt.Fprintf(&b, "|---:|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, agg := range result.ByPage {
		fmt.Fprintf(&b, "| %d | %d | %d | %d | %d | %d | %.3f | %.3f |\n", agg.TargetPage, agg.Trials, agg.Succeeded, agg.ParseOK, agg.SuspectedBleed, agg.ForbiddenHits, agg.AverageScore, agg.MinimumScore)
	}

	fmt.Fprintf(&b, "\n## Low or Notable Trials\n\n")
	if len(result.LowTrials) == 0 {
		fmt.Fprintf(&b, "No low-scoring, failed, bleed, or forbidden-hit trials.\n")
		return b.String()
	}
	fmt.Fprintf(&b, "| Trial | Page | Scenario | Status | Parse strategy | Expected hits | Expected total | Forbidden hits | Bleed | Score | Source |\n")
	fmt.Fprintf(&b, "|---|---:|---|---|---|---:|---:|---:|---|---:|---|\n")
	for _, trial := range result.LowTrials {
		forbidden := trial.Metrics.ForbiddenPhraseHits + trial.Metrics.ForbiddenCaptionHits
		fmt.Fprintf(&b, "| %s | %d | `%s` | %s | %s | %d | %d | %d | %t | %.3f | `%s` |\n", trial.ID, trial.TargetPage, trial.Scenario, trial.Status, trial.Metrics.ParseStrategy, trial.Metrics.ExpectedPhraseHits, trial.Metrics.ExpectedPhraseTotal, forbidden, trial.Metrics.SuspectedBleed, trial.Metrics.TargetOnlyScore, trial.SourceOutDir)
	}
	return b.String()
}

func reportToJSON(result *ReportResult) (string, error) {
	body, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(append(body, '\n')), nil
}
