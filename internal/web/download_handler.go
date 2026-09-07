package web

import (
	"net/http"
	"strconv"
)

const (
	downloadContentType = "text/plain; charset=utf-8"
	downloadDisposition = `attachment; filename="ascii-art.txt"`
)

// handleDownload treats the hidden snapshot exactly like any other untrusted
// request: parse, validate, regenerate, and present it through shared code.
func (app *App) handleDownload(w http.ResponseWriter, r *http.Request) {
	state, input, err := parseGenerationForm(w, r)
	if err != nil {
		app.renderFormFailure(w, state, err)
		return
	}

	generated, err := app.Generate(input)
	if err != nil {
		app.renderGenerationFailure(w, state, "generate download", err)
		return
	}

	_, plain, err := Present(generated)
	if err != nil {
		app.renderInternalFailure(w, "present download", err)
		return
	}

	// Attachment headers are deliberately set only after every fallible step.
	// Failed downloads therefore return HTML or a generic 500, never a partial
	// attachment response.
	w.Header().Set("Content-Type", downloadContentType)
	w.Header().Set("Content-Disposition", downloadDisposition)
	w.Header().Set("Content-Length", strconv.Itoa(len(plain)))
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(plain)); err != nil {
		app.logger.Error("write download response", "error", err)
	}
}
