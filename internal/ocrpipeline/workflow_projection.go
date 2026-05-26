package ocrpipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/go-go-golems/scraper/pkg/workflow"
)

func EnsureStructuredProjectionSchema(ctx context.Context, projection workflow.Projection) error {
	_, err := projection.Exec(ctx, `CREATE TABLE IF NOT EXISTS structured_pages (
		book_id TEXT NOT NULL,
		page_num INTEGER NOT NULL,
		image_path TEXT NOT NULL,
		status TEXT NOT NULL,
		step_id TEXT,
		page_dir TEXT,
		raw_response_path TEXT,
		structured_json_path TEXT,
		rendered_markdown_path TEXT,
		validation_json_path TEXT,
		structured_artifact_id TEXT,
		structured_artifact_uri TEXT,
		rendered_artifact_id TEXT,
		rendered_artifact_uri TEXT,
		validation_artifact_id TEXT,
		validation_artifact_uri TEXT,
		warning_count INTEGER NOT NULL DEFAULT 0,
		table_count INTEGER NOT NULL DEFAULT 0,
		figure_count INTEGER NOT NULL DEFAULT 0,
		rendered_bytes INTEGER NOT NULL DEFAULT 0,
		error_code TEXT,
		error_message TEXT,
		updated_at TEXT NOT NULL,
		PRIMARY KEY(book_id, page_num)
	)`)
	if err != nil {
		return err
	}
	_, err = projection.Exec(ctx, `CREATE TABLE IF NOT EXISTS structured_warnings (
		book_id TEXT NOT NULL,
		page_num INTEGER NOT NULL,
		code TEXT NOT NULL,
		message TEXT NOT NULL,
		value TEXT,
		updated_at TEXT NOT NULL
	)`)
	return err
}

func seedStructuredPage(ctx context.Context, projection workflow.Projection, input StructuredPageWorkflowInput, stepID string) error {
	_, err := projection.Exec(ctx, `INSERT OR REPLACE INTO structured_pages(book_id,page_num,image_path,status,step_id,updated_at) VALUES(?,?,?,?,?,?)`, input.BookID, input.PageNumber, input.ImagePath, "pending", stepID, nowUTC())
	return err
}

func markStructuredPageRunning(ctx context.Context, projection workflow.Projection, input StructuredPageWorkflowInput) error {
	_, err := projection.Exec(ctx, `UPDATE structured_pages SET status='running', error_code=NULL, error_message=NULL, updated_at=? WHERE book_id=? AND page_num=?`, nowUTC(), input.BookID, input.PageNumber)
	return err
}

func markStructuredPageSucceeded(ctx context.Context, projection workflow.Projection, input StructuredPageWorkflowInput, result StructuredPageWorkflowResult) error {
	_, err := projection.Exec(ctx, `UPDATE structured_pages SET status='succeeded', page_dir=?, raw_response_path=?, structured_json_path=?, rendered_markdown_path=?, validation_json_path=?, structured_artifact_id=?, structured_artifact_uri=?, rendered_artifact_id=?, rendered_artifact_uri=?, validation_artifact_id=?, validation_artifact_uri=?, warning_count=?, table_count=?, figure_count=?, rendered_bytes=?, error_code=NULL, error_message=NULL, updated_at=? WHERE book_id=? AND page_num=?`, result.PageDir, result.RawResponsePath, result.StructuredJSONPath, result.RenderedMarkdownPath, result.ValidationJSONPath, result.StructuredArtifactID, result.StructuredArtifactURI, result.RenderedArtifactID, result.RenderedArtifactURI, result.ValidationArtifactID, result.ValidationArtifactURI, result.WarningCount, result.TableCount, result.FigureCount, result.RenderedBytes, nowUTC(), input.BookID, input.PageNumber)
	return err
}

func markStructuredPageError(ctx context.Context, projection workflow.Projection, input StructuredPageWorkflowInput, code string, err error) error {
	message := ""
	if err != nil {
		message = err.Error()
	}
	_, execErr := projection.Exec(ctx, `UPDATE structured_pages SET status='failed', error_code=?, error_message=?, updated_at=? WHERE book_id=? AND page_num=?`, code, message, nowUTC(), input.BookID, input.PageNumber)
	return execErr
}

func insertStructuredWarnings(ctx context.Context, projection workflow.Projection, input StructuredPageWorkflowInput, result StructuredPageRunResult) error {
	_, _ = projection.Exec(ctx, `DELETE FROM structured_warnings WHERE book_id=? AND page_num=?`, input.BookID, input.PageNumber)
	for _, warning := range result.Validation.Warnings {
		if _, err := projection.Exec(ctx, `INSERT INTO structured_warnings(book_id,page_num,code,message,value,updated_at) VALUES(?,?,?,?,?,?)`, input.BookID, input.PageNumber, warning.Code, warning.Message, warning.Value, nowUTC()); err != nil {
			return err
		}
	}
	return nil
}

func countStructuredBlocks(page StructuredPageOCR) (tables, figures int) {
	for _, block := range page.Blocks {
		switch block.Type {
		case BlockTable:
			tables++
		case BlockFigure:
			figures++
		}
	}
	return tables, figures
}

func nowUTC() string { return time.Now().UTC().Format(time.RFC3339) }

func structuredPageStepID(pageNum int) string { return fmt.Sprintf("structured-page-%03d", pageNum) }
