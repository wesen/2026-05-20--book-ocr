package ocrquality

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writePagePNG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 120, 120))
	for y := 0; y < 120; y++ {
		for x := 0; x < 120; x++ {
			img.Set(x, y, color.White)
		}
	}
	for y := 40; y < 80; y++ {
		for x := 30; x < 90; x++ {
			img.Set(x, y, color.Black)
		}
	}
	f, err := os.Create(path)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	require.NoError(t, png.Encode(f, img))
}

func TestIndexPageImagesHandlesAnyZeroPaddingWidth(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"page_001.png", "page_0012.png", "page_1000.png"} {
		writePagePNG(t, filepath.Join(dir, name))
	}
	index := indexPageImages(dir, "")
	require.Equal(t, filepath.Join(dir, "page_001.png"), index[1])
	require.Equal(t, filepath.Join(dir, "page_0012.png"), index[12])
	require.Equal(t, filepath.Join(dir, "page_1000.png"), index[1000])
}

func TestSplitPagesAcceptsWideMarkers(t *testing.T) {
	md := "<!-- page:001 -->\n\nfirst\n\n<!-- page:1000 -->\n\nthousandth\n"
	pages := SplitPages(md)
	require.Len(t, pages, 2)
	require.Equal(t, 1, pages[0].Number)
	require.Equal(t, 1000, pages[1].Number)
}

// The finding-F4 regression: ingest writes page_0001.png (4 digits) while the
// figure extractor used to construct page_%03d.png and miss the file.
func TestEmbedExtractedFiguresFindsFourDigitPageImages(t *testing.T) {
	imageDir := t.TempDir()
	outputDir := t.TempDir()
	writePagePNG(t, filepath.Join(imageDir, "page_0001.png"))

	markdown := "<!-- page:001 -->\n\n[FIGURE: A diagram]\n"
	_, figures, err := EmbedExtractedFigures(markdown, FigureExtractionOptions{ImageDir: imageDir, OutputDir: outputDir})
	require.NoError(t, err)
	require.Len(t, figures, 1)
	_, err = os.Stat(figures[0].ImagePath)
	require.NoError(t, err)
}
