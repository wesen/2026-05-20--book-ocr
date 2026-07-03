package main

import (
	"github.com/spf13/cobra"

	"github.com/go-go-golems/book-ocr/internal/vlmseparation"
)

// newRootCommand builds the cobra command tree. Subcommands keep their
// hand-written stdlib FlagSets (DisableFlagParsing passes argv through
// untouched, including -h), so flag names and behavior are identical to the
// pre-cobra CLI; cobra contributes the command tree, help, and completion.
func newRootCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "book-ocr",
		Short: "Workflow-backed structured OCR for scanned technical books",
		Long: `book-ocr turns a directory of scanned page images (or a PDF) into durable
review artifacts: structured per-page JSON, deterministic Markdown, figure
crops, validation reports, and a rendered PDF.

Typical flow:
  book-ocr init --book-id my-book --pdf book.pdf        # ingest + drafted profile
  book-ocr structured-run --book-profile ... --dry-run  # offline check
  book-ocr structured-run --book-profile ... --profile gpt-5-mini-low
  book-ocr report --work-dir ...                        # status + warnings
  book-ocr structured-rerun-pages --pages 12,40         # targeted repair`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	type sub struct {
		name  string
		short string
		fn    func([]string) error
	}
	for _, s := range []sub{
		{"structured-run", "Run the structured OCR workflow over a page range", runStructuredRun},
		{"structured-page", "Run structured OCR for a single page (no workflow engine)", runStructuredPage},
		{"structured-rerun-pages", "Requeue selected pages and rebuild downstream artifacts", structuredRerunPages},
		{"structured-pages", "List structured page projection rows", listStructuredPages},
		{"init", "Bootstrap a new book workspace with a drafted profile", runInit},
		{"ingest", "Rasterize a PDF into page_NNNN.png images", runIngest},
		{"report", "Summarize a run from its projection and turn store", runReport},
		{"resume", "Resume workers for an existing run", resumeRun},
		{"retry", "Retry a failed step", retryStep},
		{"cancel", "Cancel a run", cancelRun},
		{"status", "Show workflow run status", showStatus},
		{"quality-pass", "Run the QA/normalization quality pass over assembled Markdown", runQualityPass},
		{"run", "Run the legacy freeform OCR workflow (superseded by structured-run)", runWorkflow},
		{"pages", "List legacy freeform page projection rows", listPages},
	} {
		root.AddCommand(&cobra.Command{
			Use:                s.name,
			Short:              s.short,
			DisableFlagParsing: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				return s.fn(args)
			},
		})
	}

	vlm, err := vlmseparation.NewRootCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(vlm)
	return root, nil
}
