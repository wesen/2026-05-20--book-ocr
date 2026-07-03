package ocrpipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestStructuredOCRPromptGolden pins the exact prompt text sent for the
// Report-794 default policy. The profile-driven prompt generalization must
// keep this byte-identical: prompt phrasing changes model behavior, so any
// diff here needs an A/B rerun of the oracle pages before merging.
// Regenerate with: go test ./internal/ocrpipeline -run PromptGolden -update
func TestStructuredOCRPromptGolden(t *testing.T) {
	input := StructuredOCRInput{BookID: "report-794", PageNumber: 32}
	got := StructuredOCRSystemPrompt + "\n=====\n" + RenderStructuredOCRPrompt(input)
	goldenPath := filepath.Join("testdata", "golden", "prompt-report-794.golden.txt")
	if *updateGolden {
		require.NoError(t, os.WriteFile(goldenPath, []byte(got), 0o644))
		return
	}
	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "golden file missing; run with -update to create it")
	require.Equal(t, string(want), got)
}
