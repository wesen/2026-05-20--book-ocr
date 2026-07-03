package bookprofile

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var updateProfiles = flag.Bool("update-profiles", false, "rewrite profiles/report-794.yaml from the built-in Report794() profile")

// TestReport794ProfileYAMLInSync pins profiles/report-794.yaml to the
// built-in Report794() profile: the YAML file is the product-facing form and
// the Go function is the regression fixture; they must never drift.
// Regenerate with: go test ./internal/bookprofile -run ProfileYAML -update-profiles
func TestReport794ProfileYAMLInSync(t *testing.T) {
	path := filepath.Join("..", "..", "profiles", "report-794.yaml")
	if *updateProfiles {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, Save(path, Report794()))
		return
	}
	loaded, err := Load(path)
	require.NoError(t, err, "profiles/report-794.yaml missing; run with -update-profiles to create it")
	require.Equal(t, Report794(), loaded)
}
