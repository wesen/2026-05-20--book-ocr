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

type RescoreConfig struct {
	OutDir     string
	SQLitePath string
	Write      bool
}

type RescoreResult struct {
	RunID   string        `json:"run_id"`
	Trials  []TrialResult `json:"trials"`
	Summary Summary       `json:"summary"`
}

func RescoreOutputDir(ctx context.Context, cfg RescoreConfig) (*RescoreResult, error) {
	if strings.TrimSpace(cfg.OutDir) == "" {
		return nil, fmt.Errorf("out-dir is required")
	}
	if strings.TrimSpace(cfg.SQLitePath) == "" {
		cfg.SQLitePath = DefaultSQLitePath(cfg.OutDir)
	}
	manifest, _ := readManifest(filepath.Join(cfg.OutDir, "manifest.json"))
	trialPaths, err := filepath.Glob(filepath.Join(cfg.OutDir, "trials", "trial-*", "trial.json"))
	if err != nil {
		return nil, err
	}
	if len(trialPaths) == 0 {
		return nil, fmt.Errorf("no trial.json files found under %s", filepath.Join(cfg.OutDir, "trials"))
	}
	sort.Strings(trialPaths)

	var db *ResultDB
	if cfg.Write {
		db, err = OpenResultDB(cfg.SQLitePath)
		if err != nil {
			return nil, err
		}
		defer db.Close()
	}

	files := NewFileLogger(cfg.OutDir)
	results := make([]TrialResult, 0, len(trialPaths))
	for _, path := range trialPaths {
		trial, err := readTrialResult(path)
		if err != nil {
			return nil, err
		}
		rescored, err := rescoreTrialResult(trial)
		if err != nil {
			return nil, fmt.Errorf("rescore %s: %w", path, err)
		}
		if cfg.Write {
			if err := files.WriteTrialArtifacts(&rescored); err != nil {
				return nil, err
			}
			if err := db.InsertTrial(ctx, rescored); err != nil {
				return nil, err
			}
		}
		results = append(results, rescored)
	}

	runID := ""
	if manifest.ID != "" {
		runID = manifest.ID
	} else if len(results) > 0 {
		runID = results[0].RunID
	}
	summary := Summarize(runID, results)
	if cfg.Write {
		if err := files.WriteSummary(summary); err != nil {
			return nil, err
		}
		if db != nil && runID != "" {
			if err := db.CompleteRun(ctx, runID, summary); err != nil {
				return nil, err
			}
		}
		if manifest.ID != "" {
			now := time.Now().UTC()
			manifest.CompletedAt = &now
			manifest.SQLitePath = cfg.SQLitePath
			if err := files.WriteManifest(manifest); err != nil {
				return nil, err
			}
		}
	}
	return &RescoreResult{RunID: runID, Trials: results, Summary: summary}, nil
}

func rescoreTrialResult(result TrialResult) (TrialResult, error) {
	raw := strings.TrimSpace(result.RawResponse)
	if result.ResponsePath != "" {
		if body, err := os.ReadFile(result.ResponsePath); err == nil && strings.TrimSpace(string(body)) != "" {
			raw = strings.TrimSpace(string(body))
		}
	}
	if raw == "" {
		result.RawResponse = ""
		result.Parsed = nil
		result.Metrics = ScoreTrial("", nil, OracleForPage(result.TargetPage), fmt.Errorf("empty raw response"))
		result.Metrics.ParseStrategy = "missing-response"
		if result.Status == "" || result.Status == "succeeded" {
			result.Status = "failed"
		}
		if result.Error == "" {
			result.Error = "empty raw response"
		}
		if result.CompletedAt.IsZero() {
			result.CompletedAt = time.Now().UTC()
		}
		return result, nil
	}
	parseResult := ParseBenchmarkResponseDetailed(raw)
	if parseResult.Response != nil && parseResult.Response.TargetPage == 0 {
		parseResult.Response.TargetPage = result.TargetPage
		parseResult.SchemaRepaired = true
		if parseResult.Strategy == "strict-json" {
			parseResult.Strategy = "schema-repair"
		}
	}
	metrics := ScoreTrial(raw, parseResult.Response, OracleForPage(result.TargetPage), parseResult.Error)
	metrics.JSONSanitized = parseResult.Sanitized
	metrics.SchemaRepaired = parseResult.SchemaRepaired
	metrics.ParseStrategy = parseResult.Strategy
	result.RawResponse = raw
	result.Parsed = parseResult.Response
	result.Metrics = metrics
	result.Status = "succeeded"
	result.Error = ""
	if parseResult.Error != nil {
		result.Status = "parse_failed"
		result.Error = parseResult.Error.Error()
	}
	if result.CompletedAt.IsZero() {
		result.CompletedAt = time.Now().UTC()
	}
	return result, nil
}

func readManifest(path string) (RunManifest, error) {
	var manifest RunManifest
	body, err := os.ReadFile(path)
	if err != nil {
		return manifest, err
	}
	if err := json.Unmarshal(body, &manifest); err != nil {
		return manifest, err
	}
	return manifest, nil
}

func readTrialResult(path string) (TrialResult, error) {
	var result TrialResult
	body, err := os.ReadFile(path)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return result, err
	}
	if result.ResponsePath == "" {
		result.ResponsePath = filepath.Join(filepath.Dir(path), "response.txt")
	}
	return result, nil
}
