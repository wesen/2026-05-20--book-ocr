package plugin

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/go-go-golems/book-ocr/internal/ocrpipeline"
	"github.com/go-go-golems/book-ocr/internal/ocrvalidation"
	"github.com/go-go-golems/devctl/pkg/runtime"
	"github.com/stretchr/testify/require"
)

func allSeamsSpec(t *testing.T) Spec {
	return Spec{ID: "test", Path: testPluginPath(t, "test_plugin.py"), Seams: []string{OpOCRPage, OpResponseParse, OpValidatePage, OpValidateBook, OpPageClassify}}
}

func TestPluginResponseParserHandlesCustomFormat(t *testing.T) {
	requirePython3(t)
	mgr := newTestManager(t, []Spec{allSeamsSpec(t)})
	parser := NewResponseParser(mgr)

	page, handled, err := parser.ParseResponse("@@7|Parsed by the plugin.", ocrpipeline.StructuredOCRInput{BookID: "b", PageNumber: 7})
	require.NoError(t, err)
	require.True(t, handled)
	require.Equal(t, 7, page.PageNumber)
	require.Equal(t, "Parsed by the plugin.", page.Blocks[0].Text)
}

func TestPluginResponseParserDeclinesToBuiltin(t *testing.T) {
	requirePython3(t)
	mgr := newTestManager(t, []Spec{allSeamsSpec(t)})
	parser := NewResponseParser(mgr)

	_, handled, err := parser.ParseResponse(`{"schema_version":"structured-ocr/v1"}`, ocrpipeline.StructuredOCRInput{BookID: "b", PageNumber: 3})
	require.NoError(t, err)
	require.False(t, handled, "JSON input must be declined to the builtin parser")
}

// End-to-end: a declining parser leaves the builtin path fully intact.
func TestRunStructuredPageWithDecliningParserUsesBuiltin(t *testing.T) {
	requirePython3(t)
	mgr := newTestManager(t, []Spec{allSeamsSpec(t)})

	workDir := t.TempDir()
	imagePath := filepath.Join(workDir, "page_012.png")
	writeTestPNG(t, imagePath, 50, 50)
	result, err := ocrpipeline.RunStructuredPage(context.Background(), ocrpipeline.StructuredOCRInput{
		BookID: "plugin-book", RunID: "r", PageNumber: 12, ImagePath: imagePath, WorkDir: workDir,
		Parser: NewResponseParser(mgr),
	}, ocrpipeline.DryRunStructuredOCRClient{})
	require.NoError(t, err)
	rendered, err := os.ReadFile(result.RenderedMD)
	require.NoError(t, err)
	require.Contains(t, string(rendered), "rudimentary user interface")
}

func TestPluginPageValidatorAppendsTaggedWarnings(t *testing.T) {
	requirePython3(t)
	mgr := newTestManager(t, []Spec{allSeamsSpec(t)})

	workDir := t.TempDir()
	imagePath := filepath.Join(workDir, "page_005.png")
	writeTestPNG(t, imagePath, 50, 50)
	// The test plugin's ocr.page emits "print('page 5')" — no TYPO marker;
	// use a custom raw via the plugin OCR path? Simpler: validate directly.
	validator := NewPageValidator(mgr)
	warnings, err := validator.ValidatePage(ocrpipeline.StructuredPageOCR{BookID: "b", PageNumber: 5}, "text with TYPO inside")
	require.NoError(t, err)
	require.Len(t, warnings, 1)
	require.Equal(t, "typo_found", warnings[0].Code)
	require.Contains(t, warnings[0].Message, "[plugin/test]")
	require.Equal(t, 5, warnings[0].Page)

	// And through RunStructuredPage: the warning lands in 06-validation.json.
	result, err := ocrpipeline.RunStructuredPage(context.Background(), ocrpipeline.StructuredOCRInput{
		BookID: "plugin-book", RunID: "r", PageNumber: 5, ImagePath: imagePath, WorkDir: workDir,
		PageValidators: []ocrpipeline.PageValidator{validator},
	}, typoClient{})
	require.NoError(t, err)
	validationJSON, err := os.ReadFile(result.ValidationJSON)
	require.NoError(t, err)
	require.Contains(t, string(validationJSON), "typo_found")
}

// typoClient produces a page whose rendered markdown contains TYPO.
type typoClient struct{}

func (typoClient) OCRPage(ctx context.Context, input ocrpipeline.StructuredOCRInput, imageBytes []byte) (ocrpipeline.StructuredOCRResult, error) {
	res, err := ocrpipeline.DryRunStructuredOCRClient{}.OCRPage(ctx, input, imageBytes)
	if err != nil {
		return res, err
	}
	res.RawResponse = `{"schema_version":"structured-ocr/v1","book_id":"` + input.BookID + `","page_number":` +
		strconv.Itoa(input.PageNumber) + `,"page_type":"body","blocks":[{"id":"b1","type":"paragraph","text":"This page has a TYPO in it."}]}`
	return res, nil
}

func TestPluginBookValidator(t *testing.T) {
	requirePython3(t)
	mgr := newTestManager(t, []Spec{allSeamsSpec(t)})
	validator := NewBookValidator(mgr)
	warnings, err := validator.ValidateBook("b", "/tmp/assembled.md", []ocrvalidation.PageText{{Number: 1}, {Number: 2}})
	require.NoError(t, err)
	require.Len(t, warnings, 1)
	require.Equal(t, "book_checked", warnings[0].Code)
	require.Contains(t, warnings[0].Message, "checked 2 pages")
}

func TestPluginPageClassifierAndStrategyRouting(t *testing.T) {
	requirePython3(t)
	mgr := newTestManager(t, []Spec{
		allSeamsSpec(t),
		{ID: "alt-ocr", Path: testPluginPath(t, "alt_ocr_plugin.py"), Seams: []string{OpOCRPage}},
	})
	classifier := NewPageClassifier(mgr)

	pageType, strategy, err := classifier.ClassifyPage("/img/page_0001.png", 1)
	require.NoError(t, err)
	require.Equal(t, "title", pageType)
	require.Empty(t, strategy)

	_, strategy, err = classifier.ClassifyPage("/img/page_0002.png", 2)
	require.NoError(t, err)
	require.Equal(t, "alt-ocr", strategy)

	// Strategy routes the OCR call to the alt plugin even though the default
	// seam binding is the main test plugin.
	client := NewStructuredOCRClient(mgr, nil)
	workDir := t.TempDir()
	imagePath := filepath.Join(workDir, "page_0002.png")
	writeTestPNG(t, imagePath, 50, 50)
	result, err := ocrpipeline.RunStructuredPage(context.Background(), ocrpipeline.StructuredOCRInput{
		BookID: "plugin-book", RunID: "r", PageNumber: 2, ImagePath: imagePath, WorkDir: workDir,
		Strategy: "alt-ocr",
	}, client)
	require.NoError(t, err)
	rendered, err := os.ReadFile(result.RenderedMD)
	require.NoError(t, err)
	require.Contains(t, string(rendered), "ALT strategy output for page 2.")
}

func TestWrapCallErrorCarriesRetryHint(t *testing.T) {
	retryable := wrapCallError(&runtime.OpError{Code: "E_RUNTIME", Details: map[string]any{"retryable": true}})
	var hinter ocrpipeline.RetryHinter
	require.True(t, errors.As(retryable, &hinter))
	require.True(t, hinter.RetryableHint())

	permanent := wrapCallError(&runtime.OpError{Code: "E_RUNTIME", Details: map[string]any{"retryable": false}})
	require.True(t, errors.As(permanent, &hinter))
	require.False(t, hinter.RetryableHint())

	timeout := wrapCallError(&runtime.OpError{Code: "E_TIMEOUT"})
	require.True(t, errors.As(timeout, &hinter))
	require.True(t, hinter.RetryableHint())

	unhinted := wrapCallError(&runtime.OpError{Code: "E_RUNTIME"})
	require.False(t, errors.As(unhinted, &hinter))
}
