package vlmseparation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/geppetto/pkg/inference/engine"
	"github.com/go-go-golems/geppetto/pkg/turns"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/pinocchio/pkg/cmds/profilebootstrap"
	"github.com/google/uuid"
)

type Runner struct {
	Config    Config
	Scenarios []Scenario
	Files     *FileLogger
	DB        *ResultDB
	Turns     *TurnStore
	Engine    engine.Engine
}

func NewRunner(ctx context.Context, cfg Config) (*Runner, func(), error) {
	if strings.TrimSpace(cfg.BookID) == "" {
		return nil, nil, fmt.Errorf("book-id is required")
	}
	if strings.TrimSpace(cfg.ImageDir) == "" {
		return nil, nil, fmt.Errorf("image-dir is required")
	}
	if strings.TrimSpace(cfg.OutDir) == "" {
		cfg.OutDir = filepath.Join(os.TempDir(), "book-ocr-vlm-separation")
	}
	if strings.TrimSpace(cfg.SQLitePath) == "" {
		cfg.SQLitePath = DefaultSQLitePath(cfg.OutDir)
	}
	if strings.TrimSpace(cfg.RunID) == "" {
		cfg.RunID = "vlmsep-" + uuid.NewString()
	}
	scenarios, err := NormalizeScenarios(cfg.Scenarios)
	if err != nil {
		return nil, nil, err
	}
	files := NewFileLogger(cfg.OutDir)
	if err := files.Init(); err != nil {
		return nil, nil, err
	}
	db, err := OpenResultDB(cfg.SQLitePath)
	if err != nil {
		return nil, nil, err
	}
	closers := []func(){func() { _ = db.Close() }}
	convID := fmt.Sprintf("vlm-separation:%s:%s", cfg.BookID, cfg.RunID)
	turnStore, closeTurns, err := OpenTurnStore(ctx, cfg.TurnsDSN, cfg.TurnsDB, cfg.OutDir, convID)
	if err != nil {
		for _, c := range closers {
			c()
		}
		return nil, nil, err
	}
	closers = append(closers, closeTurns)
	r := &Runner{Config: cfg, Scenarios: scenarios, Files: files, DB: db, Turns: turnStore}
	if !cfg.DryRun {
		eng, closeResolved, err := buildEngine(ctx, cfg)
		if err != nil {
			for _, c := range closers {
				c()
			}
			return nil, nil, err
		}
		if closeResolved != nil {
			closers = append(closers, closeResolved)
		}
		r.Engine = eng
	}
	return r, func() {
		for i := len(closers) - 1; i >= 0; i-- {
			closers[i]()
		}
	}, nil
}

func (r *Runner) Run(ctx context.Context) (*RunResult, error) {
	manifest := RunManifest{ID: r.Config.RunID, BookID: r.Config.BookID, ImageDir: r.Config.ImageDir, TargetPages: r.Config.TargetPages, Scenarios: r.Scenarios, Profile: r.Config.Profile, ProfileRegistries: r.Config.ProfileRegistries, DryRun: r.Config.DryRun, OutDir: r.Config.OutDir, SQLitePath: r.Config.SQLitePath, TurnsDB: r.Config.TurnsDB, TurnsDSN: r.Turns.DSN(), StartedAt: time.Now().UTC()}
	if err := r.Files.WriteManifest(manifest); err != nil {
		return nil, err
	}
	if err := r.Files.WriteScenarios(r.Scenarios); err != nil {
		return nil, err
	}
	if err := r.DB.InsertRun(ctx, manifest); err != nil {
		return nil, err
	}
	var results []TrialResult
	trialNum := 0
	for _, page := range r.Config.TargetPages {
		for _, scenario := range r.Scenarios {
			if r.Config.MaxTrials > 0 && len(results) >= r.Config.MaxTrials {
				break
			}
			trialNum++
			trial := Trial{ID: fmt.Sprintf("trial-%04d", trialNum), RunID: manifest.ID, Scenario: scenario, TargetPage: page, PreviousPage: page - 1, NextPage: page + 1}
			result, err := r.RunTrial(ctx, trial)
			if err != nil {
				now := time.Now().UTC()
				result = TrialResult{ID: trial.ID, RunID: trial.RunID, Scenario: scenario.Name, TargetPage: page, PreviousPage: page - 1, NextPage: page + 1, Status: "failed", StartedAt: now, CompletedAt: now, Error: err.Error(), TurnSessionID: sessionID(page, scenario.Name), TurnID: turnID(trial.ID)}
				_ = r.Files.WriteTrialArtifacts(&result)
			}
			if err := r.DB.InsertTrial(ctx, result); err != nil {
				return nil, err
			}
			results = append(results, result)
		}
	}
	summary := Summarize(manifest.ID, results)
	now := time.Now().UTC()
	manifest.CompletedAt = &now
	_ = r.Files.WriteManifest(manifest)
	if err := r.Files.WriteSummary(summary); err != nil {
		return nil, err
	}
	if err := r.DB.CompleteRun(ctx, manifest.ID, summary); err != nil {
		return nil, err
	}
	return &RunResult{Manifest: manifest, Trials: results, Summary: summary}, nil
}

func (r *Runner) RunTrial(ctx context.Context, trial Trial) (TrialResult, error) {
	started := time.Now().UTC()
	input := r.buildTrialInput(trial)
	turn, err := BuildTurnForScenario(input)
	if err != nil {
		return TrialResult{}, err
	}
	requestPath, err := r.Files.WriteTurn(trial.ID, "turn-input.yaml", turn)
	if err != nil {
		return TrialResult{}, err
	}
	if err := r.Turns.Save(ctx, input.SessionID, input.TurnID, "input", turn); err != nil {
		return TrialResult{}, err
	}
	startInference := time.Now()
	var updated *turns.Turn
	if r.Config.DryRun {
		updated = fakeVLMResponse(turn, input)
	} else {
		updated, err = r.Engine.RunInference(ctx, turn)
		if err != nil {
			return TrialResult{}, err
		}
	}
	latency := time.Since(startInference)
	if _, err := r.Files.WriteTurn(trial.ID, "turn-final.yaml", updated); err != nil {
		return TrialResult{}, err
	}
	if err := r.Turns.Save(ctx, input.SessionID, input.TurnID, "final", updated); err != nil {
		return TrialResult{}, err
	}
	raw, err := lastLLMText(updated)
	if err != nil {
		return TrialResult{}, err
	}
	parseResult := ParseBenchmarkResponseDetailed(raw)
	if parseResult.Response != nil && parseResult.Response.TargetPage == 0 {
		parseResult.Response.TargetPage = input.TargetPage
		parseResult.SchemaRepaired = true
		if parseResult.Strategy == "strict-json" {
			parseResult.Strategy = "schema-repair"
		}
	}
	metrics := ScoreTrial(raw, parseResult.Response, input.Oracle, parseResult.Error)
	metrics.JSONSanitized = parseResult.Sanitized
	metrics.SchemaRepaired = parseResult.SchemaRepaired
	metrics.ParseStrategy = parseResult.Strategy
	status := "succeeded"
	if parseResult.Error != nil {
		status = "parse_failed"
	}
	result := TrialResult{ID: trial.ID, RunID: trial.RunID, Scenario: trial.Scenario.Name, TargetPage: trial.TargetPage, PreviousPage: trial.PreviousPage, NextPage: trial.NextPage, Status: status, RequestPath: requestPath, TurnSessionID: input.SessionID, TurnID: input.TurnID, StartedAt: started, CompletedAt: time.Now().UTC(), LatencyMS: latency.Milliseconds(), RawResponse: raw, Parsed: parseResult.Response, Metrics: metrics}
	if parseResult.Error != nil {
		result.Error = parseResult.Error.Error()
	}
	if err := r.Files.WriteTrialArtifacts(&result); err != nil {
		return TrialResult{}, err
	}
	return result, nil
}

func (r *Runner) buildTrialInput(trial Trial) TrialInput {
	return TrialInput{TrialID: trial.ID, RunID: trial.RunID, BookID: r.Config.BookID, Scenario: trial.Scenario, TargetPage: trial.TargetPage, PreviousPage: trial.PreviousPage, NextPage: trial.NextPage, TargetImagePath: pagePath(r.Config.ImageDir, trial.TargetPage), PreviousImagePath: existingPagePath(r.Config.ImageDir, trial.PreviousPage), NextImagePath: existingPagePath(r.Config.ImageDir, trial.NextPage), PreviousText: fmt.Sprintf("Previous page %03d text context placeholder.", trial.PreviousPage), NextText: fmt.Sprintf("Next page %03d text context placeholder.", trial.NextPage), SessionID: sessionID(trial.TargetPage, trial.Scenario.Name), TurnID: turnID(trial.ID), Oracle: OracleForPage(trial.TargetPage)}
}

func pagePath(imageDir string, page int) string {
	return filepath.Join(imageDir, fmt.Sprintf("page_%03d.png", page))
}
func existingPagePath(imageDir string, page int) string {
	p := pagePath(imageDir, page)
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}
func sessionID(page int, scenario string) string {
	return fmt.Sprintf("page:%03d:scenario:%s", page, scenario)
}
func turnID(trialID string) string { return "trial:" + trialID }

func fakeVLMResponse(turn *turns.Turn, input TrialInput) *turns.Turn {
	response := BenchmarkResponse{TargetPage: input.TargetPage, TranscribedPageIdentity: PageIdentity{TitleOrCaptionLines: input.Oracle.ExpectedCaptions}, ContentMarkers: ContentMarkers{FigureCaptions: input.Oracle.ExpectedCaptions, UniquePhrases: input.Oracle.ExpectedPhrases}, Transcription: strings.Join(append(append([]string{}, input.Oracle.ExpectedCaptions...), input.Oracle.ExpectedPhrases...), "\n"), SuspectedContextCopy: false}
	if input.Scenario.Name == ScenarioContextFirstNegativeControl && len(input.Oracle.ForbiddenCaptions) > 0 {
		response.ContentMarkers.FigureCaptions = append(response.ContentMarkers.FigureCaptions, input.Oracle.ForbiddenCaptions...)
		response.Transcription += "\n" + strings.Join(input.Oracle.ForbiddenCaptions, "\n")
		response.SuspectedContextCopy = true
	}
	body, _ := json.Marshal(response)
	turns.AppendBlock(turn, turns.NewAssistantTextBlock(string(body)))
	return turn
}

func lastLLMText(turn *turns.Turn) (string, error) {
	blocks := turns.FindLastBlocksByKind(*turn, turns.BlockKindLLMText)
	if len(blocks) == 0 {
		return "", fmt.Errorf("no llm text block")
	}
	text, _ := blocks[len(blocks)-1].Payload[turns.PayloadKeyText].(string)
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("empty llm text")
	}
	return strings.TrimSpace(text), nil
}

func buildEngine(ctx context.Context, cfg Config) (engine.Engine, func(), error) {
	parsed, err := profilebootstrap.NewCLISelectionValues(profilebootstrap.CLISelectionInput{Profile: cfg.Profile, ProfileRegistries: cfg.ProfileRegistries})
	if err != nil {
		return nil, nil, err
	}
	// Keep values import live for compatibility with profilebootstrap APIs used elsewhere.
	_ = values.New
	resolved, err := profilebootstrap.ResolveCLIEngineSettings(ctx, parsed)
	if err != nil {
		return nil, nil, err
	}
	eng, err := profilebootstrap.NewEngineFromResolvedCLIEngineSettings(resolved)
	if err != nil {
		if resolved.Close != nil {
			resolved.Close()
		}
		return nil, nil, err
	}
	return eng, resolved.Close, nil
}
