package ocrpipeline

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var updateGolden = flag.Bool("update", false, "rewrite golden files under testdata/golden")

// TestRenderPageMarkdownGolden pins the exact Markdown produced for every
// block type and rendering heuristic. Each testdata/golden/NAME.json holds a
// StructuredPageOCR page; optional NAME.opts.json overrides RenderOptions and
// optional NAME.figures.json provides a FigureMap. The rendered output must
// match NAME.golden.md byte for byte, so any renderer change — including the
// profile-driven generalization — shows up as a reviewable diff.
// Regenerate with: go test ./internal/ocrpipeline -run Golden -update
func TestRenderPageMarkdownGolden(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("testdata", "golden", "*.json"))
	require.NoError(t, err)
	var fixtures []string
	for _, p := range paths {
		if strings.HasSuffix(p, ".opts.json") || strings.HasSuffix(p, ".figures.json") {
			continue
		}
		fixtures = append(fixtures, p)
	}
	require.NotEmpty(t, fixtures, "no golden fixtures found")

	for _, path := range fixtures {
		name := strings.TrimSuffix(filepath.Base(path), ".json")
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			var page StructuredPageOCR
			require.NoError(t, json.Unmarshal(data, &page))

			opts := DefaultRenderOptions()
			if optsData, err := os.ReadFile(strings.TrimSuffix(path, ".json") + ".opts.json"); err == nil {
				require.NoError(t, json.Unmarshal(optsData, &opts))
			}
			var figures FigureResolver
			if figData, err := os.ReadFile(strings.TrimSuffix(path, ".json") + ".figures.json"); err == nil {
				var m FigureMap
				require.NoError(t, json.Unmarshal(figData, &m))
				figures = m
			}

			got := RenderPageMarkdown(page, figures, opts)
			goldenPath := strings.TrimSuffix(path, ".json") + ".golden.md"
			if *updateGolden {
				require.NoError(t, os.WriteFile(goldenPath, []byte(got), 0o644))
				return
			}
			want, err := os.ReadFile(goldenPath)
			require.NoError(t, err, "golden file missing; run with -update to create it")
			require.Equal(t, string(want), got, "rendered markdown differs from %s", goldenPath)
		})
	}
}
