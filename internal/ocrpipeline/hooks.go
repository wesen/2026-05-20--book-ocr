package ocrpipeline

import "github.com/go-go-golems/book-ocr/internal/ocrvalidation"

// ResponseParser is the S3 seam: convert a raw model response into a
// structured page. handled=false defers to the built-in layered parser
// (strict JSON → fence-strip → sanitize → regex repair), so a plugin can
// handle only the formats it recognizes.
type ResponseParser interface {
	ParseResponse(raw string, input StructuredOCRInput) (page StructuredPageOCR, handled bool, err error)
}

// PageValidator is the page half of the S7 seam. Warnings are additive: they
// append to the built-in validation, never replace it.
type PageValidator interface {
	ValidatePage(page StructuredPageOCR, rendered string) ([]ocrvalidation.Warning, error)
}

// BookValidator is the book half of the S7 seam, run by the validate step
// after assembly. It receives paths and page identities, not page content —
// a validator that needs the text reads the files.
type BookValidator interface {
	ValidateBook(bookID string, assembledPath string, pages []ocrvalidation.PageText) ([]ocrvalidation.Warning, error)
}

// PageClassifier is the S6 seam: called once per discovered page before
// fan-out. pageType flows to the OCR client as a hint; strategy selects a
// specific ocr.page plugin binding for that page (empty = default binding).
type PageClassifier interface {
	ClassifyPage(imagePath string, pageNumber int) (pageType string, strategy string, err error)
}

// RetryHinter lets adapter errors carry an explicit retryability verdict into
// the page-step error classification (e.g. a plugin's error.details.retryable
// or an op timeout), instead of falling through to string matching.
type RetryHinter interface {
	RetryableHint() bool
}
