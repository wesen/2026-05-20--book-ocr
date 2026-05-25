package vlmseparation

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

type BenchmarkCommand struct {
	*cmds.CommandDescription
}

type RescoreCommand struct {
	*cmds.CommandDescription
}

type ReportCommand struct {
	*cmds.CommandDescription
}

type BenchmarkSettings struct {
	BookID            string   `glazed:"book-id"`
	ImageDir          string   `glazed:"image-dir"`
	TargetPages       string   `glazed:"target-pages"`
	Scenarios         []string `glazed:"scenarios"`
	Preset            string   `glazed:"preset"`
	OutDir            string   `glazed:"out-dir"`
	SQLitePath        string   `glazed:"sqlite"`
	TurnsDB           string   `glazed:"turns-db"`
	TurnsDSN          string   `glazed:"turns-dsn"`
	Profile           string   `glazed:"profile"`
	ProfileRegistries []string `glazed:"profile-registries"`
	DryRun            bool     `glazed:"dry-run"`
	MaxTrials         int      `glazed:"max-trials"`
	RunID             string   `glazed:"run-id"`
}

type RescoreSettings struct {
	OutDir     string `glazed:"out-dir"`
	SQLitePath string `glazed:"sqlite"`
	Write      bool   `glazed:"write"`
}

type ReportSettings struct {
	OutDirs      []string `glazed:"out-dir"`
	ReportDir    string   `glazed:"report-dir"`
	MarkdownPath string   `glazed:"markdown"`
	JSONPath     string   `glazed:"json-path"`
	Write        bool     `glazed:"write"`
}

func NewBenchmarkCommand() (*BenchmarkCommand, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	desc := cmds.NewCommandDescription(
		"benchmark",
		cmds.WithShort("Benchmark VLM target-page separation under multi-image prompt layouts"),
		cmds.WithLong(`Benchmark whether a VLM can keep a target page separate from neighboring page images.

The command runs page/scenario trials, writes detailed files, writes benchmark metrics to SQLite,
and stores exact Geppetto turns in a Pinocchio-compatible turns DB.

Examples:
  book-ocr vlm-separation benchmark --dry-run --book-id report-794 --image-dir /path/to/pages --target-pages 12,13 --scenarios target-only,multi-block-labeled --out-dir /tmp/vlmsep --output table
`),
		cmds.WithFlags(
			fields.New("book-id", fields.TypeString, fields.WithHelp("Book identifier")),
			fields.New("image-dir", fields.TypeString, fields.WithHelp("Directory containing page_NNN.png images")),
			fields.New("target-pages", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Comma-separated target pages or ranges, e.g. 12,13,31-32")),
			fields.New("scenarios", fields.TypeStringList, fields.WithDefault([]string{ScenarioTargetOnly}), fields.WithHelp("Scenario names; accepts repeated values or comma-separated lists")),
			fields.New("preset", fields.TypeString, fields.WithDefault("report794-bleed-smoke"), fields.WithHelp("Page preset used when --target-pages is empty")),
			fields.New("out-dir", fields.TypeString, fields.WithDefault("/tmp/book-ocr-vlm-separation"), fields.WithHelp("Output directory for manifest, trial files, and default DBs")),
			fields.New("sqlite", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Benchmark result SQLite path; defaults to <out-dir>/results.sqlite")),
			fields.New("turns-db", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Pinocchio turns DB path; defaults to <out-dir>/turns.db")),
			fields.New("turns-dsn", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Pinocchio turns DB SQLite DSN; overrides --turns-db")),
			fields.New("profile", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Pinocchio profile slug for live runs")),
			fields.New("profile-registries", fields.TypeStringList, fields.WithDefault([]string{}), fields.WithHelp("Pinocchio profile registry files/URLs")),
			fields.New("dry-run", fields.TypeBool, fields.WithDefault(true), fields.WithHelp("Use deterministic fake VLM responses instead of live inference")),
			fields.New("max-trials", fields.TypeInteger, fields.WithDefault(0), fields.WithHelp("Optional maximum number of trials")),
			fields.New("run-id", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Optional stable benchmark run ID")),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)
	return &BenchmarkCommand{CommandDescription: desc}, nil
}

func NewRescoreCommand() (*RescoreCommand, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	desc := cmds.NewCommandDescription(
		"rescore",
		cmds.WithShort("Re-score an existing VLM separation benchmark output directory"),
		cmds.WithLong(`Re-score saved benchmark trial responses without calling the provider.

The command reads <out-dir>/trials/trial-*/trial.json and response.txt, reparses each response
with the current JSON sanitizer/schema-repair logic, recomputes metrics, and optionally rewrites
trial artifacts, summary files, and SQLite metric rows.

Examples:
  book-ocr vlm-separation rescore --out-dir /tmp/book-ocr-vlm-separation-live-001 --output table
  book-ocr vlm-separation rescore --out-dir /tmp/book-ocr-vlm-separation-live-001 --write=false --output json
`),
		cmds.WithFlags(
			fields.New("out-dir", fields.TypeString, fields.WithHelp("Existing benchmark output directory")),
			fields.New("sqlite", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Benchmark result SQLite path; defaults to <out-dir>/results.sqlite")),
			fields.New("write", fields.TypeBool, fields.WithDefault(true), fields.WithHelp("Rewrite trial artifacts, summary files, and SQLite metric rows")),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)
	return &RescoreCommand{CommandDescription: desc}, nil
}

func NewReportCommand() (*ReportCommand, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	desc := cmds.NewCommandDescription(
		"report",
		cmds.WithShort("Build a grouped Markdown/JSON report from VLM separation benchmark runs"),
		cmds.WithLong(`Build a grouped report from one or more saved benchmark output directories.

The command reads saved trial artifacts, applies the current parser/scorer, merges duplicate
logical cells by target page and scenario, and writes report.md/report.json by default. Passing
multiple --out-dir values lets retry runs replace failed cells from a larger run.

Examples:
  book-ocr vlm-separation report --out-dir /tmp/main-run --output table
  book-ocr vlm-separation report --out-dir /tmp/main-run --out-dir /tmp/retry-run --report-dir /tmp/main-run --output table
`),
		cmds.WithFlags(
			fields.New("out-dir", fields.TypeStringList, fields.WithHelp("Benchmark output directory; repeat for retry/replacement runs")),
			fields.New("report-dir", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Directory for report.md/report.json; defaults to first out-dir")),
			fields.New("markdown", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Explicit Markdown report path")),
			fields.New("json-path", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Explicit JSON report path")),
			fields.New("write", fields.TypeBool, fields.WithDefault(true), fields.WithHelp("Write report.md and report.json")),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)
	return &ReportCommand{CommandDescription: desc}, nil
}

func (c *BenchmarkCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &BenchmarkSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	pages, err := ParsePages(s.TargetPages, s.Preset)
	if err != nil {
		return err
	}
	cfg := Config{
		BookID:            s.BookID,
		ImageDir:          s.ImageDir,
		TargetPages:       pages,
		Scenarios:         normalizeList(s.Scenarios),
		Preset:            s.Preset,
		OutDir:            s.OutDir,
		SQLitePath:        s.SQLitePath,
		TurnsDB:           s.TurnsDB,
		TurnsDSN:          s.TurnsDSN,
		Profile:           s.Profile,
		ProfileRegistries: normalizeList(s.ProfileRegistries),
		DryRun:            s.DryRun,
		MaxTrials:         s.MaxTrials,
		RunID:             s.RunID,
	}
	runner, closeFn, err := NewRunner(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeFn()
	result, err := runner.Run(ctx)
	if err != nil {
		return err
	}
	for _, trial := range result.Trials {
		row := types.NewRow(
			types.MRP("run_id", trial.RunID),
			types.MRP("trial_id", trial.ID),
			types.MRP("scenario", trial.Scenario),
			types.MRP("target_page", trial.TargetPage),
			types.MRP("status", trial.Status),
			types.MRP("json_parse_ok", trial.Metrics.JSONParseOK),
			types.MRP("json_sanitized", trial.Metrics.JSONSanitized),
			types.MRP("schema_repaired", trial.Metrics.SchemaRepaired),
			types.MRP("parse_strategy", trial.Metrics.ParseStrategy),
			types.MRP("suspected_bleed", trial.Metrics.SuspectedBleed),
			types.MRP("target_only_score", trial.Metrics.TargetOnlyScore),
			types.MRP("forbidden_phrase_hits", trial.Metrics.ForbiddenPhraseHits),
			types.MRP("turn_session_id", trial.TurnSessionID),
			types.MRP("turn_id", trial.TurnID),
			types.MRP("out_dir", result.Manifest.OutDir),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *RescoreCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &RescoreSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	result, err := RescoreOutputDir(ctx, RescoreConfig{OutDir: s.OutDir, SQLitePath: s.SQLitePath, Write: s.Write})
	if err != nil {
		return err
	}
	for _, trial := range result.Trials {
		row := types.NewRow(
			types.MRP("run_id", trial.RunID),
			types.MRP("trial_id", trial.ID),
			types.MRP("scenario", trial.Scenario),
			types.MRP("target_page", trial.TargetPage),
			types.MRP("status", trial.Status),
			types.MRP("json_parse_ok", trial.Metrics.JSONParseOK),
			types.MRP("json_sanitized", trial.Metrics.JSONSanitized),
			types.MRP("schema_repaired", trial.Metrics.SchemaRepaired),
			types.MRP("parse_strategy", trial.Metrics.ParseStrategy),
			types.MRP("suspected_bleed", trial.Metrics.SuspectedBleed),
			types.MRP("target_only_score", trial.Metrics.TargetOnlyScore),
			types.MRP("expected_phrase_hits", trial.Metrics.ExpectedPhraseHits),
			types.MRP("expected_phrase_total", trial.Metrics.ExpectedPhraseTotal),
			types.MRP("forbidden_phrase_hits", trial.Metrics.ForbiddenPhraseHits),
			types.MRP("out_dir", s.OutDir),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *ReportCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &ReportSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	result, err := BuildReport(ctx, ReportConfig{OutDirs: s.OutDirs, ReportDir: s.ReportDir, MarkdownPath: s.MarkdownPath, JSONPath: s.JSONPath, Write: s.Write})
	if err != nil {
		return err
	}
	for _, agg := range result.ByScenario {
		row := types.NewRow(
			types.MRP("kind", "scenario"),
			types.MRP("name", agg.Scenario),
			types.MRP("trials", agg.Trials),
			types.MRP("succeeded", agg.Succeeded),
			types.MRP("parse_ok", agg.ParseOK),
			types.MRP("schema_repaired", agg.SchemaRepaired),
			types.MRP("suspected_bleed", agg.SuspectedBleed),
			types.MRP("forbidden_hits", agg.ForbiddenHits),
			types.MRP("average_score", agg.AverageScore),
			types.MRP("minimum_score", agg.MinimumScore),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	for _, agg := range result.ByPage {
		row := types.NewRow(
			types.MRP("kind", "page"),
			types.MRP("name", fmt.Sprintf("%03d", agg.TargetPage)),
			types.MRP("trials", agg.Trials),
			types.MRP("succeeded", agg.Succeeded),
			types.MRP("parse_ok", agg.ParseOK),
			types.MRP("schema_repaired", ""),
			types.MRP("suspected_bleed", agg.SuspectedBleed),
			types.MRP("forbidden_hits", agg.ForbiddenHits),
			types.MRP("average_score", agg.AverageScore),
			types.MRP("minimum_score", agg.MinimumScore),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func NewRootCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "vlm-separation",
		Short: "Investigate VLM target/context page separation",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.InitLoggerFromCobra(cmd)
		},
	}
	if err := logging.AddLoggingSectionToRootCommand(root, "book-ocr"); err != nil {
		return nil, err
	}
	benchmark, err := NewBenchmarkCommand()
	if err != nil {
		return nil, err
	}
	cobraBenchmark, err := buildGlazedCobraCommand(benchmark)
	if err != nil {
		return nil, err
	}
	rescore, err := NewRescoreCommand()
	if err != nil {
		return nil, err
	}
	cobraRescore, err := buildGlazedCobraCommand(rescore)
	if err != nil {
		return nil, err
	}
	report, err := NewReportCommand()
	if err != nil {
		return nil, err
	}
	cobraReport, err := buildGlazedCobraCommand(report)
	if err != nil {
		return nil, err
	}
	root.AddCommand(cobraBenchmark, cobraRescore, cobraReport)
	return root, nil
}

func buildGlazedCobraCommand(command cmds.Command) (*cobra.Command, error) {
	return cli.BuildCobraCommandFromCommand(command,
		cli.WithParserConfig(cli.CobraParserConfig{
			AppName:           "book-ocr",
			ShortHelpSections: []string{schema.DefaultSlug},
			MiddlewaresFunc:   cli.CobraCommandDefaultMiddlewares,
		}),
	)
}

func ExecuteRoot(ctx context.Context, args []string) error {
	root, err := NewRootCommand()
	if err != nil {
		return err
	}
	root.SetArgs(args)
	root.SetContext(ctx)
	if err := root.Execute(); err != nil {
		return fmt.Errorf("vlm-separation: %w", err)
	}
	return nil
}

func normalizeList(values []string) []string {
	var out []string
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}
