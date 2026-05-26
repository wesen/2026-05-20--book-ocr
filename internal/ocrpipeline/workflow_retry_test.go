package ocrpipeline

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"github.com/go-go-golems/scraper/pkg/engine/model"
	"github.com/go-go-golems/scraper/pkg/workflow"
)

type failOnceStructuredClient struct {
	mu    sync.Mutex
	calls int
}

func (c *failOnceStructuredClient) OCRPage(ctx context.Context, input StructuredOCRInput, imageBytes []byte) (StructuredOCRResult, error) {
	c.mu.Lock()
	c.calls++
	call := c.calls
	c.mu.Unlock()
	if call == 1 {
		return StructuredOCRResult{}, fmt.Errorf("run structured OCR inference: transient test failure")
	}
	return DryRunStructuredOCRClient{}.OCRPage(ctx, input, imageBytes)
}

func (c *failOnceStructuredClient) Calls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

func TestStructuredWorkflowRetriesTransientPageFailure(t *testing.T) {
	ctx := context.Background()
	workDir := t.TempDir()
	imageDir := filepath.Join(workDir, "images")
	require.NoError(t, os.MkdirAll(imageDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(imageDir, "page_001.png"), []byte("fake-png"), 0o644))

	client := &failOnceStructuredClient{}
	rt, err := workflow.NewRuntime(ctx, workflow.Config{
		Store:           workflow.SQLiteStore(filepath.Join(workDir, "engine.db")),
		ArtifactStore:   workflow.NewFileArtifactStore(filepath.Join(workDir, "artifacts")),
		ProjectionStore: workflow.NewSQLiteProjectionStore(filepath.Join(workDir, "projections")),
		WorkerID:        "structured-retry-test",
		MaxWorkers:      1,
		PollInterval:    25 * time.Millisecond,
		Queues: map[model.QueueKey]workflow.QueueConfig{
			QueueStructuredControl:    {MaxWorkers: 1},
			QueueStructuredVision:     {MaxWorkers: 1},
			QueueStructuredAssemble:   {MaxWorkers: 1},
			QueueStructuredValidation: {MaxWorkers: 1},
		},
	})
	require.NoError(t, err)
	defer func() { _ = rt.Close() }()
	require.NoError(t, RegisterStructuredWorkflow(rt, StructuredWorkflowConfig{Client: client}))

	handle, err := rt.StartRun(ctx, StructuredPackageName, StructuredRunInput{BookID: "retry-test", ImageDir: imageDir, WorkDir: workDir, DryRun: true, ExpectedPages: 1})
	require.NoError(t, err)

	deadline := time.Now().Add(10 * time.Second)
	for {
		_, err := rt.RunOnce(ctx)
		require.NoError(t, err)
		wf, err := rt.Workflow(ctx, handle.ID)
		require.NoError(t, err)
		if wf.Status == model.WorkflowStatusSucceeded {
			break
		}
		if wf.Status == model.WorkflowStatusFailed || wf.Status == model.WorkflowStatusCanceled {
			t.Fatalf("workflow finished with status %s", wf.Status)
		}
		if time.Now().After(deadline) {
			t.Fatalf("workflow did not finish before deadline; status=%s calls=%d", wf.Status, client.Calls())
		}
		time.Sleep(50 * time.Millisecond)
	}

	require.GreaterOrEqual(t, client.Calls(), 2)
	require.FileExists(t, filepath.Join(workDir, "assembled.md"))

	db, err := sql.Open("sqlite3", filepath.Join(workDir, "projections", StructuredProjectionName+".db"))
	require.NoError(t, err)
	defer db.Close()
	var status string
	require.NoError(t, db.QueryRow(`select status from structured_pages where book_id=? and page_num=1`, "retry-test").Scan(&status))
	require.Equal(t, "succeeded", status)
}
