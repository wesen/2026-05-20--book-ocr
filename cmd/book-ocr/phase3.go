package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-go-golems/book-ocr/internal/bookprofile"
	"github.com/go-go-golems/book-ocr/internal/ocrmvp"
)

// ingestManifest records how a page-image directory was produced so later
// runs (and book-ocr init, eventually) can validate their input.
type ingestManifest struct {
	Source     string `json:"source"`
	SourceSHA  string `json:"source_sha256"`
	DPI        int    `json:"dpi"`
	Grayscale  bool   `json:"grayscale"`
	PageCount  int    `json:"page_count"`
	PagePrefix string `json:"page_prefix"`
	CreatedAt  string `json:"created_at"`
	Tool       string `json:"tool"`
}

var pdftoppmPagePattern = regexp.MustCompile(`^page-(\d+)\.png$`)

// runIngest rasterizes a PDF into the page_NNNN.png layout the pipeline
// expects, via poppler's pdftoppm (checked at startup, same pattern as the
// pandoc dependency on the output side).
func runIngest(args []string) error {
	fs := flag.NewFlagSet("ingest", flag.ContinueOnError)
	pdfPath := fs.String("pdf", "", "Source PDF to rasterize")
	outDir := fs.String("out-dir", "", "Output directory for page images")
	dpi := fs.Int("dpi", 300, "Rasterization resolution")
	grayscale := fs.Bool("grayscale", false, "Render grayscale instead of RGB")
	pdftoppmPath := fs.String("pdftoppm-path", "pdftoppm", "pdftoppm executable")
	logLevel := fs.String("log-level", "info", "zerolog level")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := configureLogLevel(*logLevel); err != nil {
		return err
	}
	if strings.TrimSpace(*pdfPath) == "" || strings.TrimSpace(*outDir) == "" {
		return fmt.Errorf("--pdf and --out-dir are required")
	}
	manifest, err := ingestPDF(*pdfPath, *outDir, *dpi, *grayscale, *pdftoppmPath)
	if err != nil {
		return err
	}
	fmt.Printf("ingested %d pages at %d dpi into %s (page_0001.png …)\n", manifest.PageCount, *dpi, *outDir)
	fmt.Printf("next: book-ocr structured-run --image-dir %s --page-glob 'page_*.png' ...\n", *outDir)
	return nil
}

// ingestPDF rasterizes a PDF into outDir as page_NNNN.png files and writes an
// ingest-manifest.json recording the provenance.
func ingestPDF(pdfPath string, outDir string, dpi int, grayscale bool, pdftoppmPath string) (ingestManifest, error) {
	tool, err := exec.LookPath(pdftoppmPath)
	if err != nil {
		return ingestManifest{}, fmt.Errorf("pdftoppm not found (%q): install poppler-utils or pass --pdftoppm-path", pdftoppmPath)
	}
	absPDF, err := filepath.Abs(pdfPath)
	if err != nil {
		return ingestManifest{}, err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return ingestManifest{}, err
	}

	cmdArgs := []string{"-r", strconv.Itoa(dpi), "-png"}
	if grayscale {
		cmdArgs = append(cmdArgs, "-gray")
	}
	cmdArgs = append(cmdArgs, absPDF, filepath.Join(outDir, "page"))
	cmd := exec.Command(tool, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return ingestManifest{}, fmt.Errorf("pdftoppm failed: %w", err)
	}

	// pdftoppm writes page-N.png with padding derived from the page count;
	// normalize to the pipeline's page_NNNN.png convention (4 digits: books
	// beyond 999 pages exist, and page-number inference tolerates any width).
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return ingestManifest{}, err
	}
	pageCount := 0
	for _, entry := range entries {
		match := pdftoppmPagePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		num, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}
		target := filepath.Join(outDir, fmt.Sprintf("page_%04d.png", num))
		if err := os.Rename(filepath.Join(outDir, entry.Name()), target); err != nil {
			return ingestManifest{}, err
		}
		pageCount++
	}
	if pageCount == 0 {
		return ingestManifest{}, fmt.Errorf("pdftoppm produced no page images in %s", outDir)
	}

	sha, err := fileSHA256(absPDF)
	if err != nil {
		return ingestManifest{}, err
	}
	manifest := ingestManifest{Source: absPDF, SourceSHA: sha, DPI: dpi, Grayscale: grayscale, PageCount: pageCount, PagePrefix: "page_", CreatedAt: time.Now().UTC().Format(time.RFC3339), Tool: tool}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return ingestManifest{}, err
	}
	if err := os.WriteFile(filepath.Join(outDir, "ingest-manifest.json"), append(manifestBytes, '\n'), 0o644); err != nil {
		return ingestManifest{}, err
	}
	return manifest, nil
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// runReport summarizes a run's projection and turn store: page status
// counts, block/warning totals, warning codes, and model-call counts.
func runReport(args []string) error {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	workDir := fs.String("work-dir", defaultWorkDir, "Run work directory")
	bookID := fs.String("book-id", "", "Optional book filter")
	logLevel := fs.String("log-level", "warn", "zerolog level")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := configureLogLevel(*logLevel); err != nil {
		return err
	}
	paths, err := resolveWorkDir(*workDir, false)
	if err != nil {
		return err
	}

	projPath := filepath.Join(paths.projectionsDir, "book_ocr_structured.db")
	proj, err := sql.Open("sqlite3", projPath)
	if err != nil {
		return err
	}
	defer func() { _ = proj.Close() }()

	where, params := "", []any{}
	if strings.TrimSpace(*bookID) != "" {
		where, params = " WHERE book_id = ?", []any{*bookID}
	}

	fmt.Printf("run report for %s\n\n", paths.root)
	fmt.Println("pages by status:")
	if err := printGroupCounts(proj, "SELECT status, COUNT(*) FROM structured_pages"+where+" GROUP BY status ORDER BY status", params); err != nil {
		return fmt.Errorf("query structured_pages (no structured run in this work dir?): %w", err)
	}

	var pages, warnings, tables, figures, renderedBytes sql.NullInt64
	if err := proj.QueryRow("SELECT COUNT(*), SUM(warning_count), SUM(table_count), SUM(figure_count), SUM(rendered_bytes) FROM structured_pages"+where, params...).Scan(&pages, &warnings, &tables, &figures, &renderedBytes); err != nil {
		return err
	}
	fmt.Printf("\ntotals: pages=%d warnings=%d tables=%d figures=%d rendered_bytes=%d\n",
		pages.Int64, warnings.Int64, tables.Int64, figures.Int64, renderedBytes.Int64)

	fmt.Println("\nwarnings by code:")
	if err := printGroupCounts(proj, "SELECT code, COUNT(*) FROM structured_warnings"+strings.Replace(where, "book_id", "book_id", 1)+" GROUP BY code ORDER BY COUNT(*) DESC", params); err != nil {
		fmt.Println("  (none)")
	}

	turnsPath := filepath.Join(paths.root, "turns.db")
	if _, err := os.Stat(turnsPath); err == nil {
		turnsDB, err := sql.Open("sqlite3", turnsPath)
		if err == nil {
			defer func() { _ = turnsDB.Close() }()
			var turnCount int
			if err := turnsDB.QueryRow("SELECT COUNT(*) FROM turns").Scan(&turnCount); err == nil {
				fmt.Printf("\nmodel calls: %d persisted turns (turns.db)\n", turnCount)
			}
		}
	}
	return nil
}

func printGroupCounts(db *sql.DB, query string, params []any) error {
	rows, err := db.Query(query, params...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	type kv struct {
		key   string
		count int64
	}
	var results []kv
	for rows.Next() {
		var r kv
		if err := rows.Scan(&r.key, &r.count); err != nil {
			return err
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	sort.SliceStable(results, func(i, j int) bool { return results[i].count > results[j].count })
	if len(results) == 0 {
		fmt.Println("  (none)")
		return nil
	}
	for _, r := range results {
		fmt.Printf("  %-32s %d\n", r.key, r.count)
	}
	return nil
}

// runInit bootstraps a new book workspace: optional PDF ingest, page
// discovery, and a drafted book profile the user reviews and edits. This is
// the "a new book is a YAML file, not a code change" entry point.
func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	bookID := fs.String("book-id", "", "Identifier for the new book (required)")
	pdfPath := fs.String("pdf", "", "Source PDF to ingest (alternative to --image-dir)")
	imageDir := fs.String("image-dir", "", "Directory of existing page images (alternative to --pdf)")
	outDir := fs.String("out-dir", "", "Workspace directory (default ./<book-id>)")
	dpi := fs.Int("dpi", 300, "Rasterization resolution when ingesting a PDF")
	grayscale := fs.Bool("grayscale", false, "Render grayscale when ingesting a PDF")
	pdftoppmPath := fs.String("pdftoppm-path", "pdftoppm", "pdftoppm executable")
	logLevel := fs.String("log-level", "info", "zerolog level")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := configureLogLevel(*logLevel); err != nil {
		return err
	}
	if strings.TrimSpace(*bookID) == "" {
		return fmt.Errorf("--book-id is required")
	}
	hasPDF := strings.TrimSpace(*pdfPath) != ""
	hasImages := strings.TrimSpace(*imageDir) != ""
	if hasPDF == hasImages {
		return fmt.Errorf("provide exactly one of --pdf or --image-dir")
	}
	workspace := strings.TrimSpace(*outDir)
	if workspace == "" {
		workspace = *bookID
	}
	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(absWorkspace, 0o755); err != nil {
		return err
	}

	pagesDir := *imageDir
	if hasPDF {
		pagesDir = filepath.Join(absWorkspace, "pages")
		manifest, err := ingestPDF(*pdfPath, pagesDir, *dpi, *grayscale, *pdftoppmPath)
		if err != nil {
			return err
		}
		fmt.Printf("ingested %d pages into %s\n", manifest.PageCount, pagesDir)
	}
	absPages, err := filepath.Abs(pagesDir)
	if err != nil {
		return err
	}
	pages, err := ocrmvp.DiscoverPageImages(ocrmvp.RunInput{BookID: *bookID, ImageDir: absPages, PageGlob: "page_*.png"})
	if err != nil {
		return err
	}
	if len(pages) == 0 {
		return fmt.Errorf("no page images matching page_*.png in %s", absPages)
	}

	profile := bookprofile.Profile{
		ID:          *bookID,
		DisplayName: *bookID,
		Family:      bookprofile.FamilyTechnicalReport,
		PageImages:  bookprofile.PageImagePolicy{Glob: "page_*.png", PageNumberRegex: `page_(\d+)\.png`},
		QA:          bookprofile.QAPolicy{ExpectedPages: len(pages)},
		Render:      bookprofile.RenderPolicy{WrapWidth: 88},
	}
	profilePath := filepath.Join(absWorkspace, *bookID+".profile.yaml")
	if err := bookprofile.Save(profilePath, profile); err != nil {
		return err
	}

	fmt.Printf("workspace: %s\n", absWorkspace)
	fmt.Printf("pages:     %d images in %s\n", len(pages), absPages)
	fmt.Printf("profile:   %s (drafted — review before a live run)\n\n", profilePath)
	fmt.Println("review checklist:")
	fmt.Println("  - qa.expected_pages: confirm against the physical book")
	fmt.Println("  - code.default_language / code.prompt_note: set if the book has code listings")
	fmt.Println("  - prompt.preserve_terms: proper nouns and historical spellings to transcribe exactly")
	fmt.Println("  - render.suppress_*_cues: only if figures duplicate transcribed tables/code")
	fmt.Println("  - plugins: bind ocr.page/prompt.render/figures.segment experiments (see design doc 02)")
	fmt.Println("\nnext:")
	fmt.Printf("  book-ocr structured-run --book-id %s --image-dir %s \\\n", *bookID, absPages)
	fmt.Printf("    --book-profile %s \\\n", profilePath)
	fmt.Printf("    --work-dir %s --dry-run --expected-pages %d   # offline check first\n", filepath.Join(absWorkspace, "run"), len(pages))
	return nil
}
