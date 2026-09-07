package web

import "net/http"

func (app *App) handleHome(w http.ResponseWriter, _ *http.Request) {
	app.renderPage(w, http.StatusOK, PageData{Form: defaultFormState()})
}

// handleGenerate parses, validates, generates, and presents ASCII art for the
// submitted form, then re-renders the main page with the result.
//
// Errors are routed through the shared render* helpers so this handler makes
// no HTTP-status decisions of its own: parseGenerationForm and Generate errors
// already carry their status via renderFormFailure and renderGenerationFailure,
// and a Present failure is always an internal error, never the client's fault.
func (app *App) handleGenerate(w http.ResponseWriter, r *http.Request) {
	state, input, err := parseGenerationForm(w, r)
	if err != nil {
		app.renderFormFailure(w, state, err)
		return
	}

	generated, err := app.Generate(input)
	if err != nil {
		app.renderGenerationFailure(w, state, "generate", err)
		return
	}

	// The download endpoint calls the same three functions above and keeps
	// only the plain text; the browser page keeps only the styled lines.
	lines, _, err := Present(generated)
	if err != nil {
		app.renderInternalFailure(w, "present", err)
		return
	}

	app.renderPage(w, http.StatusOK, PageData{
		Form:          state,
		Output:        lines,
		HasResult:     true,
		RenderedWidth: generated.Width,
		Color:         input.Color,
	})
}

func defaultFormState() FormState {
	return FormState{
		Banner:    BannerStandard,
		Color:     "#ff0000",
		Alignment: "left",
		Width:     "80",
	}
}
