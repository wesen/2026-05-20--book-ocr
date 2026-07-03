package ocrpipeline

import (
	"github.com/go-go-golems/book-ocr/internal/bookprofile"
)

// PolicyFromProfile compiles a book profile into the prompt and render policy
// the structured pipeline consumes. The profile is authoritative: fields the
// profile leaves empty become generic behavior (plain code fences, no figure
// suppression cues), NOT the Report-794 defaults — a profile-driven run is
// exactly as book-specific as its profile says.
func PolicyFromProfile(profile bookprofile.Profile) (*PromptSpec, *RenderOptions) {
	prompt := &PromptSpec{
		PreserveTerms: append([]string(nil), profile.Prompt.PreserveTerms...),
		LanguageNote:  profile.Code.PromptNote,
		Example:       profile.Prompt.Example,
	}
	wrap := profile.Render.WrapWidth
	if wrap <= 0 {
		wrap = 88
	}
	render := &RenderOptions{
		WrapWidth:                 wrap,
		IncludeFooters:            profile.Render.IncludeFooters,
		CodeLanguage:              profile.Code.DefaultLanguage,
		SuppressTextualFigureCues: append([]string(nil), profile.Render.SuppressTextualFigureCues...),
		SuppressTableFigureCues:   append([]string(nil), profile.Render.SuppressTableFigureCues...),
		EnableBoxedSetFallback:    profile.Render.EnableBoxedSetFallback,
	}
	return prompt, render
}
