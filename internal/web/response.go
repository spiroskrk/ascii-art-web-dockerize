package web

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
)

// renderPage buffers template output so execution failures can become a clean
// error response instead of a partially written successful page.
func (app *App) renderPage(w http.ResponseWriter, status int, data PageData) {
	app.renderHTML(w, status, "page", data)
}

func (app *App) renderHTML(w http.ResponseWriter, status int, name string, data any) {
	var body bytes.Buffer
	if err := app.pageTemplate.ExecuteTemplate(&body, name, data); err != nil {
		app.logger.Error("execute page template", slog.Any("error", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(body.Len()))
	w.WriteHeader(status)
	if _, err := w.Write(body.Bytes()); err != nil {
		app.logger.Error("write page response", slog.Any("error", err))
	}
}

// renderFormFailure returns a safe HTML error while preserving any raw form
// state that was decoded successfully. Detailed parsing errors remain available
// to callers for logging but are never exposed to the browser.
func (app *App) renderFormFailure(w http.ResponseWriter, state FormState, err error) {
	var formErr *formRequestError
	if !errors.As(err, &formErr) {
		app.renderInternalFailure(w, "unexpected form failure", err)
		return
	}

	app.renderPage(w, formErr.status, PageData{
		Form:  state,
		Error: formErr.message,
	})
}

func (app *App) renderInternalFailure(w http.ResponseWriter, operation string, err error) {
	app.logger.Error(operation, slog.Any("error", err))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// renderGenerationFailure maps the renderer's one expected client error while
// keeping all other generation failures internal. This compatibility check is
// centralized here until the renderer exposes a typed unsupported-character
// error; handlers must not duplicate string-based error classification.
func (app *App) renderGenerationFailure(w http.ResponseWriter, state FormState, operation string, err error) {
	if err.Error() == "unsupported character" {
		app.renderFormFailure(w, state, badFormRequest("Text contains an unsupported character."))
		return
	}
	app.renderInternalFailure(w, operation, err)
}
