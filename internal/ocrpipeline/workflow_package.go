package ocrpipeline

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/book-ocr/internal/ocrquality"
	"github.com/go-go-golems/scraper/pkg/engine/model"
	"github.com/go-go-golems/scraper/pkg/workflow"
)

type StructuredWorkflowConfig struct {
	Client         StructuredOCRClient
	ProjectionName string
	// Segmenter computes figure crops during assembly; nil selects the
	// built-in ink-band heuristic.
	Segmenter ocrquality.FigureSegmenter
	// Parser optionally overrides raw-response parsing (response.parse seam).
	Parser ResponseParser
	// PageValidators / BookValidators append plugin warnings to the built-in
	// validation (validate.page / validate.book seams).
	PageValidators []PageValidator
	BookValidators []BookValidator
	// Classifier runs per discovered page and stamps a page-type hint and
	// strategy into the page input (page.classify seam).
	Classifier PageClassifier
}

func RegisterStructuredWorkflow(rt *workflow.Runtime, cfg StructuredWorkflowConfig) error {
	if rt == nil {
		return fmt.Errorf("workflow runtime is nil")
	}
	if cfg.Client == nil {
		cfg.Client = DryRunStructuredOCRClient{}
	}
	if cfg.ProjectionName == "" {
		cfg.ProjectionName = StructuredProjectionName
	}
	if err := rt.RegisterPackage(StructuredPackage()); err != nil {
		return err
	}
	for _, executor := range []workflow.Executor{
		StructuredDiscoverExecutor(cfg.ProjectionName, cfg.Classifier),
		StructuredPageExecutor(cfg.ProjectionName, cfg.Client, cfg.Parser, cfg.PageValidators),
		StructuredAssembleExecutor(cfg.ProjectionName, cfg.Segmenter),
		StructuredValidateExecutor(cfg.ProjectionName, cfg.BookValidators),
	} {
		if err := rt.RegisterExecutor(executor); err != nil {
			return err
		}
	}
	return nil
}

func StructuredPackage() *workflow.Package {
	return workflow.NewPackage(StructuredPackageName).
		DisplayName("Structured OCR Workflow").
		Entrypoint(workflow.EntrypointFunc[StructuredRunInput](func(ctx context.Context, run *workflow.RunBuilder, input StructuredRunInput) error {
			input = normalizeStructuredRunInput(input)
			if err := validateStructuredRunInput(input); err != nil {
				return err
			}
			run.Metadata("book_id", input.BookID)
			run.Metadata("image_dir", input.ImageDir)
			run.Metadata("work_dir", input.WorkDir)
			_, err := run.Step("discover-structured-pages", input, workflow.StepOpts{Kind: KindStructuredDiscover, Queue: QueueStructuredControl, Retry: model.RetryPolicy{MaxAttempts: 1}})
			return err
		})).
		Build()
}

func normalizeStructuredRunInput(input StructuredRunInput) StructuredRunInput {
	input.BookID = strings.TrimSpace(input.BookID)
	input.ImageDir = strings.TrimSpace(input.ImageDir)
	input.PageGlob = strings.TrimSpace(input.PageGlob)
	if input.PageGlob == "" {
		input.PageGlob = "page_*.png"
	}
	input.WorkDir = strings.TrimSpace(input.WorkDir)
	input.RunID = strings.TrimSpace(input.RunID)
	input.Profile = strings.TrimSpace(input.Profile)
	return input
}

func validateStructuredRunInput(input StructuredRunInput) error {
	if input.BookID == "" {
		return fmt.Errorf("book_id is required")
	}
	if input.ImageDir == "" {
		return fmt.Errorf("image_dir is required")
	}
	if input.WorkDir == "" {
		return fmt.Errorf("work_dir is required")
	}
	if input.StartPage > 0 && input.EndPage > 0 && input.StartPage > input.EndPage {
		return fmt.Errorf("start_page must be <= end_page")
	}
	return nil
}

func structuredOCRRetryPolicy() model.RetryPolicy {
	return model.RetryPolicy{MaxAttempts: 3, BackoffKind: model.BackoffKindExponential, InitialBackoff: time.Second, MaxBackoff: 30 * time.Second, Multiplier: 2}
}
