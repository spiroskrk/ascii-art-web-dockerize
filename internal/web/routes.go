package web

import "net/http"

// Routes returns the application's complete HTTP handler.
func (app *App) Routes() http.Handler {
	mux := http.NewServeMux()
	registerRoute(mux, http.MethodGet, "/{$}", app.handleHome)
	registerRoute(mux, http.MethodPost, "/ascii-art", app.handleGenerate)
	registerRoute(mux, http.MethodPost, "/ascii-art/download", app.handleDownload)
	registerRoute(mux, http.MethodGet, "/static/style.css", app.handleStyleCSS)
	registerRoute(mux, http.MethodGet, "/static/app.js", app.handleAppJS)
	mux.HandleFunc("/", app.handleNotFound)
	return app.logRequests(mux)
}

// registerRoute keeps method errors distinct from the final not-found route.
// GET routes also accept HEAD, matching net/http's normal method semantics.
func registerRoute(mux *http.ServeMux, method string, pattern string, handler http.HandlerFunc) {
	mux.HandleFunc(method+" "+pattern, handler)
	mux.HandleFunc(pattern, func(w http.ResponseWriter, _ *http.Request) {
		allowed := method
		if method == http.MethodGet {
			allowed += ", " + http.MethodHead
		}
		w.Header().Set("Allow", allowed)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	})
}

func (app *App) handleNotFound(w http.ResponseWriter, _ *http.Request) {
	app.renderHTML(w, http.StatusNotFound, "not_found", nil)
}

func (app *App) handleStyleCSS(w http.ResponseWriter, r *http.Request) {
	// Serve a fixed asset path only; no browser value is joined into a path.
	http.ServeFile(w, r, "static/style.css")
}

func (app *App) handleAppJS(w http.ResponseWriter, r *http.Request) {
	// Serve a fixed asset path only; no browser value is joined into a path.
	http.ServeFile(w, r, "static/app.js")
}
