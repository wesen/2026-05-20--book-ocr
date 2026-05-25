package vlmseparation

import "time"

const (
	ScenarioTargetOnly                  = "target-only"
	ScenarioSingleBlockTargetFirst      = "single-block-target-first"
	ScenarioSingleBlockLabeledImages    = "single-block-labeled-images"
	ScenarioMultiBlockLabeled           = "multi-block-labeled"
	ScenarioContextFirstNegativeControl = "context-first-negative-control"
	ScenarioTargetPlusTextContext       = "target-plus-text-context"
)

type Scenario struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Config struct {
	BookID            string
	ImageDir          string
	TargetPages       []int
	Scenarios         []string
	Preset            string
	OutDir            string
	SQLitePath        string
	TurnsDB           string
	TurnsDSN          string
	Profile           string
	ProfileRegistries []string
	DryRun            bool
	MaxTrials         int
	RunID             string
}

type RunManifest struct {
	ID                string     `json:"id"`
	BookID            string     `json:"book_id"`
	ImageDir          string     `json:"image_dir"`
	TargetPages       []int      `json:"target_pages"`
	Scenarios         []Scenario `json:"scenarios"`
	Profile           string     `json:"profile,omitempty"`
	ProfileRegistries []string   `json:"profile_registries,omitempty"`
	DryRun            bool       `json:"dry_run"`
	OutDir            string     `json:"out_dir"`
	SQLitePath        string     `json:"sqlite_path"`
	TurnsDB           string     `json:"turns_db,omitempty"`
	TurnsDSN          string     `json:"turns_dsn,omitempty"`
	StartedAt         time.Time  `json:"started_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
}

type Trial struct {
	ID           string   `json:"id"`
	RunID        string   `json:"run_id"`
	Scenario     Scenario `json:"scenario"`
	TargetPage   int      `json:"target_page"`
	PreviousPage int      `json:"previous_page,omitempty"`
	NextPage     int      `json:"next_page,omitempty"`
}

type TrialInput struct {
	TrialID           string
	RunID             string
	BookID            string
	Scenario          Scenario
	TargetPage        int
	PreviousPage      int
	NextPage          int
	TargetImagePath   string
	PreviousImagePath string
	NextImagePath     string
	PreviousText      string
	NextText          string
	SessionID         string
	TurnID            string
	Oracle            PageOracle
}

type ContentMarkers struct {
	FigureCaptions  []string `json:"figure_captions,omitempty"`
	SectionHeadings []string `json:"section_headings,omitempty"`
	UniquePhrases   []string `json:"unique_phrases,omitempty"`
}

type PageIdentity struct {
	PageNumberVisible   string   `json:"page_number_visible,omitempty"`
	TitleOrCaptionLines []string `json:"title_or_caption_lines,omitempty"`
}

type BenchmarkResponse struct {
	TargetPage              int            `json:"target_page"`
	TranscribedPageIdentity PageIdentity   `json:"transcribed_page_identity"`
	ContentMarkers          ContentMarkers `json:"content_markers"`
	Transcription           string         `json:"transcription"`
	SuspectedContextCopy    bool           `json:"suspected_context_copy"`
	Notes                   []string       `json:"notes,omitempty"`
}

type TrialMetrics struct {
	JSONParseOK          bool    `json:"json_parse_ok"`
	JSONSanitized        bool    `json:"json_sanitized,omitempty"`
	SchemaRepaired       bool    `json:"schema_repaired,omitempty"`
	ParseStrategy        string  `json:"parse_strategy,omitempty"`
	ExpectedPhraseHits   int     `json:"expected_phrase_hits"`
	ExpectedPhraseTotal  int     `json:"expected_phrase_total"`
	ForbiddenPhraseHits  int     `json:"forbidden_phrase_hits"`
	ForbiddenCaptionHits int     `json:"forbidden_caption_hits"`
	SuspectedBleed       bool    `json:"suspected_bleed"`
	TargetOnlyScore      float64 `json:"target_only_score"`
}

type TrialResult struct {
	ID                 string             `json:"id"`
	RunID              string             `json:"run_id"`
	Scenario           string             `json:"scenario"`
	TargetPage         int                `json:"target_page"`
	PreviousPage       int                `json:"previous_page,omitempty"`
	NextPage           int                `json:"next_page,omitempty"`
	Status             string             `json:"status"`
	RequestPath        string             `json:"request_path,omitempty"`
	ResponsePath       string             `json:"response_path,omitempty"`
	ParsedResponsePath string             `json:"parsed_response_path,omitempty"`
	MetricsPath        string             `json:"metrics_path,omitempty"`
	TurnSessionID      string             `json:"turn_session_id"`
	TurnID             string             `json:"turn_id"`
	StartedAt          time.Time          `json:"started_at"`
	CompletedAt        time.Time          `json:"completed_at"`
	LatencyMS          int64              `json:"latency_ms"`
	RawResponse        string             `json:"raw_response,omitempty"`
	Parsed             *BenchmarkResponse `json:"parsed,omitempty"`
	Metrics            TrialMetrics       `json:"metrics"`
	Error              string             `json:"error,omitempty"`
}

type RunResult struct {
	Manifest RunManifest   `json:"manifest"`
	Trials   []TrialResult `json:"trials"`
	Summary  Summary       `json:"summary"`
}

type Summary struct {
	RunID                  string             `json:"run_id"`
	TrialCount             int                `json:"trial_count"`
	Succeeded              int                `json:"succeeded"`
	Failed                 int                `json:"failed"`
	SuspectedBleed         int                `json:"suspected_bleed"`
	AverageTargetOnlyScore float64            `json:"average_target_only_score"`
	ByScenario             map[string]Summary `json:"by_scenario,omitempty"`
}

type PageOracle struct {
	TargetPage        int      `json:"target_page"`
	ExpectedPhrases   []string `json:"expected_phrases,omitempty"`
	ForbiddenPhrases  []string `json:"forbidden_phrases,omitempty"`
	ExpectedCaptions  []string `json:"expected_captions,omitempty"`
	ForbiddenCaptions []string `json:"forbidden_captions,omitempty"`
}
