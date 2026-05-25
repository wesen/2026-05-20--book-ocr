package vlmseparation

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/geppetto/pkg/turns"
	"github.com/stretchr/testify/require"
)

func TestBuildTurnForScenarioLayouts(t *testing.T) {
	dir := t.TempDir()
	for _, page := range []int{11, 12, 13} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "page_"+formatPage(page)+".png"), []byte("not-real-png-but-readable"), 0o644))
	}
	base := TrialInput{TargetPage: 12, PreviousPage: 11, NextPage: 13, TargetImagePath: filepath.Join(dir, "page_012.png"), PreviousImagePath: filepath.Join(dir, "page_011.png"), NextImagePath: filepath.Join(dir, "page_013.png"), SessionID: "page:012", TurnID: "trial:1"}

	base.Scenario = Scenario{Name: ScenarioTargetOnly}
	turn, err := BuildTurnForScenario(base)
	require.NoError(t, err)
	require.Equal(t, 2, len(turn.Blocks))
	require.Equal(t, 1, imageCount(turn))

	base.Scenario = Scenario{Name: ScenarioSingleBlockTargetFirst}
	turn, err = BuildTurnForScenario(base)
	require.NoError(t, err)
	require.Equal(t, 2, len(turn.Blocks))
	require.Equal(t, 3, imageCount(turn))

	base.Scenario = Scenario{Name: ScenarioMultiBlockLabeled}
	turn, err = BuildTurnForScenario(base)
	require.NoError(t, err)
	require.Greater(t, len(turn.Blocks), 4)
	require.Equal(t, 3, imageCount(turn))
}

func imageCount(turn *turns.Turn) int {
	count := 0
	for _, block := range turn.Blocks {
		images, ok := block.Payload[turns.PayloadKeyImages].([]map[string]any)
		if ok {
			count += len(images)
		}
	}
	return count
}

func formatPage(page int) string {
	return fmt.Sprintf("%03d", page)
}
