// Package web exposes the ASCII-art renderer through HTTP handlers.
package web

const (
	BannerStandard   = "standard"
	BannerShadow     = "shadow"
	BannerThinkertoy = "thinkertoy"
)

// BannerRegistry maps public banner identifiers to preloaded glyph maps.
// Its contents are read-only after application construction.
type BannerRegistry map[string]map[rune][]string

// FormState preserves raw, untrusted form values for redisplay.
type FormState struct {
	Text      string
	Banner    string
	UseColor  bool
	Color     string
	Substring string
	Alignment string
	Width     string
}

// SelectedColor is a validated web color in both hexadecimal and RGB forms.
type SelectedColor struct {
	Enabled bool
	Hex     string
	Red     int
	Green   int
	Blue    int
}

// GenerationInput contains validated, bounded values for generation.
type GenerationInput struct {
	Text      string
	Banner    string
	Color     SelectedColor
	Substring string
	Alignment string
	Width     int
}

// GeneratedASCII contains aligned ANSI text without HTML concerns.
type GeneratedASCII struct {
	ANSIText string
	Width    int
}

// StyledRun is a safe piece of text with its renderer-controlled color state.
type StyledRun struct {
	Text    string
	Colored bool
}

// StyledLine contains the styled runs for one rendered output line.
type StyledLine struct {
	Runs []StyledRun
}

// PageData is the complete presentation model passed to the page template.
type PageData struct {
	Form          FormState
	Output        []StyledLine
	HasResult     bool
	Error         string
	RenderedWidth int
	Color         SelectedColor
}
