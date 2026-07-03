package plugin

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/go-go-golems/devctl/pkg/runtime"
	"github.com/stretchr/testify/require"
)

func requirePython3(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available")
	}
}

func testPluginPath(t *testing.T, name string) string {
	t.Helper()
	path := "testdata/" + name
	require.NoError(t, os.Chmod(path, 0o755))
	return path
}

func newTestManager(t *testing.T, specs []Spec) *Manager {
	t.Helper()
	mgr, err := NewManager(context.Background(), specs, runtime.RequestMeta{DryRun: true})
	require.NoError(t, err)
	t.Cleanup(func() { mgr.Close(context.Background()) })
	return mgr
}

func TestManagerBindsSeamsAndCalls(t *testing.T) {
	requirePython3(t)
	mgr := newTestManager(t, []Spec{{ID: "test", Path: testPluginPath(t, "test_plugin.py"), Seams: []string{OpOCRPage, OpPromptRender, OpFiguresSegment}}})

	require.True(t, mgr.Has(OpOCRPage))
	require.True(t, mgr.Has(OpPromptRender))
	require.True(t, mgr.Has(OpFiguresSegment))

	var out PromptRenderOutput
	require.NoError(t, mgr.Call(context.Background(), OpPromptRender, PromptRenderInput{OpSchema: "prompt.render/v1", BookID: "b", PageNumber: 7}, &out))
	require.Equal(t, "TEST SYSTEM PROMPT", out.System)
	require.Contains(t, out.User, "page 7")

	prov := mgr.Provenance()
	require.Len(t, prov, 1)
	require.Equal(t, "test-plugin", prov[0].Name)
	require.Contains(t, prov[0].Ops, OpOCRPage)
}

func TestManagerRejectsSeamNotAdvertisedInHandshake(t *testing.T) {
	requirePython3(t)
	_, err := NewManager(context.Background(), []Spec{{ID: "prompt-only", Path: testPluginPath(t, "prompt_only_plugin.py"), Seams: []string{OpOCRPage}}}, runtime.RequestMeta{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "claims seam")
	require.Contains(t, err.Error(), "ocr.page")
}

func TestManagerRejectsUnknownSeam(t *testing.T) {
	requirePython3(t)
	_, err := NewManager(context.Background(), []Spec{{ID: "test", Path: testPluginPath(t, "test_plugin.py"), Seams: []string{"workflow.hijack"}}}, runtime.RequestMeta{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown seam")
}

func TestManagerRejectsStdoutContamination(t *testing.T) {
	requirePython3(t)
	_, err := NewManager(context.Background(), []Spec{{ID: "bad", Path: testPluginPath(t, "bad_handshake.py"), Seams: []string{OpPromptRender}}}, runtime.RequestMeta{})
	require.Error(t, err)
}

func TestNilManagerHasNoSeams(t *testing.T) {
	var mgr *Manager
	require.False(t, mgr.Has(OpOCRPage))
	require.Nil(t, mgr.Provenance())
	mgr.Close(context.Background()) // must not panic
}

func TestParseSeamBinding(t *testing.T) {
	spec, err := ParseSeamBinding("ocr.page=./my_plugin.py")
	require.NoError(t, err)
	require.Equal(t, []string{OpOCRPage}, spec.Seams)
	require.Equal(t, "./my_plugin.py", spec.Path)

	_, err = ParseSeamBinding("nonsense")
	require.Error(t, err)

	_, err = ParseSeamBinding("bogus.seam=./x.py")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown plugin seam")
}
