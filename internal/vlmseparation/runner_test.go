package vlmseparation

import (
	"context"
	"database/sql"
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
