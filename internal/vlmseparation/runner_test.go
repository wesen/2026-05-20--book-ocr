package vlmseparation

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestRunnerDryRunPersistsFilesSQLiteAndTurns(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	imageDir := filepath.Join(dir, "pages")
	require.NoError(t, os.MkdirAll(imageDir, 0o755))
	for _, page := range []int{11, 12, 13} {
		require.NoError(t, os.WriteFile(filepath.Join(imageDir, "page_"+formatPage(page)+".png"), []byte("fake image"), 0o644))
	}
	outDir := filepath.Join(dir, "out")
	cfg := Config{BookID: "report-794", ImageDir: imageDir, TargetPages: []int{12}, Scenarios: []string{ScenarioTargetOnly, ScenarioContextFirstNegativeControl}, OutDir: outDir, DryRun: true, RunID: "test-run"}
	runner, closeFn, err := NewRunner(ctx, cfg)
	require.NoError(t, err)
	defer closeFn()
	result, err := runner.Run(ctx)
	require.NoError(t, err)
	require.Len(t, result.Trials, 2)
	require.FileExists(t, filepath.Join(outDir, "manifest.json"))
	require.FileExists(t, filepath.Join(outDir, "summary.json"))
	require.FileExists(t, filepath.Join(outDir, "trials", "trial-0001", "turn-input.yaml"))
	require.FileExists(t, filepath.Join(outDir, "trials", "trial-0001", "turn-final.yaml"))

	resultsDB, err := sql.Open("sqlite3", filepath.Join(outDir, "results.sqlite"))
	require.NoError(t, err)
	defer resultsDB.Close()
	var trialCount int
	require.NoError(t, resultsDB.QueryRow(`select count(*) from benchmark_trials`).Scan(&trialCount))
	require.Equal(t, 2, trialCount)

	turnsDB, err := sql.Open("sqlite3", filepath.Join(outDir, "turns.db"))
	require.NoError(t, err)
	defer turnsDB.Close()
	var turnCount int
	require.NoError(t, turnsDB.QueryRow(`select count(*) from turns`).Scan(&turnCount))
	require.Equal(t, 2, turnCount)
}

func TestRescoreOutputDirRepairsSavedResponseAndUpdatesSQLite(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	imageDir := filepath.Join(dir, "pages")
	require.NoError(t, os.MkdirAll(imageDir, 0o755))
	for _, page := range []int{12, 13, 14} {
		require.NoError(t, os.WriteFile(filepath.Join(imageDir, "page_"+formatPage(page)+".png"), []byte("fake image"), 0o644))
	}
	outDir := filepath.Join(dir, "out")
	runner, closeFn, err := NewRunner(ctx, Config{BookID: "report-794", ImageDir: imageDir, TargetPages: []int{13}, Scenarios: []string{ScenarioTargetOnly}, OutDir: outDir, DryRun: true, RunID: "rescore-test"})
	require.NoError(t, err)
	_, err = runner.Run(ctx)
	require.NoError(t, err)
	closeFn()

	variant := `{"page_number":13,"text":"Figure 1-1: A Rudimentary User Interface\nApplication Data Base\nobservables\nqueries"}`
	require.NoError(t, os.WriteFile(filepath.Join(outDir, "trials", "trial-0001", "response.txt"), []byte(variant), 0o644))

	result, err := RescoreOutputDir(ctx, RescoreConfig{OutDir: outDir, Write: true})
	require.NoError(t, err)
	require.Len(t, result.Trials, 1)
	require.True(t, result.Trials[0].Metrics.JSONParseOK)
	require.True(t, result.Trials[0].Metrics.SchemaRepaired)
	require.Equal(t, "schema-repair", result.Trials[0].Metrics.ParseStrategy)
	require.Equal(t, 1.0, result.Trials[0].Metrics.TargetOnlyScore)

	resultsDB, err := sql.Open("sqlite3", filepath.Join(outDir, "results.sqlite"))
	require.NoError(t, err)
	defer resultsDB.Close()
	var schemaRepaired int
	var parseStrategy string
	require.NoError(t, resultsDB.QueryRow(`select schema_repaired, parse_strategy from trial_metrics where trial_id='trial-0001'`).Scan(&schemaRepaired, &parseStrategy))
	require.Equal(t, 1, schemaRepaired)
	require.Equal(t, "schema-repair", parseStrategy)
}

func TestBuildReportMergesRetryRunByLogicalPageScenario(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	mainDir := filepath.Join(dir, "main")
	retryDir := filepath.Join(dir, "retry")
	writeReportTrial := func(outDir, trialID, status, response string) {
		t.Helper()
		trialDir := filepath.Join(outDir, "trials", trialID)
		require.NoError(t, os.MkdirAll(trialDir, 0o755))
		responsePath := filepath.Join(trialDir, "response.txt")
		require.NoError(t, os.WriteFile(responsePath, []byte(response), 0o644))
		result := TrialResult{ID: trialID, RunID: "run", Scenario: ScenarioTargetOnly, TargetPage: 13, Status: status, ResponsePath: responsePath}
		if response == "" {
			result.Error = "provider failed"
		}
		body, err := json.Marshal(result)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(trialDir, "trial.json"), body, 0o644))
	}
	writeReportTrial(mainDir, "trial-0001", "failed", "")
	writeReportTrial(retryDir, "trial-0001", "succeeded", `{"page":13,"ocr":"Figure 1-1: A Rudimentary User Interface\nApplication Data Base\nobservables\nqueries"}`)

	report, err := BuildReport(ctx, ReportConfig{OutDirs: []string{mainDir, retryDir}, ReportDir: mainDir, Write: true})
	require.NoError(t, err)
	require.Equal(t, 2, report.TrialCount)
	require.Equal(t, 1, report.LogicalTrialCount)
	require.Equal(t, 1, report.DuplicateCount)
	require.Equal(t, 1, report.ReplacementCount)
	require.Equal(t, 1, report.Succeeded)
	require.Equal(t, 1, report.ParseOK)
	require.Equal(t, 1.0, report.AverageScore)
	require.FileExists(t, filepath.Join(mainDir, "report.md"))
	require.FileExists(t, filepath.Join(mainDir, "report.json"))
}
