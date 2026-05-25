package vlmseparation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ResultDB struct {
	db *sql.DB
}

func OpenResultDB(path string) (*ResultDB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	r := &ResultDB{db: db}
	if err := r.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return r, nil
}

func (r *ResultDB) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.Close()
}

func (r *ResultDB) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS benchmark_runs (
			id TEXT PRIMARY KEY,
			book_id TEXT NOT NULL,
			profile TEXT NOT NULL,
			prompt_version TEXT NOT NULL,
			started_at TEXT NOT NULL,
			completed_at TEXT,
			out_dir TEXT NOT NULL,
			turns_dsn TEXT,
			dry_run INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS benchmark_trials (
			id TEXT PRIMARY KEY,
			run_id TEXT NOT NULL,
			scenario TEXT NOT NULL,
			target_page INTEGER NOT NULL,
			previous_page INTEGER,
			next_page INTEGER,
			status TEXT NOT NULL,
			request_path TEXT NOT NULL,
			response_path TEXT,
			parsed_response_path TEXT,
			metrics_path TEXT,
			turn_session_id TEXT NOT NULL,
			turn_id TEXT NOT NULL,
			started_at TEXT NOT NULL,
			completed_at TEXT,
			latency_ms INTEGER,
			error TEXT,
			FOREIGN KEY(run_id) REFERENCES benchmark_runs(id)
		);`,
		`CREATE TABLE IF NOT EXISTS trial_metrics (
			trial_id TEXT PRIMARY KEY,
			json_parse_ok INTEGER NOT NULL,
			expected_phrase_hits INTEGER NOT NULL,
			expected_phrase_total INTEGER NOT NULL,
			forbidden_phrase_hits INTEGER NOT NULL,
			forbidden_caption_hits INTEGER NOT NULL,
			suspected_bleed INTEGER NOT NULL,
			target_only_score REAL NOT NULL,
			raw_json TEXT NOT NULL,
			FOREIGN KEY(trial_id) REFERENCES benchmark_trials(id)
		);`,
	}
	for _, stmt := range stmts {
		if _, err := r.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (r *ResultDB) InsertRun(ctx context.Context, manifest RunManifest) error {
	turnsDSN := manifest.TurnsDSN
	if turnsDSN == "" {
		turnsDSN = manifest.TurnsDB
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO benchmark_runs(id, book_id, profile, prompt_version, started_at, out_dir, turns_dsn, dry_run) VALUES(?,?,?,?,?,?,?,?)`, manifest.ID, manifest.BookID, manifest.Profile, "vlm-separation-v1", manifest.StartedAt.Format(time.RFC3339Nano), manifest.OutDir, turnsDSN, boolInt(manifest.DryRun))
	return err
}

func (r *ResultDB) CompleteRun(ctx context.Context, runID string, _ Summary) error {
	_, err := r.db.ExecContext(ctx, `UPDATE benchmark_runs SET completed_at=? WHERE id=?`, time.Now().UTC().Format(time.RFC3339Nano), runID)
	return err
}

func (r *ResultDB) InsertTrial(ctx context.Context, result TrialResult) error {
	_, err := r.db.ExecContext(ctx, `INSERT OR REPLACE INTO benchmark_trials(id, run_id, scenario, target_page, previous_page, next_page, status, request_path, response_path, parsed_response_path, metrics_path, turn_session_id, turn_id, started_at, completed_at, latency_ms, error) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, result.ID, result.RunID, result.Scenario, result.TargetPage, nullInt(result.PreviousPage), nullInt(result.NextPage), result.Status, result.RequestPath, result.ResponsePath, result.ParsedResponsePath, result.MetricsPath, result.TurnSessionID, result.TurnID, result.StartedAt.Format(time.RFC3339Nano), result.CompletedAt.Format(time.RFC3339Nano), result.LatencyMS, result.Error)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(result.Metrics)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT OR REPLACE INTO trial_metrics(trial_id, json_parse_ok, expected_phrase_hits, expected_phrase_total, forbidden_phrase_hits, forbidden_caption_hits, suspected_bleed, target_only_score, raw_json) VALUES(?,?,?,?,?,?,?,?,?)`, result.ID, boolInt(result.Metrics.JSONParseOK), result.Metrics.ExpectedPhraseHits, result.Metrics.ExpectedPhraseTotal, result.Metrics.ForbiddenPhraseHits, result.Metrics.ForbiddenCaptionHits, boolInt(result.Metrics.SuspectedBleed), result.Metrics.TargetOnlyScore, string(raw))
	return err
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func nullInt(v int) any {
	if v == 0 {
		return nil
	}
	return v
}

func DefaultSQLitePath(outDir string) string { return filepath.Join(outDir, "results.sqlite") }

func ValidateSQLitePath(path string) error {
	if path == "" {
		return fmt.Errorf("sqlite path is empty")
	}
	return nil
}
