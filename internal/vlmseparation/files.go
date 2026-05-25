package vlmseparation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/geppetto/pkg/turns"
	"github.com/go-go-golems/geppetto/pkg/turns/serde"
)

type FileLogger struct {
	OutDir string
}

func NewFileLogger(outDir string) *FileLogger { return &FileLogger{OutDir: outDir} }

func (l *FileLogger) Init() error {
	return os.MkdirAll(filepath.Join(l.OutDir, "trials"), 0o755)
}

func (l *FileLogger) WriteManifest(manifest RunManifest) error {
	return writeJSON(filepath.Join(l.OutDir, "manifest.json"), manifest)
}

func (l *FileLogger) WriteScenarios(scenarios []Scenario) error {
	return writeJSON(filepath.Join(l.OutDir, "scenarios.json"), scenarios)
}

func (l *FileLogger) TrialDir(trialID string) string {
	return filepath.Join(l.OutDir, "trials", trialID)
}

func (l *FileLogger) WriteTurn(trialID string, name string, t *turns.Turn) (string, error) {
	dir := l.TrialDir(trialID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	payload, err := serde.ToYAML(t, serde.Options{})
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	return path, os.WriteFile(path, payload, 0o644)
}

func (l *FileLogger) WriteTrialArtifacts(result *TrialResult) error {
	dir := l.TrialDir(result.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	result.ResponsePath = filepath.Join(dir, "response.txt")
	if err := os.WriteFile(result.ResponsePath, []byte(result.RawResponse), 0o644); err != nil {
		return err
	}
	if result.Parsed != nil {
		result.ParsedResponsePath = filepath.Join(dir, "response.json")
		if err := writeJSON(result.ParsedResponsePath, result.Parsed); err != nil {
			return err
		}
	}
	result.MetricsPath = filepath.Join(dir, "metrics.json")
	if err := writeJSON(result.MetricsPath, result.Metrics); err != nil {
		return err
	}
	return writeJSON(filepath.Join(dir, "trial.json"), result)
}

func (l *FileLogger) WriteSummary(summary Summary) error {
	if err := writeJSON(filepath.Join(l.OutDir, "summary.json"), summary); err != nil {
		return err
	}
	md := fmt.Sprintf(`# VLM Separation Benchmark Summary

- Run ID: %s
- Trials: %d
- Succeeded: %d
- Failed: %d
- Suspected bleed: %d
- Average target-only score: %.3f
`, summary.RunID, summary.TrialCount, summary.Succeeded, summary.Failed, summary.SuspectedBleed, summary.AverageTargetOnlyScore)
	return os.WriteFile(filepath.Join(l.OutDir, "summary.md"), []byte(md), 0o644)
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return os.WriteFile(path, body, 0o644)
}
