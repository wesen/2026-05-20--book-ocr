package plugin

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/book-ocr/internal/ocrpipeline"
	"github.com/go-go-golems/book-ocr/internal/ocrquality"
	"github.com/stretchr/testify/require"
)

func writeTestPNG(t *testing.T, path string, width, height int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.White)
		}
	}
	// A block of ink so the built-in segmenter would also find something.
	for y := height / 4; y < height/2; y++ {
		for x := width / 4; x < width/2; x++ {
			img.Set(x, y, color.Black)
		}
	}
	f, err := os.Create(path)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	require.NoError(t, png.Encode(f, img))
}

func TestPluginOCRPageThroughRunStructuredPage(t *testing.T) {
	requirePython3(t)
	mgr := newTestManager(t, []Spec{{ID: "test", Path: testPluginPath(t, "test_plugin.py"), Seams: []string{OpOCRPage}}})
	client := NewStructuredOCRClient(mgr, nil)

	workDir := t.TempDir()
	imagePath := filepath.Join(workDir, "page_005.png")
	writeTestPNG(t, imagePath, 100, 100)

	result, err := ocrpipeline.RunStructuredPage(context.Background(), ocrpipeline.StructuredOCRInput{
		BookID: "plugin-book", RunID: "test-run", PageNumber: 5, ImagePath: imagePath, WorkDir: workDir,
	}, client)
	require.NoError(t, err)

	rendered, err := os.ReadFile(result.RenderedMD)
	require.NoError(t, err)
	require.Contains(t, string(rendered), "Plugin OCR output for page 5.")
	require.Contains(t, string(rendered), "<!-- page:005 -->")

	// The full page artifact contract must hold on the plugin path too.
	for _, name := range []string{"01-turn-input.yaml", "02-turn-final.yaml", "03-raw-response.json", "04-structured.json", "05-rendered.md", "06-validation.json"} {
		info, err := os.Stat(filepath.Join(result.PageDir, name))
		require.NoError(t, err, name)
		require.NotZero(t, info.Size(), name)
	}

	// The audit turn records the delegation and exactly one image.
	turnInput, err := os.ReadFile(filepath.Join(result.PageDir, "01-turn-input.yaml"))
	require.NoError(t, err)
	require.Contains(t, string(turnInput), "delegated to plugin")
}

func TestPluginOCRPageEnforcesPageNumberGate(t *testing.T) {
	requirePython3(t)
	mgr := newTestManager(t, []Spec{{ID: "test", Path: testPluginPath(t, "test_plugin.py"), Seams: []string{OpOCRPage}}})
	client := NewStructuredOCRClient(mgr, nil)

	workDir := t.TempDir()
	imagePath := filepath.Join(workDir, "page_005.png")
	writeTestPNG(t, imagePath, 50, 50)
	// The "wrong-page-book" book ID makes the test plugin answer for page 6
	// when asked for page 5; the host-side gate must reject it.
	_, err := ocrpipeline.RunStructuredPage(context.Background(), ocrpipeline.StructuredOCRInput{
		BookID: "wrong-page-book", RunID: "test-run", PageNumber: 5, ImagePath: imagePath, WorkDir: workDir,
	}, client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "page mismatch")
}

func TestPluginOCRFallsBackWhenSeamUnbound(t *testing.T) {
	requirePython3(t)
	mgr := newTestManager(t, []Spec{{ID: "prompt-only", Path: testPluginPath(t, "prompt_only_plugin.py"), Seams: []string{OpPromptRender}}})
	client := NewStructuredOCRClient(mgr, ocrpipeline.DryRunStructuredOCRClient{})

	workDir := t.TempDir()
	imagePath := filepath.Join(workDir, "page_012.png")
	writeTestPNG(t, imagePath, 50, 50)
	result, err := ocrpipeline.RunStructuredPage(context.Background(), ocrpipeline.StructuredOCRInput{
		BookID: "plugin-book", RunID: "test-run", PageNumber: 12, ImagePath: imagePath, WorkDir: workDir,
	}, client)
	require.NoError(t, err)
	rendered, err := os.ReadFile(result.RenderedMD)
	require.NoError(t, err)
	// Dry-run fixture content, not plugin content.
	require.Contains(t, string(rendered), "rudimentary user interface")
}

func TestPluginPromptRendererAppendsSchemaContract(t *testing.T) {
	requirePython3(t)
	mgr := newTestManager(t, []Spec{{ID: "test", Path: testPluginPath(t, "test_plugin.py"), Seams: []string{OpPromptRender}}})
	renderer := NewPromptRenderer(mgr)

	system, user, err := renderer.RenderPrompts(ocrpipeline.StructuredOCRInput{BookID: "b", PageNumber: 3})
	require.NoError(t, err)
	require.Equal(t, "TEST SYSTEM PROMPT", system)
	require.True(t, strings.HasPrefix(user, "TEST USER PROMPT for page 3"))
	// The host-owned contract must survive any plugin prompt.
	require.Contains(t, user, "Return only strict JSON")
	require.Contains(t, user, "page_number: 003")
}

func TestPluginFigureSegmenterDrivesFigureExtraction(t *testing.T) {
	requirePython3(t)
	mgr := newTestManager(t, []Spec{{ID: "test", Path: testPluginPath(t, "test_plugin.py"), Seams: []string{OpFiguresSegment}}})
	segmenter := NewFigureSegmenter(mgr)

	imageDir := t.TempDir()
	outputDir := t.TempDir()
	writeTestPNG(t, filepath.Join(imageDir, "page_001.png"), 200, 200)

	markdown := "<!-- page:001 -->\n\n[FIGURE: A diagram of the test fixture]\n"
	_, figures, err := ocrquality.EmbedExtractedFigures(markdown, ocrquality.FigureExtractionOptions{ImageDir: imageDir, OutputDir: outputDir, Segmenter: segmenter})
	require.NoError(t, err)
	require.Len(t, figures, 1)
	require.Equal(t, "test-seg-v1", figures[0].Method)
	require.Equal(t, 10, figures[0].CropRect.X)
	require.Equal(t, 20, figures[0].CropRect.Y)
	require.Equal(t, 30, figures[0].CropRect.Width)
	require.Equal(t, 40, figures[0].CropRect.Height)
	_, err = os.Stat(figures[0].ImagePath)
	require.NoError(t, err)
}
