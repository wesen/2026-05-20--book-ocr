package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/book-ocr/internal/ocrpipeline"
	"github.com/go-go-golems/book-ocr/internal/ocrvalidation"
	"github.com/go-go-golems/devctl/pkg/protocol"
	"github.com/go-go-golems/devctl/pkg/runtime"
)

// hintedError carries a plugin's explicit retryability verdict into the
// workflow error classification (ocrpipeline.RetryHinter).
type hintedError struct {
	err       error
	retryable bool
}

func (e *hintedError) Error() string       { return e.err.Error() }
func (e *hintedError) Unwrap() error       { return e.err }
func (e *hintedError) RetryableHint() bool { return e.retryable }

var _ ocrpipeline.RetryHinter = &hintedError{}

// wrapCallError attaches a retryability hint when the plugin error carries
// one: error.details.retryable, or a timeout/cancellation code. Errors
// without a verdict pass through untouched and fall back to the host's
// string-based classification.
func wrapCallError(err error) error {
	if err == nil {
		return nil
	}
	var opErr *runtime.OpError
	if !errors.As(err, &opErr) {
		return err
	}
	if v, ok := opErr.Details["retryable"].(bool); ok {
		return &hintedError{err: err, retryable: v}
	}
	if opErr.Code == protocol.ErrTimeout || opErr.Code == protocol.ErrCanceled {
		return &hintedError{err: err, retryable: true}
	}
	return err
}

// declined reports whether the plugin explicitly deferred to the built-in
// implementation (error code E_DECLINED).
func declined(err error) bool {
	var opErr *runtime.OpError
	return errors.As(err, &opErr) && opErr.Code == ErrDeclined
}

// ResponseParser delegates raw-response parsing (seam S3, op response.parse).
// A plugin that doesn't recognize the format answers E_DECLINED and the
// built-in layered parser takes over.
type ResponseParser struct {
	mgr *Manager
}

var _ ocrpipeline.ResponseParser = &ResponseParser{}

func NewResponseParser(mgr *Manager) *ResponseParser {
	return &ResponseParser{mgr: mgr}
}

func (p *ResponseParser) ParseResponse(raw string, input ocrpipeline.StructuredOCRInput) (ocrpipeline.StructuredPageOCR, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	req := ResponseParseInput{OpSchema: "response.parse/v1", BookID: input.BookID, PageNumber: input.PageNumber, RawResponse: raw}
	var out OCRPageOutput
	if err := p.mgr.Call(ctx, OpResponseParse, req, &out); err != nil {
		if declined(err) {
			return ocrpipeline.StructuredPageOCR{}, false, nil
		}
		return ocrpipeline.StructuredPageOCR{}, true, wrapCallError(fmt.Errorf("plugin response.parse (%s): %w", p.mgr.PluginIDFor(OpResponseParse), err))
	}
	if len(out.Page) == 0 {
		return ocrpipeline.StructuredPageOCR{}, true, fmt.Errorf("plugin response.parse (%s) returned no page object", p.mgr.PluginIDFor(OpResponseParse))
	}
	var page ocrpipeline.StructuredPageOCR
	if err := json.Unmarshal(out.Page, &page); err != nil {
		return ocrpipeline.StructuredPageOCR{}, true, fmt.Errorf("plugin response.parse (%s) page object: %w", p.mgr.PluginIDFor(OpResponseParse), err)
	}
	return page, true, nil
}

// PageValidator delegates per-page validation (seam S7, op validate.page).
// Warnings are tagged with the plugin id and appended to the built-ins.
type PageValidator struct {
	mgr *Manager
}

var _ ocrpipeline.PageValidator = &PageValidator{}

func NewPageValidator(mgr *Manager) *PageValidator {
	return &PageValidator{mgr: mgr}
}

func (v *PageValidator) ValidatePage(page ocrpipeline.StructuredPageOCR, rendered string) ([]ocrvalidation.Warning, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	structured, err := json.Marshal(page)
	if err != nil {
		return nil, err
	}
	req := ValidatePageInput{OpSchema: "validate.page/v1", BookID: page.BookID, PageNumber: page.PageNumber, Structured: structured, Markdown: rendered}
	var out ValidateOutput
	if err := v.mgr.Call(ctx, OpValidatePage, req, &out); err != nil {
		return nil, fmt.Errorf("plugin validate.page (%s): %w", v.mgr.PluginIDFor(OpValidatePage), err)
	}
	return tagWarnings(out.Warnings, v.mgr.PluginIDFor(OpValidatePage), page.PageNumber), nil
}

// BookValidator delegates book-level validation (seam S7, op validate.book).
type BookValidator struct {
	mgr *Manager
}

var _ ocrpipeline.BookValidator = &BookValidator{}

func NewBookValidator(mgr *Manager) *BookValidator {
	return &BookValidator{mgr: mgr}
}

func (v *BookValidator) ValidateBook(bookID string, assembledPath string, pages []ocrvalidation.PageText) ([]ocrvalidation.Warning, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	numbers := make([]int, 0, len(pages))
	for _, page := range pages {
		numbers = append(numbers, page.Number)
	}
	req := ValidateBookInput{OpSchema: "validate.book/v1", BookID: bookID, AssembledPath: assembledPath, PageNumbers: numbers}
	var out ValidateOutput
	if err := v.mgr.Call(ctx, OpValidateBook, req, &out); err != nil {
		return nil, fmt.Errorf("plugin validate.book (%s): %w", v.mgr.PluginIDFor(OpValidateBook), err)
	}
	return tagWarnings(out.Warnings, v.mgr.PluginIDFor(OpValidateBook), 0), nil
}

func tagWarnings(in []PluginWarning, pluginID string, defaultPage int) []ocrvalidation.Warning {
	out := make([]ocrvalidation.Warning, 0, len(in))
	for _, w := range in {
		page := w.Page
		if page == 0 {
			page = defaultPage
		}
		message := strings.TrimSpace(w.Message)
		out = append(out, ocrvalidation.Warning{Code: w.Code, Message: fmt.Sprintf("[plugin/%s] %s", pluginID, message), Page: page})
	}
	return out
}

// PageClassifier delegates page routing (seam S6, op page.classify).
type PageClassifier struct {
	mgr *Manager
}

var _ ocrpipeline.PageClassifier = &PageClassifier{}

func NewPageClassifier(mgr *Manager) *PageClassifier {
	return &PageClassifier{mgr: mgr}
}

func (c *PageClassifier) ClassifyPage(imagePath string, pageNumber int) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	req := PageClassifyInput{OpSchema: "page.classify/v1", ImagePath: imagePath, PageNumber: pageNumber}
	var out PageClassifyOutput
	if err := c.mgr.Call(ctx, OpPageClassify, req, &out); err != nil {
		return "", "", fmt.Errorf("plugin page.classify (%s): %w", c.mgr.PluginIDFor(OpPageClassify), err)
	}
	if strategy := strings.TrimSpace(out.Strategy); strategy != "" && !c.mgr.HasPlugin(strategy, OpOCRPage) {
		return "", "", fmt.Errorf("plugin page.classify (%s) routed page %d to unknown ocr.page strategy %q", c.mgr.PluginIDFor(OpPageClassify), pageNumber, strategy)
	}
	return strings.TrimSpace(out.PageType), strings.TrimSpace(out.Strategy), nil
}
