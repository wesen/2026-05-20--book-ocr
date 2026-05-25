package ocrpipeline

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestBuildStructuredOCRInputTurnContainsExactlyOneTargetImage(t *testing.T) {
	turn, err := BuildStructuredOCRInputTurn(StructuredOCRInput{BookID: "report-794", PageNumber: 32, ImagePath: "page_032.png"}, []byte("fake-png"))
	require.NoError(t, err)
	require.Equal(t, PageTurnID(32, 1, "structured-ocr"), turn.ID)
	require.Equal(t, 1, CountTurnImages(turn))
}

func TestRunStructuredPageDryRunWritesArtifactsAndMarkdownTable(t *testing.T) {
	ctx := context.Background()
	workDir := t.TempDir()
	imagePath := filepath.Join(workDir, "page_032.png")
	require.NoError(t, os.WriteFile(imagePath, []byte("fake-png"), 0o644))

	result, err := RunStructuredPage(ctx, StructuredOCRInput{BookID: "report-794", RunID: "test-run", PageNumber: 32, ImagePath: imagePath, WorkDir: workDir, DryRun: true}, DryRunStructuredOCRClient{})
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(workDir, "pages/page_032/01-turn-input.yaml"))
	require.FileExists(t, filepath.Join(workDir, "pages/page_032/02-turn-final.yaml"))
	require.FileExists(t, result.RawResponse)
	require.FileExists(t, result.StructuredJSON)
	require.FileExists(t, result.RenderedMD)
	require.FileExists(t, result.ValidationJSON)

	md, err := os.ReadFile(result.RenderedMD)
	require.NoError(t, err)
	require.Contains(t, string(md), "<!-- page:032 -->")
	require.Contains(t, string(md), "Figure 2-2: PPSCalc -- Formula Display")
	require.Contains(t, string(md), "|  | A | B | C |")
	require.Contains(t, string(md), "| 1 | 100 | 20 | A1*B1 |")
	require.Contains(t, string(md), "Figure 2-3: PPSCalc -- Value Display")
	require.Contains(t, string(md), "| 3 |  |  | 2375 |")

	db, err := sql.Open("sqlite3", filepath.Join(workDir, "turns.db"))
	require.NoError(t, err)
	defer db.Close()
	var turnsCount int
	require.NoError(t, db.QueryRow(`select count(*) from turns`).Scan(&turnsCount))
	require.Equal(t, 1, turnsCount)
	var phases string
	require.NoError(t, db.QueryRow(`select group_concat(distinct phase) from turn_block_membership`).Scan(&phases))
	require.Contains(t, phases, "input")
	require.Contains(t, phases, "final")
}

func TestParseStructuredOCRResponseValidatesRequiredFields(t *testing.T) {
	_, err := ParseStructuredOCRResponse(`{}`)
	require.Error(t, err)
	page, err := ParseStructuredOCRResponse(`{"schema_version":"structured-ocr/v1","book_id":"report-794","page_number":32,"page_type":"body","blocks":[]}`)
	require.NoError(t, err)
	require.Equal(t, 32, page.PageNumber)
}
