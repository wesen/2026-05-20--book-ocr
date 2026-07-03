package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"

	"github.com/go-go-golems/book-ocr/internal/bookprofile"
	"github.com/go-go-golems/book-ocr/internal/ocrmvp"
	"github.com/go-go-golems/book-ocr/internal/ocrpipeline"
	"github.com/go-go-golems/book-ocr/internal/ocrquality"
	"github.com/go-go-golems/book-ocr/internal/plugin"
	devctlruntime "github.com/go-go-golems/devctl/pkg/runtime"
	"github.com/go-go-golems/scraper/pkg/engine/model"
	"github.com/go-go-golems/scraper/pkg/workflow"
)

const defaultWorkDir = ".ocr-mvp"

type registryFlags []string

func (f *registryFlags) String() string { return strings.Join(*f, ",") }
func (f *registryFlags) Set(value string) error {
	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			*f = append(*f, trimmed)
		}
	}
	return nil
}

// pluginBindingFlags collects repeated --plugin seam=path values verbatim
// (no comma splitting: paths may contain commas).
type pluginBindingFlags []string

func (f *pluginBindingFlags) String() string { return strings.Join(*f, " ") }
func (f *pluginBindingFlags) Set(value string) error {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		*f = append(*f, trimmed)
	}
	return nil
}

// buildPluginSpecs parses --plugin bindings and reports whether an ocr.page
// binding is present (which lifts the --profile requirement for live runs:
// the plugin owns model access, not Geppetto).
func buildPluginSpecs(values []string) ([]plugin.Spec, bool, error) {
	var specs []plugin.Spec
	hasOCRPage := false
	for _, value := range values {
		spec, err := plugin.ParseSeamBinding(value)
		if err != nil {
			return nil, false, err
		}
		for _, seam := range spec.Seams {
			if seam == plugin.OpOCRPage {
				hasOCRPage = true
			}
		}
		specs = append(specs, spec)
	}
	return specs, hasOCRPage, nil
}

// pluginSpecsFromProfile loads a book profile's plugins: section. Plugin
// paths resolve relative to the profile file so a book workspace (profile +
// plugin scripts) is self-contained and relocatable.
func pluginSpecsFromProfile(profilePath string) ([]plugin.Spec, bool, error) {
	if strings.TrimSpace(profilePath) == "" {
		return nil, false, nil
	}
	profile, err := bookprofile.Load(profilePath)
	if err != nil {
		return nil, false, fmt.Errorf("load book profile %s: %w", profilePath, err)
	}
	baseDir := filepath.Dir(profilePath)
	var specs []plugin.Spec
	hasOCRPage := false
	for _, ref := range profile.Plugins {
		path := ref.Path
		if !filepath.IsAbs(path) {
			path = filepath.Join(baseDir, path)
		}
		for _, seam := range ref.Seams {
			if seam == plugin.OpOCRPage {
				hasOCRPage = true
			}
		}
		specs = append(specs, plugin.Spec{ID: ref.ID, Path: path, Args: append([]string(nil), ref.Args...), Env: ref.Env, Seams: append([]string(nil), ref.Seams...)})
	}
	return specs, hasOCRPage, nil
}

// pluginSeamSet bundles everything the plugin manager contributes to a
// structured workflow registration.
type pluginSeamSet struct {
	mgr            *plugin.Manager
	client         ocrpipeline.StructuredOCRClient
	segmenter      ocrquality.FigureSegmenter
	parser         ocrpipeline.ResponseParser
	pageValidators []ocrpipeline.PageValidator
	bookValidators []ocrpipeline.BookValidator
	classifier     ocrpipeline.PageClassifier
}

func (s *pluginSeamSet) close() {
	if s != nil {
		s.mgr.Close(context.Background())
	}
}

func (s *pluginSeamSet) workflowConfig() ocrpipeline.StructuredWorkflowConfig {
	return ocrpipeline.StructuredWorkflowConfig{
		Client:         s.client,
		Segmenter:      s.segmenter,
		Parser:         s.parser,
		PageValidators: s.pageValidators,
		BookValidators: s.bookValidators,
		Classifier:     s.classifier,
	}
}

// setupPluginSeams starts the plugin manager (nil manager when no bindings)
// and composes every seam adapter for the structured workflow.
func setupPluginSeams(ctx context.Context, specs []plugin.Spec, workDir string, dryRun bool, baseClient ocrpipeline.StructuredOCRClient) (*pluginSeamSet, error) {
	mgr, err := plugin.NewManager(ctx, specs, devctlruntime.RequestMeta{RepoRoot: workDir, DryRun: dryRun})
	if err != nil {
		return nil, err
	}
	seams := &pluginSeamSet{mgr: mgr, client: baseClient}
	if mgr.Has(plugin.OpPromptRender) {
		if geppetto, ok := baseClient.(*ocrpipeline.GeppettoStructuredOCRClient); ok {
			geppetto.Prompts = plugin.NewPromptRenderer(mgr)
		}
	}
	if mgr.Has(plugin.OpOCRPage) {
		seams.client = plugin.NewStructuredOCRClient(mgr, baseClient)
	}
	if mgr.Has(plugin.OpFiguresSegment) {
		seams.segmenter = plugin.NewFigureSegmenter(mgr)
	}
	if mgr.Has(plugin.OpResponseParse) {
		seams.parser = plugin.NewResponseParser(mgr)
	}
	if mgr.Has(plugin.OpValidatePage) {
		seams.pageValidators = append(seams.pageValidators, plugin.NewPageValidator(mgr))
	}
	if mgr.Has(plugin.OpValidateBook) {
		seams.bookValidators = append(seams.bookValidators, plugin.NewBookValidator(mgr))
	}
	if mgr.Has(plugin.OpPageClassify) {
		seams.classifier = plugin.NewPageClassifier(mgr)
	}
	for _, prov := range mgr.Provenance() {
		fmt.Printf("plugin %s (%s) ops=%v\n", prov.ID, prov.Name, prov.Ops)
	}
	return seams, nil
}

func main() {
	args := os.Args[1:]
	// Backwards-compatible shorthand: bare flags mean the legacy run command.
	if len(args) > 0 && strings.HasPrefix(args[0], "-") && args[0] != "-h" && args[0] != "--help" {
		args = append([]string{"run"}, args...)
	}
	root, err := newRootCommand()
	if err == nil {
		root.SetArgs(args)
		err = root.Execute()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}


// requireProfileForLiveRun guards the live-by-default structured commands: a
// live run without an explicit profile would silently resolve to whatever the
// local Pinocchio default is, which is never what an operator wants for a
// full-book run.
func requireProfileForLiveRun(dryRun bool, profile string) error {
	if dryRun || strings.TrimSpace(profile) != "" {
		return nil
	}
	return fmt.Errorf("live structured OCR requires --profile (or pass --dry-run for the deterministic offline mode)")
}

func runWorkflow(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	var registries registryFlags
	bookID := fs.String("book-id", "", "Book identifier for workflow metadata and output names")
	imageDir := fs.String("image-dir", "", "Directory containing page images")
	pageGlob := fs.String("page-glob", "page_*.png", "Glob used inside image-dir to discover page images")
	startPage := fs.Int("start-page", 0, "Optional first page number to process")
	endPage := fs.Int("end-page", 0, "Optional last page number to process")
	workDir := fs.String("work-dir", defaultWorkDir, "Directory for engine DB, artifacts, and projections")
	profile := fs.String("profile", "", "Optional Pinocchio profile slug; empty uses Pinocchio defaults")
	promptVersion := fs.String("prompt-version", ocrmvp.DefaultPromptVersion, "OCR prompt version to use")
	contextWindow := fs.Int("context-window", 0, "Optional number of previous/next page images to include as OCR continuity context (0-2)")
	logLevel := fs.String("log-level", "info", "zerolog level: trace, debug, info, warn, error, disabled")
	dryRun := fs.Bool("dry-run", false, "Use deterministic dry-run OCR instead of live Geppetto inference")
	maxWorkers := fs.Int("max-workers", 4, "Maximum concurrent workflow workers")
	pollInterval := fs.Duration("poll-interval", 250*time.Millisecond, "Worker polling interval")
	runID := fs.String("run-id", "", "Optional stable workflow run ID")
	fs.Var(&registries, "profile-registries", "Pinocchio profile registry source (repeatable or comma-separated)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := configureLogLevel(*logLevel); err != nil {
		return err
	}

	if strings.TrimSpace(*bookID) == "" {
		return fmt.Errorf("--book-id is required")
	}
	if strings.TrimSpace(*imageDir) == "" {
		return fmt.Errorf("--image-dir is required")
	}
	absImageDir, err := filepath.Abs(*imageDir)
	if err != nil {
		return err
	}
	paths, err := resolveWorkDir(*workDir, true)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := ocrmvp.OCRClient(ocrmvp.NewGeppettoOCRClient())
	if *dryRun {
		client = ocrmvp.DryRunOCRClient{}
	}
	rt, err := newRuntime(ctx, paths, *maxWorkers, *pollInterval)
	if err != nil {
		return err
	}
	defer func() { _ = rt.Close() }()
	if err := ocrmvp.Register(rt, ocrmvp.Config{Client: client}); err != nil {
		return err
	}

	runOpts := []workflow.RunOption{workflow.WithRunName("OCR MVP: " + *bookID)}
	if strings.TrimSpace(*runID) != "" {
		runOpts = append(runOpts, workflow.WithRunID(*runID))
	}
	handle, err := rt.StartRun(ctx, ocrmvp.PackageName, ocrmvp.RunInput{
		BookID:            *bookID,
		ImageDir:          absImageDir,
		PageGlob:          *pageGlob,
		StartPage:         *startPage,
		EndPage:           *endPage,
		Profile:           *profile,
		ProfileRegistries: append([]string(nil), registries...),
		PromptVersion:     *promptVersion,
		ContextWindow:     *contextWindow,
		DryRun:            *dryRun,
	}, runOpts...)
	if err != nil {
		return err
	}
	fmt.Printf("started run %s in %s\n", handle.ID, paths.root)

	for {
		cycle, err := rt.RunOnce(ctx)
		if err != nil {
			return err
		}
		wf, err := rt.Workflow(ctx, handle.ID)
		if err != nil {
			return err
		}
		fmt.Printf("status=%s processed=%d succeeded=%d failed=%d retried=%d\n", wf.Status, cycle.Processed, cycle.Succeeded, cycle.Failed, cycle.Retried)
		if isTerminal(wf.Status) {
			if wf.Status != model.WorkflowStatusSucceeded {
				return fmt.Errorf("workflow finished with status %s", wf.Status)
			}
			result, err := rt.Result(ctx, handle.ID, "assemble-markdown")
			if err == nil && result != nil {
				fmt.Printf("assemble result: %s\n", string(result.Data))
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(*pollInterval):
		}
	}
}

func runStructuredPage(args []string) error {
	fs := flag.NewFlagSet("structured-page", flag.ContinueOnError)
	var registries registryFlags
	bookID := fs.String("book-id", "", "Book identifier for structured OCR metadata")
	pageNumber := fs.Int("page", 0, "Target page number")
	imagePath := fs.String("image", "", "Target page image path")
	workDir := fs.String("work-dir", ".structured-ocr", "Directory for structured OCR artifacts and turns DB")
	turnsDB := fs.String("turns-db", "", "Optional explicit SQLite turns DB path; defaults to <work-dir>/turns.db")
	turnsDSN := fs.String("turns-dsn", "", "Optional explicit turns DB DSN")
	runID := fs.String("run-id", "structured-page", "Run identifier used in turn conv_id")
	profile := fs.String("profile", "", "Optional Pinocchio profile slug for live structured OCR")
	logLevel := fs.String("log-level", "info", "zerolog level: trace, debug, info, warn, error, disabled")
	dryRun := fs.Bool("dry-run", false, "Use deterministic dry-run structured OCR instead of live inference")
	fs.Var(&registries, "profile-registries", "Pinocchio profile registry source (repeatable or comma-separated)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := configureLogLevel(*logLevel); err != nil {
		return err
	}
	if err := requireProfileForLiveRun(*dryRun, *profile); err != nil {
		return err
	}
	if strings.TrimSpace(*bookID) == "" {
		return fmt.Errorf("--book-id is required")
	}
	if *pageNumber <= 0 {
		return fmt.Errorf("--page must be positive")
	}
	if strings.TrimSpace(*imagePath) == "" {
		return fmt.Errorf("--image is required")
	}
	absImage, err := filepath.Abs(*imagePath)
	if err != nil {
		return err
	}
	absWorkDir, err := filepath.Abs(*workDir)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	client := ocrpipeline.StructuredOCRClient(ocrpipeline.NewGeppettoStructuredOCRClient())
	if *dryRun {
		client = ocrpipeline.DryRunStructuredOCRClient{}
	}
	result, err := ocrpipeline.RunStructuredPage(ctx, ocrpipeline.StructuredOCRInput{
		BookID:            *bookID,
		RunID:             *runID,
		PageNumber:        *pageNumber,
		ImagePath:         absImage,
		WorkDir:           absWorkDir,
		TurnsDB:           *turnsDB,
		TurnsDSN:          *turnsDSN,
		Profile:           *profile,
		ProfileRegistries: append([]string(nil), registries...),
		DryRun:            *dryRun,
	}, client)
	if err != nil {
		return err
	}
	fmt.Printf("structured page %03d written to %s\n", result.PageNumber, result.PageDir)
	fmt.Printf("structured json: %s\n", result.StructuredJSON)
	fmt.Printf("rendered markdown: %s\n", result.RenderedMD)
	fmt.Printf("validation: %s\n", result.ValidationJSON)
	return nil
}

type structuredRunReport struct {
	BookID        string                                `json:"book_id"`
	PageCount     int                                   `json:"page_count"`
	AssembledPath string                                `json:"assembled_path"`
	Pages         []ocrpipeline.StructuredPageRunResult `json:"pages"`
	Warnings      []string                              `json:"warnings,omitempty"`
}

func runStructuredRun(args []string) error {
	fs := flag.NewFlagSet("structured-run", flag.ContinueOnError)
	var registries registryFlags
	bookID := fs.String("book-id", "", "Book identifier for structured OCR metadata")
	imageDir := fs.String("image-dir", "", "Directory containing page images")
	pageGlob := fs.String("page-glob", "page_*.png", "Glob used inside image-dir to discover page images")
	startPage := fs.Int("start-page", 0, "Optional first page number to process")
	endPage := fs.Int("end-page", 0, "Optional last page number to process")
	workDir := fs.String("work-dir", ".structured-ocr-run", "Directory for structured OCR workflow DB, artifacts, and outputs")
	runID := fs.String("run-id", "", "Optional stable workflow run ID")
	profile := fs.String("profile", "", "Optional Pinocchio profile slug for live structured OCR")
	logLevel := fs.String("log-level", "info", "zerolog level: trace, debug, info, warn, error, disabled")
	dryRun := fs.Bool("dry-run", false, "Use deterministic dry-run structured OCR instead of live inference")
	maxWorkers := fs.Int("max-workers", 2, "Maximum concurrent workflow workers")
	pollInterval := fs.Duration("poll-interval", 250*time.Millisecond, "Worker polling interval")
	expectedPages := fs.Int("expected-pages", 0, "Expected page marker count for validation; 0 disables exact check")
	minRenderedBytes := fs.Int("min-rendered-bytes", 0, "Warn in validation report for successful pages rendered shorter than this byte count; 0 disables")
	embedFigures := fs.Bool("embed-figures", false, "Extract and embed figure images from structured figure markers during assembly")
	figuresDir := fs.String("figures-dir", "", "Optional output directory for extracted figures; defaults to <work-dir>/figures when --embed-figures is set")
	renderPDF := fs.Bool("render-pdf", false, "Render assembled Markdown to PDF with pandoc during assembly")
	pdfPath := fs.String("pdf-path", "", "Optional PDF output path; defaults to <work-dir>/book.pdf when --render-pdf is set")
	pandocPath := fs.String("pandoc-path", "", "Optional pandoc executable path; defaults to pandoc")
	bookProfile := fs.String("book-profile", "", "Optional book profile YAML; drives prompt lexicon, code language, render policy, and plugin bindings")
	fs.Var(&registries, "profile-registries", "Pinocchio profile registry source (repeatable or comma-separated)")
	var pluginBindings pluginBindingFlags
	fs.Var(&pluginBindings, "plugin", "Bind an NDJSON-stdio plugin to a seam as seam=path (repeatable; seams: ocr.page, prompt.render, figures.segment)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := configureLogLevel(*logLevel); err != nil {
		return err
	}
	if strings.TrimSpace(*bookID) == "" {
		return fmt.Errorf("--book-id is required")
	}
	if strings.TrimSpace(*imageDir) == "" {
		return fmt.Errorf("--image-dir is required")
	}
	if strings.TrimSpace(*bookProfile) != "" {
		abs, err := filepath.Abs(*bookProfile)
		if err != nil {
			return err
		}
		*bookProfile = abs
	}
	pluginSpecs, hasOCRPagePlugin, err := buildPluginSpecs(pluginBindings)
	if err != nil {
		return err
	}
	profilePluginSpecs, profileHasOCRPage, err := pluginSpecsFromProfile(*bookProfile)
	if err != nil {
		return err
	}
	// CLI --plugin bindings take precedence over profile bindings
	// (first-wins per seam in the manager).
	pluginSpecs = append(pluginSpecs, profilePluginSpecs...)
	hasOCRPagePlugin = hasOCRPagePlugin || profileHasOCRPage
	if !hasOCRPagePlugin {
		if err := requireProfileForLiveRun(*dryRun, *profile); err != nil {
			return err
		}
	}
	absImageDir, err := filepath.Abs(*imageDir)
	if err != nil {
		return err
	}
	paths, err := resolveWorkDir(*workDir, true)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	client := ocrpipeline.StructuredOCRClient(ocrpipeline.NewGeppettoStructuredOCRClient())
	if *dryRun {
		client = ocrpipeline.DryRunStructuredOCRClient{}
	}
	seams, err := setupPluginSeams(ctx, pluginSpecs, paths.root, *dryRun, client)
	if err != nil {
		return err
	}
	defer seams.close()
	rt, err := newRuntime(ctx, paths, *maxWorkers, *pollInterval)
	if err != nil {
		return err
	}
	defer func() { _ = rt.Close() }()
	if err := ocrpipeline.RegisterStructuredWorkflow(rt, seams.workflowConfig()); err != nil {
		return err
	}
	runOpts := []workflow.RunOption{workflow.WithRunName("Structured OCR: " + *bookID)}
	if strings.TrimSpace(*runID) != "" {
		runOpts = append(runOpts, workflow.WithRunID(*runID))
	}
	handle, err := rt.StartRun(ctx, ocrpipeline.StructuredPackageName, ocrpipeline.StructuredRunInput{BookID: *bookID, ImageDir: absImageDir, PageGlob: *pageGlob, StartPage: *startPage, EndPage: *endPage, WorkDir: paths.root, RunID: string(*runID), Profile: *profile, ProfileRegistries: append([]string(nil), registries...), DryRun: *dryRun, ExpectedPages: *expectedPages, MinRenderedBytes: *minRenderedBytes, EmbedFigures: *embedFigures, FiguresDir: *figuresDir, RenderPDF: *renderPDF, PDFPath: *pdfPath, PandocPath: *pandocPath, BookProfile: *bookProfile}, runOpts...)
	if err != nil {
		return err
	}
	mode := "LIVE (model calls will be billed)"
	if *dryRun {
		mode = "dry-run (deterministic fake OCR, no model calls)"
	} else if hasOCRPagePlugin {
		mode = fmt.Sprintf("plugin (ocr.page via %s)", seams.mgr.PluginIDFor(plugin.OpOCRPage))
	}
	fmt.Printf("mode=%s profile=%q pages=%d-%d max-workers=%d\n", mode, *profile, *startPage, *endPage, *maxWorkers)
	fmt.Printf("started structured run %s in %s\n", handle.ID, paths.root)
	for {
		cycle, err := rt.RunOnce(ctx)
		if err != nil {
			return err
		}
		wf, err := rt.Workflow(ctx, handle.ID)
		if err != nil {
			return err
		}
		fmt.Printf("status=%s processed=%d succeeded=%d failed=%d retried=%d\n", wf.Status, cycle.Processed, cycle.Succeeded, cycle.Failed, cycle.Retried)
		if isTerminal(wf.Status) {
			if wf.Status != model.WorkflowStatusSucceeded {
				return fmt.Errorf("workflow finished with status %s", wf.Status)
			}
			if result, err := rt.Result(ctx, handle.ID, "assemble-structured-markdown"); err == nil && result != nil {
				fmt.Printf("assemble result: %s\n", string(result.Data))
			}
			if result, err := rt.Result(ctx, handle.ID, "validate-structured-run"); err == nil && result != nil {
				fmt.Printf("validation result: %s\n", string(result.Data))
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(*pollInterval):
		}
	}
}

func existingStructuredPageResult(workDir string, pageNumber int) (ocrpipeline.StructuredPageRunResult, bool, error) {
	pageDir := filepath.Join(workDir, "pages", fmt.Sprintf("page_%03d", pageNumber))
	result := ocrpipeline.StructuredPageRunResult{
		PageNumber:     pageNumber,
		PageDir:        pageDir,
		RawResponse:    filepath.Join(pageDir, "03-raw-response.json"),
		StructuredJSON: filepath.Join(pageDir, "04-structured.json"),
		RenderedMD:     filepath.Join(pageDir, "05-rendered.md"),
		ValidationJSON: filepath.Join(pageDir, "06-validation.json"),
	}
	for _, path := range []string{result.StructuredJSON, result.RenderedMD, result.ValidationJSON} {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return ocrpipeline.StructuredPageRunResult{}, false, nil
			}
			return ocrpipeline.StructuredPageRunResult{}, false, err
		}
	}
	validationBytes, err := os.ReadFile(result.ValidationJSON)
	if err != nil {
		return ocrpipeline.StructuredPageRunResult{}, false, err
	}
	_ = json.Unmarshal(validationBytes, &result.Validation)
	return result, true, nil
}

func runQualityPass(args []string) error {
	fs := flag.NewFlagSet("quality-pass", flag.ContinueOnError)
	markdownPath := fs.String("markdown", "", "OCR markdown file to QA and normalize")
	outputDir := fs.String("output-dir", "", "Directory for normalized markdown and diff")
	workDir := fs.String("work-dir", defaultWorkDir, "Directory for engine DB, artifacts, and projections")
	bookID := fs.String("book-id", "", "Optional book identifier for report metadata")
	bookProfile := fs.String("book-profile", "", "Optional YAML book profile; workflow writes discoveries separately and never mutates this file")
	discoveryPath := fs.String("discovery", "", "Optional output path for machine-updated book discovery YAML")
	profilePatchPath := fs.String("profile-patch", "", "Optional output path for proposed profile patch YAML")
	expectedPages := fs.Int("expected-pages", 0, "Expected page marker count; 0 uses book profile/default")
	logPath := fs.String("log", "", "Optional OCR run log to import into SQLite")
	imageDir := fs.String("image-dir", "", "Optional page image directory for embedded figure extraction")
	embedFigures := fs.Bool("embed-figures", false, "Extract figure images from source pages and embed markdown image links")
	maxWorkers := fs.Int("max-workers", 2, "Maximum concurrent workflow workers")
	pollInterval := fs.Duration("poll-interval", 250*time.Millisecond, "Worker polling interval")
	runID := fs.String("run-id", "", "Optional stable workflow run ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*markdownPath) == "" {
		return fmt.Errorf("--markdown is required")
	}
	absMarkdown, err := filepath.Abs(*markdownPath)
	if err != nil {
		return err
	}
	outDir := strings.TrimSpace(*outputDir)
	if outDir == "" {
		outDir = filepath.Join(filepath.Dir(absMarkdown), "quality-pass")
	}
	absOutputDir, err := filepath.Abs(outDir)
	if err != nil {
		return err
	}
	paths, err := resolveWorkDir(*workDir, true)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	rt, err := newRuntime(ctx, paths, *maxWorkers, *pollInterval)
	if err != nil {
		return err
	}
	defer func() { _ = rt.Close() }()
	if err := ocrquality.Register(rt); err != nil {
		return err
	}
	runOpts := []workflow.RunOption{workflow.WithRunName("OCR Quality Pass")}
	if strings.TrimSpace(*runID) != "" {
		runOpts = append(runOpts, workflow.WithRunID(*runID))
	}
	handle, err := rt.StartRun(ctx, ocrquality.PackageName, ocrquality.RunInput{BookID: *bookID, BookProfilePath: *bookProfile, DiscoveryPath: *discoveryPath, ProfilePatchPath: *profilePatchPath, MarkdownPath: absMarkdown, OutputDir: absOutputDir, ExpectedPages: *expectedPages, LogPath: *logPath, ImageDir: *imageDir, EmbedFigures: *embedFigures}, runOpts...)
	if err != nil {
		return err
	}
	fmt.Printf("started quality run %s in %s\n", handle.ID, paths.root)
	for {
		cycle, err := rt.RunOnce(ctx)
		if err != nil {
			return err
		}
		wf, err := rt.Workflow(ctx, handle.ID)
		if err != nil {
			return err
		}
		fmt.Printf("status=%s processed=%d succeeded=%d failed=%d retried=%d\n", wf.Status, cycle.Processed, cycle.Succeeded, cycle.Failed, cycle.Retried)
		if isTerminal(wf.Status) {
			if wf.Status != model.WorkflowStatusSucceeded {
				return fmt.Errorf("workflow finished with status %s", wf.Status)
			}
			fmt.Printf("quality output: %s\n", filepath.Join(absOutputDir, "normalized.md"))
			if *embedFigures {
				fmt.Printf("quality embedded output: %s\n", filepath.Join(absOutputDir, "embedded-figures.md"))
			}
			fmt.Printf("quality diff: %s\n", filepath.Join(absOutputDir, "cleanup.diff"))
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(*pollInterval):
		}
	}
}

func structuredRerunPages(args []string) error {
	fs := flag.NewFlagSet("structured-rerun-pages", flag.ContinueOnError)
	workDir := fs.String("work-dir", defaultWorkDir, "Directory containing engine.db")
	runID := fs.String("run-id", "", "Workflow run ID")
	pagesArg := fs.String("pages", "", "Comma-separated page numbers to rerun")
	renderPDF := fs.Bool("render-pdf", false, "Render PDF during reassembly")
	pdfPath := fs.String("pdf-path", "", "Optional PDF output path; defaults to <work-dir>/book.pdf")
	pandocPath := fs.String("pandoc-path", "", "Optional pandoc executable path")
	logLevel := fs.String("log-level", "info", "zerolog level: trace, debug, info, warn, error, disabled")
	maxWorkers := fs.Int("max-workers", 2, "Maximum concurrent workflow workers")
	pollInterval := fs.Duration("poll-interval", 250*time.Millisecond, "Worker polling interval")
	var pluginBindings pluginBindingFlags
	fs.Var(&pluginBindings, "plugin", "Bind an NDJSON-stdio plugin to a seam as seam=path (repeatable; seams: ocr.page, prompt.render, figures.segment)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := configureLogLevel(*logLevel); err != nil {
		return err
	}
	if strings.TrimSpace(*runID) == "" {
		return fmt.Errorf("--run-id is required")
	}
	pages, err := parsePageList(*pagesArg)
	if err != nil {
		return err
	}
	pluginSpecs, _, err := buildPluginSpecs(pluginBindings)
	if err != nil {
		return err
	}
	paths, err := resolveWorkDir(*workDir, false)
	if err != nil {
		return err
	}
	if err := requeueStructuredPages(paths, model.WorkflowID(*runID), pages, *renderPDF, *pdfPath, *pandocPath); err != nil {
		return err
	}
	fmt.Printf("requeued structured pages %v in run %s\n", pages, *runID)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	structuredClient := ocrpipeline.StructuredOCRClient(ocrpipeline.NewGeppettoStructuredOCRClient())
	seams, err := setupPluginSeams(ctx, pluginSpecs, paths.root, false, structuredClient)
	if err != nil {
		return err
	}
	defer seams.close()
	rt, err := newRuntime(ctx, paths, *maxWorkers, *pollInterval)
	if err != nil {
		return err
	}
	defer func() { _ = rt.Close() }()
	if err := ocrpipeline.RegisterStructuredWorkflow(rt, seams.workflowConfig()); err != nil {
		return err
	}
	for {
		cycle, err := rt.RunOnce(ctx)
		if err != nil {
			return err
		}
		wf, err := rt.Workflow(ctx, model.WorkflowID(*runID))
		if err != nil {
			return err
		}
		fmt.Printf("status=%s processed=%d succeeded=%d failed=%d retried=%d\n", wf.Status, cycle.Processed, cycle.Succeeded, cycle.Failed, cycle.Retried)
		if isTerminal(wf.Status) {
			if wf.Status != model.WorkflowStatusSucceeded {
				return fmt.Errorf("workflow finished with status %s", wf.Status)
			}
			if result, err := rt.Result(ctx, model.WorkflowID(*runID), "assemble-structured-markdown"); err == nil && result != nil {
				fmt.Printf("assemble result: %s\n", string(result.Data))
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(*pollInterval):
		}
	}
}

func parsePageList(value string) ([]int, error) {
	parts := strings.Split(value, ",")
	seen := map[int]bool{}
	var pages []int
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		var page int
		if _, err := fmt.Sscanf(part, "%d", &page); err != nil || page <= 0 {
			return nil, fmt.Errorf("invalid page number %q", part)
		}
		if !seen[page] {
			seen[page] = true
			pages = append(pages, page)
		}
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("--pages is required")
	}
	sort.Ints(pages)
	return pages, nil
}

// knownEngineMigrations pins the engine schema requeueStructuredPages was
// written against (scraper v0.0.4). The direct SQL below re-implements
// scheduler invariants by hand; if a scraper upgrade adds a migration we have
// not reviewed, refuse to mutate the run instead of corrupting it. Delete this
// guard (and the SQL) once the runtime grows a first-class RequeueSteps API —
// see ttmp WORKFLOW-RUNTIME-HARDENING-001.
var knownEngineMigrations = []string{"001_engine_core.sql", "002_engine_runtime.sql"}

func guardEngineSchema(db *sql.DB) error {
	rows, err := db.Query(`SELECT name FROM schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("reading engine schema_migrations: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(names) != len(knownEngineMigrations) {
		return fmt.Errorf("engine.db schema %v differs from the schema structured-rerun-pages was written against %v; refusing direct SQL requeue (update book-ocr or use a matching scraper version)", names, knownEngineMigrations)
	}
	for i, name := range names {
		if name != knownEngineMigrations[i] {
			return fmt.Errorf("engine.db schema %v differs from the schema structured-rerun-pages was written against %v; refusing direct SQL requeue (update book-ocr or use a matching scraper version)", names, knownEngineMigrations)
		}
	}
	return nil
}

func requeueStructuredPages(paths workPaths, runID model.WorkflowID, pages []int, renderPDF bool, pdfPath string, pandocPath string) error {
	db, err := sql.Open("sqlite3", paths.engineDB)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	if err := guardEngineSchema(db); err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	rollback := true
	defer func() {
		if rollback {
			_ = tx.Rollback()
		}
	}()
	stamp := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.Exec(`UPDATE workflows SET status = ?, updated_at = ? WHERE id = ?`, model.WorkflowStatusRunning, stamp, runID); err != nil {
		return err
	}
	for _, page := range pages {
		stepID := fmt.Sprintf("structured-page-%03d", page)
		if _, err := tx.Exec(`DELETE FROM leases WHERE op_id = ?`, stepID); err != nil {
			return err
		}
		res, err := tx.Exec(`UPDATE ops SET status = ?, retry_state_json = ?, next_attempt_at = NULL, updated_at = ? WHERE workflow_id = ? AND id = ?`, model.OpStatusReady, `{"Attempt":0,"NextAttemptAt":null,"LastError":""}`, stamp, runID, stepID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return fmt.Errorf("structured page op %s not found", stepID)
		}
	}
	for _, stepID := range []string{"assemble-structured-markdown", "validate-structured-run"} {
		if _, err := tx.Exec(`DELETE FROM leases WHERE op_id = ?`, stepID); err != nil {
			return err
		}
		// Downstream ops must wait for the requeued page ops to finish. Mark them
		// pending instead of ready; the scheduler will make them ready once their
		// dependencies are succeeded again.
		if _, err := tx.Exec(`UPDATE ops SET status = ?, retry_state_json = ?, next_attempt_at = NULL, updated_at = ? WHERE workflow_id = ? AND id = ?`, model.OpStatusPending, `{"Attempt":0,"NextAttemptAt":null,"LastError":""}`, stamp, runID, stepID); err != nil {
			return err
		}
	}
	if renderPDF {
		var inputJSON string
		if err := tx.QueryRow(`SELECT input_json FROM ops WHERE workflow_id = ? AND id = ?`, runID, "assemble-structured-markdown").Scan(&inputJSON); err != nil {
			return err
		}
		var input map[string]any
		if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
			return err
		}
		input["render_pdf"] = true
		if strings.TrimSpace(pdfPath) != "" {
			input["pdf_path"] = pdfPath
		} else if _, ok := input["pdf_path"]; !ok {
			input["pdf_path"] = filepath.Join(paths.root, "book.pdf")
		}
		if strings.TrimSpace(pandocPath) != "" {
			input["pandoc_path"] = pandocPath
		}
		updated, err := json.Marshal(input)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE ops SET input_json = ? WHERE workflow_id = ? AND id = ?`, string(updated), runID, "assemble-structured-markdown"); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	rollback = false
	return nil
}

func retryStep(args []string) error {
	fs := flag.NewFlagSet("retry", flag.ContinueOnError)
	workDir := fs.String("work-dir", defaultWorkDir, "Directory containing engine.db")
	runID := fs.String("run-id", "", "Workflow run ID")
	stepID := fs.String("step-id", "", "Failed step/op ID to retry")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*runID) == "" {
		return fmt.Errorf("--run-id is required")
	}
	if strings.TrimSpace(*stepID) == "" {
		return fmt.Errorf("--step-id is required")
	}
	ctx := context.Background()
	rt, err := openExistingRuntime(ctx, *workDir)
	if err != nil {
		return err
	}
	defer func() { _ = rt.Close() }()
	if err := rt.RetryStep(ctx, model.WorkflowID(*runID), model.OpID(*stepID)); err != nil {
		return err
	}
	fmt.Printf("retried step %s in run %s\n", *stepID, *runID)
	return nil
}

func resumeRun(args []string) error {
	fs := flag.NewFlagSet("resume", flag.ContinueOnError)
	workDir := fs.String("work-dir", defaultWorkDir, "Directory containing engine.db")
	runID := fs.String("run-id", "", "Workflow run ID")
	logLevel := fs.String("log-level", "info", "zerolog level: trace, debug, info, warn, error, disabled")
	maxWorkers := fs.Int("max-workers", 2, "Maximum concurrent workflow workers")
	pollInterval := fs.Duration("poll-interval", 250*time.Millisecond, "Worker polling interval")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := configureLogLevel(*logLevel); err != nil {
		return err
	}
	if strings.TrimSpace(*runID) == "" {
		return fmt.Errorf("--run-id is required")
	}
	paths, err := resolveWorkDir(*workDir, false)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	client := ocrmvp.OCRClient(ocrmvp.NewGeppettoOCRClient())
	structuredClient := ocrpipeline.StructuredOCRClient(ocrpipeline.NewGeppettoStructuredOCRClient())
	rt, err := newRuntime(ctx, paths, *maxWorkers, *pollInterval)
	if err != nil {
		return err
	}
	defer func() { _ = rt.Close() }()
	if err := ocrmvp.Register(rt, ocrmvp.Config{Client: client}); err != nil {
		return err
	}
	if err := ocrpipeline.RegisterStructuredWorkflow(rt, ocrpipeline.StructuredWorkflowConfig{Client: structuredClient}); err != nil {
		return err
	}
	fmt.Printf("resuming run %s in %s\n", *runID, paths.root)
	for {
		cycle, err := rt.RunOnce(ctx)
		if err != nil {
			return err
		}
		wf, err := rt.Workflow(ctx, model.WorkflowID(*runID))
		if err != nil {
			return err
		}
		fmt.Printf("status=%s processed=%d succeeded=%d failed=%d retried=%d\n", wf.Status, cycle.Processed, cycle.Succeeded, cycle.Failed, cycle.Retried)
		if isTerminal(wf.Status) {
			if wf.Status != model.WorkflowStatusSucceeded {
				return fmt.Errorf("workflow finished with status %s", wf.Status)
			}
			result, err := rt.Result(ctx, model.WorkflowID(*runID), "assemble-markdown")
			if err == nil && result != nil {
				fmt.Printf("assemble result: %s\n", string(result.Data))
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(*pollInterval):
		}
	}
}

func cancelRun(args []string) error {
	fs := flag.NewFlagSet("cancel", flag.ContinueOnError)
	workDir := fs.String("work-dir", defaultWorkDir, "Directory containing engine.db")
	runID := fs.String("run-id", "", "Workflow run ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*runID) == "" {
		return fmt.Errorf("--run-id is required")
	}
	ctx := context.Background()
	rt, err := openExistingRuntime(ctx, *workDir)
	if err != nil {
		return err
	}
	defer func() { _ = rt.Close() }()
	if err := rt.CancelRun(ctx, model.WorkflowID(*runID)); err != nil {
		return err
	}
	fmt.Printf("canceled run %s\n", *runID)
	return nil
}

func showStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	workDir := fs.String("work-dir", defaultWorkDir, "Directory containing engine.db")
	runID := fs.String("run-id", "", "Workflow run ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*runID) == "" {
		return fmt.Errorf("--run-id is required")
	}
	ctx := context.Background()
	rt, err := openExistingRuntime(ctx, *workDir)
	if err != nil {
		return err
	}
	defer func() { _ = rt.Close() }()
	wf, err := rt.Workflow(ctx, model.WorkflowID(*runID))
	if err != nil {
		return err
	}
	fmt.Printf("run_id=%s\nsite=%s\nname=%s\nstatus=%s\ncreated_at=%s\nupdated_at=%s\n", wf.ID, wf.Site, wf.Name, wf.Status, wf.CreatedAt.Format(time.RFC3339), wf.UpdatedAt.Format(time.RFC3339))
	return nil
}

func listPages(args []string) error {
	fs := flag.NewFlagSet("pages", flag.ContinueOnError)
	workDir := fs.String("work-dir", defaultWorkDir, "Directory containing projections")
	bookID := fs.String("book-id", "", "Book ID to list")
	status := fs.String("status", "", "Optional page status filter")
	limit := fs.Int("limit", 0, "Optional maximum number of rows")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*bookID) == "" {
		return fmt.Errorf("--book-id is required")
	}
	paths, err := resolveWorkDir(*workDir, false)
	if err != nil {
		return err
	}
	ctx := context.Background()
	store := workflow.NewSQLiteProjectionStore(paths.projectionsDir)
	defer func() { _ = store.Close() }()
	projection, err := store.Projection(ctx, ocrmvp.PackageName)
	if err != nil {
		return err
	}
	if err := ocrmvp.EnsureProjectionSchema(ctx, projection); err != nil {
		return err
	}
	query := `SELECT page_num, status, ocr_step_id, char_count, error_code, markdown_artifact_uri, updated_at FROM pages WHERE book_id = ?`
	queryArgs := []any{*bookID}
	if strings.TrimSpace(*status) != "" {
		query += ` AND status = ?`
		queryArgs = append(queryArgs, *status)
	}
	query += ` ORDER BY page_num`
	if *limit > 0 {
		query += ` LIMIT ?`
		queryArgs = append(queryArgs, *limit)
	}
	rows, err := projection.Query(ctx, query, queryArgs...)
	if err != nil {
		return err
	}
	printPageRows(rows)
	return nil
}

type workPaths struct {
	root           string
	engineDB       string
	artifactsDir   string
	projectionsDir string
}

func resolveWorkDir(workDir string, create bool) (workPaths, error) {
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return workPaths{}, err
	}
	if create {
		if err := os.MkdirAll(absWorkDir, 0o755); err != nil {
			return workPaths{}, err
		}
	}
	return workPaths{
		root:           absWorkDir,
		engineDB:       filepath.Join(absWorkDir, "engine.db"),
		artifactsDir:   filepath.Join(absWorkDir, "artifacts"),
		projectionsDir: filepath.Join(absWorkDir, "projections"),
	}, nil
}

func openExistingRuntime(ctx context.Context, workDir string) (*workflow.Runtime, error) {
	paths, err := resolveWorkDir(workDir, false)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(paths.engineDB); err != nil {
		return nil, fmt.Errorf("open engine DB %s: %w", paths.engineDB, err)
	}
	return newRuntime(ctx, paths, 1, 250*time.Millisecond)
}

func newRuntime(ctx context.Context, paths workPaths, maxWorkers int, pollInterval time.Duration) (*workflow.Runtime, error) {
	return workflow.NewRuntime(ctx, workflow.Config{
		Store:           workflow.SQLiteStore(paths.engineDB),
		ArtifactStore:   workflow.NewFileArtifactStore(paths.artifactsDir),
		ProjectionStore: workflow.NewSQLiteProjectionStore(paths.projectionsDir),
		WorkerID:        "ocr-mvp-cli",
		MaxWorkers:      maxWorkers,
		PollInterval:    pollInterval,
		Queues: map[model.QueueKey]workflow.QueueConfig{
			ocrmvp.QueueControl:                   {MaxWorkers: 1},
			ocrmvp.QueueOCR:                       {MaxWorkers: maxWorkers},
			ocrmvp.QueueAssemble:                  {MaxWorkers: 1},
			ocrquality.QueueQuality:               {MaxWorkers: maxWorkers},
			ocrpipeline.QueueStructuredControl:    {MaxWorkers: 1},
			ocrpipeline.QueueStructuredVision:     {MaxWorkers: maxWorkers},
			ocrpipeline.QueueStructuredAssemble:   {MaxWorkers: 1},
			ocrpipeline.QueueStructuredValidation: {MaxWorkers: 1},
		},
	})
}

func configureLogLevel(value string) error {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		value = "info"
	}
	if value == "disabled" || value == "off" || value == "none" {
		zerolog.SetGlobalLevel(zerolog.Disabled)
		return nil
	}
	level, err := zerolog.ParseLevel(value)
	if err != nil {
		return fmt.Errorf("invalid --log-level %q: %w", value, err)
	}
	zerolog.SetGlobalLevel(level)
	return nil
}

func listStructuredPages(args []string) error {
	fs := flag.NewFlagSet("structured-pages", flag.ContinueOnError)
	workDir := fs.String("work-dir", defaultWorkDir, "Directory containing projections")
	bookID := fs.String("book-id", "", "Book ID to list")
	status := fs.String("status", "", "Optional page status filter")
	limit := fs.Int("limit", 0, "Optional maximum number of rows")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*bookID) == "" {
		return fmt.Errorf("--book-id is required")
	}
	paths, err := resolveWorkDir(*workDir, false)
	if err != nil {
		return err
	}
	ctx := context.Background()
	store := workflow.NewSQLiteProjectionStore(paths.projectionsDir)
	defer func() { _ = store.Close() }()
	projection, err := store.Projection(ctx, ocrpipeline.StructuredProjectionName)
	if err != nil {
		return err
	}
	if err := ocrpipeline.EnsureStructuredProjectionSchema(ctx, projection); err != nil {
		return err
	}
	query := `SELECT page_num, status, step_id, warning_count, table_count, figure_count, rendered_bytes, error_code, rendered_markdown_path, updated_at FROM structured_pages WHERE book_id = ?`
	queryArgs := []any{*bookID}
	if strings.TrimSpace(*status) != "" {
		query += ` AND status = ?`
		queryArgs = append(queryArgs, *status)
	}
	query += ` ORDER BY page_num`
	if *limit > 0 {
		query += ` LIMIT ?`
		queryArgs = append(queryArgs, *limit)
	}
	rows, err := projection.Query(ctx, query, queryArgs...)
	if err != nil {
		return err
	}
	printPageRows(rows)
	return nil
}

func printPageRows(rows []map[string]any) {
	if len(rows) == 0 {
		fmt.Println("no pages")
		return
	}
	for _, row := range rows {
		keys := make([]string, 0, len(row))
		for key := range row {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			parts = append(parts, fmt.Sprintf("%s=%v", key, row[key]))
		}
		fmt.Println(strings.Join(parts, " "))
	}
}

func isTerminal(status model.WorkflowStatus) bool {
	switch status {
	case model.WorkflowStatusSucceeded, model.WorkflowStatusFailed, model.WorkflowStatusCanceled:
		return true
	case model.WorkflowStatusPending, model.WorkflowStatusRunning:
		return false
	default:
		return false
	}
}
