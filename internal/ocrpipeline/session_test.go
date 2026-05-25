package ocrpipeline

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/geppetto/pkg/turns"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestOCRTurnStoreSavesInputAndFinalPhases(t *testing.T) {
	ctx := context.Background()
	workDir := t.TempDir()
	store, closeFn, err := OpenOCRTurnStore(TurnStoreConfig{WorkDir: workDir, ConvID: BookRunConvID("report-794", "test-run")})
	require.NoError(t, err)
	defer closeFn()

	turn := &turns.Turn{ID: PageTurnID(13, 1, "structured-ocr")}
	turns.AppendBlock(turn, turns.NewSystemTextBlock("structured OCR contract"))
	turns.AppendBlock(turn, turns.NewUserTextBlock("target page only"))
	sessionID := PageSessionID(13)
	turnID := PageTurnID(13, 1, "structured-ocr")
	require.NoError(t, store.Save(ctx, sessionID, turnID, "input", turn))
	turns.AppendBlock(turn, turns.NewAssistantTextBlock(`{"schema_version":"structured-ocr/v1"}`))
	require.NoError(t, store.Save(ctx, sessionID, turnID, "final", turn))

	db, err := sql.Open("sqlite3", filepath.Join(workDir, "turns.db"))
	require.NoError(t, err)
	defer db.Close()
	var count int
	require.NoError(t, db.QueryRow(`select count(*) from turns`).Scan(&count))
	require.Equal(t, 1, count)
	var phases string
	require.NoError(t, db.QueryRow(`select group_concat(distinct phase) from turn_block_membership`).Scan(&phases))
	require.Contains(t, phases, "input")
	require.Contains(t, phases, "final")
}

func TestTurnIdentifiers(t *testing.T) {
	require.Equal(t, "book-ocr:report-794:run-1", BookRunConvID("report-794", "run-1"))
	require.Equal(t, "page:013", PageSessionID(13))
	require.Equal(t, "page:013:02-normalize", PageTurnID(13, 2, "normalize"))
}
