package ocrvalidation

type Warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Page    int    `json:"page,omitempty"`
	Other   int    `json:"other_page,omitempty"`
	Value   string `json:"value,omitempty"`
}

type PageText struct {
	Number   int
	Markdown string
}

type AnchorOracle struct {
	TargetPage        int      `json:"target_page"`
	ExpectedPhrases   []string `json:"expected_phrases,omitempty"`
	ForbiddenPhrases  []string `json:"forbidden_phrases,omitempty"`
	ExpectedCaptions  []string `json:"expected_captions,omitempty"`
	ForbiddenCaptions []string `json:"forbidden_captions,omitempty"`
}

type AnchorResult struct {
	TargetPage           int      `json:"target_page"`
	ExpectedHits         int      `json:"expected_hits"`
	ExpectedTotal        int      `json:"expected_total"`
	ForbiddenHits        int      `json:"forbidden_hits"`
	ForbiddenCaptionHits int      `json:"forbidden_caption_hits"`
	MissingExpected      []string `json:"missing_expected,omitempty"`
	MatchedForbidden     []string `json:"matched_forbidden,omitempty"`
	SuspectedBleed       bool     `json:"suspected_bleed"`
}
