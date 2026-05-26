package ocrpipeline

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestBuildStructuredOCRInputTurnContainsExactlyOneTargetImage(t *testing.T) {
	turn, err := BuildStructuredOCRInputTurn(StructuredOCRInput{BookID: "report-794", PageNumber: 32, ImagePath: "page_032.png"}, []byte("fake-png"))
	require.NoError(t, err)
	require.Equal(t, PageTurnID(32, 1, "structured-ocr"), turn.ID)
	require.Equal(t, 1, CountTurnImages(turn))
}

func TestRunStructuredPageDryRunWritesArtifactsAndMarkdownTable(t *testing.T) {
	ctx := context.Background()
	workDir := t.TempDir()
	imagePath := filepath.Join(workDir, "page_032.png")
	require.NoError(t, os.WriteFile(imagePath, []byte("fake-png"), 0o644))

	result, err := RunStructuredPage(ctx, StructuredOCRInput{BookID: "report-794", RunID: "test-run", PageNumber: 32, ImagePath: imagePath, WorkDir: workDir, DryRun: true}, DryRunStructuredOCRClient{})
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(workDir, "pages/page_032/01-turn-input.yaml"))
	require.FileExists(t, filepath.Join(workDir, "pages/page_032/02-turn-final.yaml"))
	require.FileExists(t, result.RawResponse)
	require.FileExists(t, result.StructuredJSON)
	require.FileExists(t, result.RenderedMD)
	require.FileExists(t, result.ValidationJSON)

	md, err := os.ReadFile(result.RenderedMD)
	require.NoError(t, err)
	require.Contains(t, string(md), "<!-- page:032 -->")
	require.Contains(t, string(md), "Figure 2-2: PPSCalc -- Formula Display")
	require.Contains(t, string(md), "|  | A | B | C |")
	require.Contains(t, string(md), "| 1 | 100 | 20 | A1*B1 |")
	require.Contains(t, string(md), "Figure 2-3: PPSCalc -- Value Display")
	require.Contains(t, string(md), "| 3 |  |  | 2375 |")
	require.Empty(t, result.Validation.Warnings)

	db, err := sql.Open("sqlite3", filepath.Join(workDir, "turns.db"))
	require.NoError(t, err)
	defer db.Close()
	var turnsCount int
	require.NoError(t, db.QueryRow(`select count(*) from turns`).Scan(&turnsCount))
	require.Equal(t, 1, turnsCount)
	var phases string
	require.NoError(t, db.QueryRow(`select group_concat(distinct phase) from turn_block_membership`).Scan(&phases))
	require.Contains(t, phases, "input")
	require.Contains(t, phases, "final")
}

func TestParseStructuredOCRResponseValidatesRequiredFields(t *testing.T) {
	_, err := ParseStructuredOCRResponse(`{}`)
	require.Error(t, err)
	page, err := ParseStructuredOCRResponse(`{"schema_version":"structured-ocr/v1","book_id":"report-794","page_number":32,"page_type":"body","blocks":[]}`)
	require.NoError(t, err)
	require.Equal(t, 32, page.PageNumber)
}

func TestValidateStructuredPageWarnsForMissingFigureCaption(t *testing.T) {
	page := StructuredPageOCR{SchemaVersion: StructuredOCRSchemaVersion, BookID: "report-794", PageNumber: 32, PageType: PageTypeFigure, Blocks: []OCRBlock{{ID: "p032-f001", Type: BlockFigure}}}
	validation := ValidateStructuredPage(page, RenderPageMarkdown(page, nil, DefaultRenderOptions()))
	require.Len(t, validation.Warnings, 1)
	require.Equal(t, "figure_missing_caption", validation.Warnings[0].Code)
}

func TestParseStructuredOCRResponseRepairsCommonLiveShape(t *testing.T) {
	raw := `{
  "schema_version": "structured-ocr/v1",
  "book_id": "report-794",
  "page_number": "032",
  "page_type": "content",
  "blocks": [
    {
      "type": "figure",
      "caption": "Figure 2-2: PPSCalc -- Formula Display",
      "diagram_text": "Columns: A | B | C\nRow 1: A1 = 100"
    }
  ]
}`
	page, err := ParseStructuredOCRResponse(raw)
	require.NoError(t, err)
	require.Equal(t, 32, page.PageNumber)
	require.Equal(t, "p032-b001", page.Blocks[0].ID)
	require.Equal(t, []string{"Columns: A | B | C", "Row 1: A1 = 100"}, page.Blocks[0].DiagramText)
}

func TestParseStructuredOCRResponseAcceptsStringListItems(t *testing.T) {
	raw := `{"schema_version":"structured-ocr/v1","book_id":"report-794","page_number":6,"page_type":"contents","blocks":[{"id":"p006-list001","type":"list","items":["1.1 The Primitive Presentation System Model .... 9"]}]}`
	page, err := ParseStructuredOCRResponse(raw)
	require.NoError(t, err)
	require.Equal(t, "1.1 The Primitive Presentation System Model .... 9", page.Blocks[0].Items[0].Text)
}

func TestParseStructuredOCRResponseRepairsNestedFigureAndHeadingCaption(t *testing.T) {
	raw := `{
  "schema_version": "structured-ocr/v1",
  "book_id": "report-794",
  "page_number": "042",
  "page_type": "figure",
  "blocks": [
    {"id":"p042-h001","type":"heading","level":2,"text":"Figure 2-9: Presenter Parts"},
    {"id":"p042-f001","type":"figure","figure":{"caption":"Figure 2-9: Presenter Parts","description":"Presenter diagram","diagram_text":["Presentation Editor"]}}
  ]
}`
	page, err := ParseStructuredOCRResponse(raw)
	require.NoError(t, err)
	require.Len(t, page.Blocks, 1)
	require.Equal(t, "Figure 2-9: Presenter Parts", page.Blocks[0].Caption)
	require.Equal(t, "Presenter diagram", page.Blocks[0].Description)
	require.Equal(t, []string{"Presentation Editor"}, page.Blocks[0].DiagramText)
}

func TestParseStructuredOCRResponseRepairsFigureCaptionFromHeading(t *testing.T) {
	raw := `{
  "schema_version": "structured-ocr/v1",
  "book_id": "report-794",
  "page_number": 115,
  "page_type": "figure",
  "blocks": [
    {"id":"p115-h001","type":"heading","level":1,"text":"Figure 5-7: Reference Resolution"},
    {"id":"p115-f001","type":"figure","caption":"","description":"Reference resolution diagram"}
  ]
}`
	page, err := ParseStructuredOCRResponse(raw)
	require.NoError(t, err)
	require.Len(t, page.Blocks, 1)
	require.Equal(t, "Figure 5-7: Reference Resolution", page.Blocks[0].Caption)
}
