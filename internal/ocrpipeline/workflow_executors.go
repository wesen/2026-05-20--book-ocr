package ocrpipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/book-ocr/internal/ocrmvp"
	"github.com/go-go-golems/book-ocr/internal/ocrquality"
	"github.com/go-go-golems/book-ocr/internal/ocrvalidation"
	"github.com/go-go-golems/scraper/pkg/engine/model"
	"github.com/go-go-golems/scraper/pkg/workflow"
)

func StructuredDiscoverExecutor(projectionName string) workflow.Executor {
	return workflow.NewTypedExecutor(KindStructuredDiscover, func(ctx context.Context, step *workflow.StepContext, input StructuredRunInput) error {
		input = normalizeStructuredRunInput(input)
		if err := validateStructuredRunInput(input); err != nil {
			return workflow.Permanent("structured_input_invalid", err)
		}
		projection, err := step.Projection(projectionNameOrDefaultStructured(projectionName))
		if err != nil {
			return workflow.Retryable("structured_projection_unavailable", err)
		}
		if err := EnsureStructuredProjectionSchema(ctx, projection); err != nil {
			return workflow.Retryable("structured_projection_schema_failed", err)
		}
		pages, err := ocrmvp.DiscoverPageImages(ocrmvp.RunInput{BookID: input.BookID, ImageDir: input.ImageDir, PageGlob: input.PageGlob, StartPage: input.StartPage, EndPage: input.EndPage})
		if err != nil {
			return workflow.Permanent("structured_discover_pages_failed", err)
		}
		pageHandles := make([]workflow.StepHandle, 0, len(pages))
		pageStepIDs := make([]string, 0, len(pages))
		pageSpecs := make([]StructuredPageSpec, 0, len(pages))
		for _, page := range pages {
			pageInput := StructuredPageWorkflowInput{BookID: input.BookID, RunID: firstNonEmpty(input.RunID, string(step.Workflow().ID)), PageNumber: page.PageNumber, ImagePath: page.ImagePath, WorkDir: input.WorkDir, TurnsDB: filepath.Join(input.WorkDir, "turns.db"), Profile: input.Profile, ProfileRegistries: append([]string(nil), input.ProfileRegistries...), DryRun: input.DryRun}
			stepID := structuredPageStepID(page.PageNumber)
			if err := seedStructuredPage(ctx, projection, pageInput, stepID); err != nil {
				return workflow.Retryable("structured_projection_page_seed_failed", err)
			}
			emittedID, err := step.Emit(stepID, pageInput, workflow.StepOpts{Kind: KindStructuredPage, Queue: model.QueueKey(QueueStructuredVision), Retry: structuredOCRRetryPolicy(), Metadata: map[string]string{"book_id": input.BookID, "page": fmt.Sprintf("%d", page.PageNumber)}})
			if err != nil {
				return err
			}
			pageHandles = append(pageHandles, workflow.StepHandle{ID: emittedID})
			pageStepIDs = append(pageStepIDs, string(emittedID))
			pageSpecs = append(pageSpecs, StructuredPageSpec{BookID: page.BookID, PageNumber: page.PageNumber, ImagePath: page.ImagePath})
		}
		figuresDir := strings.TrimSpace(input.FiguresDir)
		if figuresDir == "" && input.EmbedFigures {
			figuresDir = filepath.Join(input.WorkDir, "figures")
		}
		assembleID, err := step.Emit("assemble-structured-markdown", StructuredAssembleInput{BookID: input.BookID, WorkDir: input.WorkDir, ImageDir: input.ImageDir, EmbedFigures: input.EmbedFigures, FiguresDir: figuresDir}, workflow.StepOpts{Kind: KindStructuredAssemble, Queue: model.QueueKey(QueueStructuredAssemble), DependsOn: workflow.Require(pageHandles...), Metadata: map[string]string{"book_id": input.BookID}})
		if err != nil {
			return err
		}
		if _, err := step.Emit("validate-structured-run", StructuredValidateInput{BookID: input.BookID, WorkDir: input.WorkDir, ExpectedPages: input.ExpectedPages, MinRenderedBytes: input.MinRenderedBytes}, workflow.StepOpts{Kind: KindStructuredValidate, Queue: model.QueueKey(QueueStructuredValidation), DependsOn: workflow.Require(workflow.StepHandle{ID: assembleID}), Metadata: map[string]string{"book_id": input.BookID}}); err != nil {
			return err
		}
		return step.Result(StructuredDiscoverResult{BookID: input.BookID, PageCount: len(pages), PageStepIDs: pageStepIDs, Pages: pageSpecs})
	})
}

func StructuredPageExecutor(projectionName string, client StructuredOCRClient) workflow.Executor {
	if client == nil {
		client = DryRunStructuredOCRClient{}
	}
	return workflow.NewTypedExecutor(KindStructuredPage, func(ctx context.Context, step *workflow.StepContext, input StructuredPageWorkflowInput) error {
		projection, err := step.Projection(projectionNameOrDefaultStructured(projectionName))
		if err != nil {
			return workflow.Retryable("structured_projection_unavailable", err)
		}
		if err := markStructuredPageRunning(ctx, projection, input); err != nil {
			return workflow.Retryable("structured_projection_update_failed", err)
		}
		pageResult, err := RunStructuredPage(ctx, StructuredOCRInput{BookID: input.BookID, RunID: input.RunID, PageNumber: input.PageNumber, ImagePath: input.ImagePath, WorkDir: input.WorkDir, TurnsDB: input.TurnsDB, TurnsDSN: input.TurnsDSN, Profile: input.Profile, ProfileRegistries: input.ProfileRegistries, DryRun: input.DryRun}, client)
		if err != nil {
			code := classifyStructuredPageErrorCode(err)
			_ = markStructuredPageError(ctx, projection, input, code, err)
			return classifyStructuredPageError(err)
		}
		structuredBytes, err := os.ReadFile(pageResult.StructuredJSON)
		if err != nil {
			return workflow.Retryable("structured_artifact_read_failed", err)
		}
		renderedBytes, err := os.ReadFile(pageResult.RenderedMD)
		if err != nil {
			return workflow.Retryable("structured_artifact_read_failed", err)
		}
		validationBytes, err := os.ReadFile(pageResult.ValidationJSON)
		if err != nil {
			return workflow.Retryable("structured_artifact_read_failed", err)
		}
		structuredRef, err := step.StoreArtifact(fmt.Sprintf("page_%03d_structured.json", input.PageNumber), "application/json", structuredBytes, workflow.ArtifactKind("structured-ocr-json"))
		if err != nil {
			return workflow.Retryable("structured_store_artifact_failed", err)
		}
		renderedRef, err := step.StoreArtifact(fmt.Sprintf("page_%03d_rendered.md", input.PageNumber), "text/markdown", renderedBytes, workflow.ArtifactKind("structured-rendered-markdown"))
		if err != nil {
			return workflow.Retryable("structured_store_artifact_failed", err)
		}
		validationRef, err := step.StoreArtifact(fmt.Sprintf("page_%03d_validation.json", input.PageNumber), "application/json", validationBytes, workflow.ArtifactKind("structured-validation"))
		if err != nil {
			return workflow.Retryable("structured_store_artifact_failed", err)
		}
		var page StructuredPageOCR
		_ = json.Unmarshal(structuredBytes, &page)
		tableCount, figureCount := countStructuredBlocks(page)
		result := StructuredPageWorkflowResult{BookID: input.BookID, PageNumber: input.PageNumber, PageDir: pageResult.PageDir, RawResponsePath: pageResult.RawResponse, StructuredJSONPath: pageResult.StructuredJSON, RenderedMarkdownPath: pageResult.RenderedMD, ValidationJSONPath: pageResult.ValidationJSON, StructuredArtifactID: structuredRef.ID, StructuredArtifactURI: structuredRef.URI, RenderedArtifactID: renderedRef.ID, RenderedArtifactURI: renderedRef.URI, ValidationArtifactID: validationRef.ID, ValidationArtifactURI: validationRef.URI, WarningCount: len(pageResult.Validation.Warnings), TableCount: tableCount, FigureCount: figureCount, RenderedBytes: len(renderedBytes)}
		if err := insertStructuredWarnings(ctx, projection, input, pageResult); err != nil {
			return workflow.Retryable("structured_projection_warning_failed", err)
		}
		if err := markStructuredPageSucceeded(ctx, projection, input, result); err != nil {
			return workflow.Retryable("structured_projection_update_failed", err)
		}
		return step.Result(result)
	})
}

func StructuredAssembleExecutor(projectionName string) workflow.Executor {
	return workflow.NewTypedExecutor(KindStructuredAssemble, func(ctx context.Context, step *workflow.StepContext, input StructuredAssembleInput) error {
		deps := step.Step().DependsOn
		pages := make([]StructuredPageWorkflowResult, 0, len(deps))
		for _, dep := range deps {
			var page StructuredPageWorkflowResult
			if err := step.DependencyData(dep.OpID, &page); err != nil {
				return workflow.Retryable("structured_dependency_load_failed", err)
			}
			pages = append(pages, page)
		}
		sort.Slice(pages, func(i, j int) bool { return pages[i].PageNumber < pages[j].PageNumber })
		var out strings.Builder
		for _, page := range pages {
			body, err := os.ReadFile(page.RenderedMarkdownPath)
			if err != nil {
				return workflow.Retryable("structured_rendered_read_failed", err)
			}
			if out.Len() > 0 {
				out.WriteString("\n")
			}
			out.Write(body)
		}
		if err := os.MkdirAll(input.WorkDir, 0o755); err != nil {
			return workflow.Retryable("structured_workdir_failed", err)
		}
		assembledPath := filepath.Join(input.WorkDir, "assembled.md")
		assembledBytes := []byte(out.String())
		if err := os.WriteFile(assembledPath, assembledBytes, 0o644); err != nil {
			return workflow.Retryable("structured_assemble_write_failed", err)
		}
		ref, err := step.StoreArtifact("assembled.md", "text/markdown", assembledBytes, workflow.ArtifactKind("structured-assembled-markdown"))
		if err != nil {
			return workflow.Retryable("structured_store_artifact_failed", err)
		}
		result := StructuredAssembleResult{BookID: input.BookID, PageCount: len(pages), MarkdownPath: assembledPath, MarkdownRefID: ref.ID, MarkdownURI: ref.URI, CharCount: len(assembledBytes)}
		if input.EmbedFigures {
			figuresDir := strings.TrimSpace(input.FiguresDir)
			if figuresDir == "" {
				figuresDir = filepath.Join(input.WorkDir, "figures")
			}
			embedded, figures, err := ocrquality.EmbedExtractedFigures(string(assembledBytes), ocrquality.FigureExtractionOptions{ImageDir: input.ImageDir, OutputDir: figuresDir})
			if err != nil {
				return workflow.Permanent("structured_embed_figures_failed", err)
			}
			embeddedPath := filepath.Join(input.WorkDir, "embedded-figures.md")
			if err := os.WriteFile(embeddedPath, []byte(embedded), 0o644); err != nil {
				return workflow.Retryable("structured_embedded_write_failed", err)
			}
			embeddedRef, err := step.StoreArtifact("embedded-figures.md", "text/markdown", []byte(embedded), workflow.ArtifactKind("structured-embedded-markdown"))
			if err != nil {
				return workflow.Retryable("structured_store_artifact_failed", err)
			}
			for _, figure := range figures {
				imageBytes, err := os.ReadFile(figure.ImagePath)
				if err != nil {
					return workflow.Retryable("structured_figure_read_failed", err)
				}
				metadata := map[string]string{"page": fmt.Sprintf("%03d", figure.PageNumber), "description": figure.Description}
				if _, err := step.StoreArtifact(filepath.Base(figure.ImagePath), "image/png", imageBytes, workflow.ArtifactKind("structured-extracted-figure"), workflow.ArtifactMetadata(metadata)); err != nil {
					return workflow.Retryable("structured_figure_artifact_failed", err)
				}
				if strings.TrimSpace(figure.SidecarPath) != "" {
					sidecarBytes, err := os.ReadFile(figure.SidecarPath)
					if err != nil {
						return workflow.Retryable("structured_figure_sidecar_read_failed", err)
					}
					if _, err := step.StoreArtifact(filepath.Base(figure.SidecarPath), "application/json", sidecarBytes, workflow.ArtifactKind("structured-figure-sidecar"), workflow.ArtifactMetadata(metadata)); err != nil {
						return workflow.Retryable("structured_figure_sidecar_artifact_failed", err)
					}
				}
				if strings.TrimSpace(figure.DebugPath) != "" {
					debugBytes, err := os.ReadFile(figure.DebugPath)
					if err != nil {
						return workflow.Retryable("structured_figure_debug_read_failed", err)
					}
					if _, err := step.StoreArtifact(filepath.Base(figure.DebugPath), "image/png", debugBytes, workflow.ArtifactKind("structured-figure-debug-overlay"), workflow.ArtifactMetadata(metadata)); err != nil {
						return workflow.Retryable("structured_figure_debug_artifact_failed", err)
					}
				}
			}
			result.EmbeddedMarkdownPath = embeddedPath
			result.EmbeddedRefID = embeddedRef.ID
			result.EmbeddedURI = embeddedRef.URI
			result.FiguresDir = figuresDir
			result.FigureCount = len(figures)
		}
		return step.Result(result)
	})
}

func StructuredValidateExecutor(projectionName string) workflow.Executor {
	return workflow.NewTypedExecutor(KindStructuredValidate, func(ctx context.Context, step *workflow.StepContext, input StructuredValidateInput) error {
		var assemble StructuredAssembleResult
		if len(step.Step().DependsOn) > 0 {
			if err := step.DependencyData(step.Step().DependsOn[0].OpID, &assemble); err != nil {
				return workflow.Retryable("structured_dependency_load_failed", err)
			}
		}
		body, err := os.ReadFile(assemble.MarkdownPath)
		if err != nil {
			return workflow.Retryable("structured_assembled_read_failed", err)
		}
		pages := splitRenderedPages(string(body))
		warnings := ocrvalidation.DetectAdjacentDuplicateFigureCaptions(pages)
		adjacent := make([]string, 0, len(warnings))
		for _, warning := range warnings {
			adjacent = append(adjacent, warning.Message)
		}
		if input.ExpectedPages > 0 && len(pages) != input.ExpectedPages {
			return workflow.Permanent("structured_expected_pages_mismatch", fmt.Errorf("expected %d pages, got %d", input.ExpectedPages, len(pages)))
		}
		shortPages, err := loadShortStructuredPages(ctx, step, projectionNameOrDefaultStructured(projectionName), input.BookID, input.MinRenderedBytes)
		if err != nil {
			return workflow.Retryable("structured_projection_short_page_query_failed", err)
		}
		reportPath := filepath.Join(input.WorkDir, "validation-report.json")
		result := StructuredValidateResult{BookID: input.BookID, PageCount: len(pages), ExpectedPages: input.ExpectedPages, WarningCount: len(warnings) + len(shortPages), AdjacentDuplicateCaptions: adjacent, ShortPages: shortPages, MinRenderedBytes: input.MinRenderedBytes, ReportPath: reportPath}
		reportBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		ref, err := step.StoreArtifact("validation-report.json", "application/json", append(reportBytes, '\n'), workflow.ArtifactKind("structured-validation-report"))
		if err != nil {
			return workflow.Retryable("structured_store_artifact_failed", err)
		}
		result.ReportRefID = ref.ID
		result.ReportURI = ref.URI
		reportBytes, err = json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(reportPath, append(reportBytes, '\n'), 0o644); err != nil {
			return workflow.Retryable("structured_validation_write_failed", err)
		}
		return step.Result(result)
	})
}

func loadShortStructuredPages(ctx context.Context, step *workflow.StepContext, projectionName, bookID string, minRenderedBytes int) ([]StructuredShortPageWarning, error) {
	if minRenderedBytes <= 0 {
		return nil, nil
	}
	projection, err := step.Projection(projectionName)
	if err != nil {
		return nil, err
	}
	rows, err := projection.Query(ctx, `SELECT page_num, rendered_bytes, rendered_markdown_path FROM structured_pages WHERE book_id = ? AND status = 'succeeded' AND rendered_bytes > 0 AND rendered_bytes < ? ORDER BY page_num`, bookID, minRenderedBytes)
	if err != nil {
		return nil, err
	}
	out := make([]StructuredShortPageWarning, 0, len(rows))
	for _, row := range rows {
		out = append(out, StructuredShortPageWarning{PageNumber: intFromAny(row["page_num"]), RenderedBytes: intFromAny(row["rendered_bytes"]), RenderedMarkdownPath: fmt.Sprint(row["rendered_markdown_path"])})
	}
	return out, nil
}

func intFromAny(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case []byte:
		var n int
		_, _ = fmt.Sscanf(string(v), "%d", &n)
		return n
	case string:
		var n int
		_, _ = fmt.Sscanf(v, "%d", &n)
		return n
	default:
		var n int
		_, _ = fmt.Sscanf(fmt.Sprint(v), "%d", &n)
		return n
	}
}

func splitRenderedPages(markdown string) []ocrvalidation.PageText {
	parts := strings.Split(markdown, "<!-- page:")
	out := make([]ocrvalidation.PageText, 0, len(parts))
	for _, part := range parts[1:] {
		if len(part) < 3 {
			continue
		}
		var num int
		if _, err := fmt.Sscanf(part[:3], "%d", &num); err != nil {
			continue
		}
		out = append(out, ocrvalidation.PageText{Number: num, Markdown: "<!-- page:" + part})
	}
	return out
}

func classifyStructuredPageError(err error) error {
	code := classifyStructuredPageErrorCode(err)
	switch code {
	case "structured_image_read_failed", "structured_parse_failed", "structured_page_mismatch":
		return workflow.Permanent(code, err)
	default:
		return workflow.Retryable(code, err)
	}
}

func classifyStructuredPageErrorCode(err error) string {
	if err == nil {
		return "structured_page_failed"
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "read target image"):
		return "structured_image_read_failed"
	case strings.Contains(msg, "parse structured ocr json"):
		return "structured_parse_failed"
	case strings.Contains(msg, "page mismatch"):
		return "structured_page_mismatch"
	case strings.Contains(msg, "tls"):
		return "structured_provider_tls"
	case strings.Contains(msg, "rate limit") || strings.Contains(msg, "429"):
		return "structured_provider_rate_limit"
	case strings.Contains(msg, "run structured ocr inference"):
		return "structured_provider_failed"
	default:
		return "structured_page_failed"
	}
}

func projectionNameOrDefaultStructured(name string) string {
	if name == "" {
		return StructuredProjectionName
	}
	return name
}
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return "structured-run"
}
