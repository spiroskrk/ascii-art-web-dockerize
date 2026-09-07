package web

import (
	"bytes"
	"fmt"

	"ascii-art/internal/output"
	"ascii-art/internal/renderer"
)

// Generate runs the one shared generation workflow: banner lookup,
// normalization, structured rendering, and alignment.
//
// Both the browser page and the download endpoint call this function, so a
// single definition of correct output serves both. If each handler built its
// own pipeline, the displayed result and the downloaded file could drift apart.
//
// This function is deliberately thin. It performs no ASCII algorithm of its
// own: it does not assemble glyphs, match substrings, convert colors, or
// distribute justify spacing. It coordinates the packages that already own
// those responsibilities.
func (app *App) Generate(in GenerationInput) (GeneratedASCII, error) {
	// Step 1 — resolve the banner from the preloaded registry.
	//
	// The registry was built and validated at startup, so this is a plain map
	// lookup rather than a file read. No request value ever becomes part of a
	// filesystem path, which removes path-traversal risk entirely.
	//
	// The identifier was already validated during form validation, but we
	// check again here: Generate must not assume that every future caller has
	// done that work. The check is cheap and keeps this function safe on its
	// own terms.
	glyphs, ok := app.banners[in.Banner]
	if !ok {
		return GeneratedASCII{}, fmt.Errorf("unknown banner %q", in.Banner)
	}

	// Step 2 — split the submitted text into logical lines.
	//
	// NormalizeInput treats LF, CRLF, lone CR, and the literal \n escape as
	// equivalent separators, so browser input and legacy CLI-style input
	// produce the same logical lines.
	lines := renderer.NormalizeInput(in.Text)

	// Step 3 — decide what colour argument the renderer receives.
	//
	// SelectedColor separates "is colouring on?" from "which colour?". When
	// colouring is disabled we pass an empty string, which is how the renderer
	// expresses "no colour" — the submitted colour value is simply ignored
	// rather than silently applied.
	//
	// The validated Hex is already in the exact #RRGGBB form the renderer
	// parses, so no conversion is needed here.
	colorArg := ""
	if in.Color.Enabled {
		colorArg = in.Color.Hex
	}

	// Step 4 — build the structured ASCII.
	//
	// BuildASCII returns word and space segments rather than flat rows,
	// because justify must be able to expand only the real gaps between words,
	// not the spaces that exist inside a glyph.
	rendered, err := renderer.BuildASCII(lines, glyphs, in.Substring, colorArg)
	if err != nil {
		return GeneratedASCII{}, err
	}

	// Step 5 — apply alignment into an in-memory buffer.
	//
	// WriteAlignedASCII writes to any io.Writer. The CLI passed standard
	// output; here we pass a buffer, so the exact same alignment code produces
	// the text we hand back. This is the reuse rule in practice: the web layer
	// changes the destination, not the algorithm.
	var buf bytes.Buffer
	if err := output.WriteAlignedASCII(&buf, rendered, in.Alignment, in.Width); err != nil {
		return GeneratedASCII{}, err
	}

	// Step 6 — return aligned ANSI text plus the width it was aligned at.
	//
	// The width is recorded so the page can report it and the download can
	// reproduce the same snapshot. Errors are returned unwrapped: deciding
	// which HTTP status a failure deserves belongs to the handler layer, not
	// here.
	return GeneratedASCII{
		ANSIText: buf.String(),
		Width:    in.Width,
	}, nil
}
