package web

import (
	"bytes"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestNewAppValidatesDependencies(t *testing.T) {
	pageTemplate := parseProductionTemplate(t)
	registry := testBannerRegistry()
	logger := testLogger()

	tests := []struct {
		name      string
		template  *template.Template
		registry  BannerRegistry
		logger    *slog.Logger
		wantError string
	}{
		{
			name:      "missing template",
			registry:  registry,
			logger:    logger,
			wantError: "page template is required",
		},
		{
			name:      "missing definition",
			template:  template.Must(template.New("incomplete").Parse(`{{define "page"}}page{{end}}`)),
			registry:  registry,
			logger:    logger,
			wantError: "missing template definition",
		},
		{
			name:      "missing logger",
			template:  pageTemplate,
			registry:  registry,
			wantError: "logger is required",
		},
		{
			name:      "missing standard banner",
			template:  pageTemplate,
			registry:  registryWithout(BannerStandard),
			logger:    logger,
			wantError: `banner "standard" is required`,
		},
		{
			name:      "missing shadow banner",
			template:  pageTemplate,
			registry:  registryWithout(BannerShadow),
			logger:    logger,
			wantError: `banner "shadow" is required`,
		},
		{
			name:      "missing thinkertoy banner",
			template:  pageTemplate,
			registry:  registryWithout(BannerThinkertoy),
			logger:    logger,
			wantError: `banner "thinkertoy" is required`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewApp(test.template, test.registry, test.logger)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("NewApp() error = %v, want error containing %q", err, test.wantError)
			}
		})
	}
}

func TestNewAppCopiesBannerRegistry(t *testing.T) {
	registry := testBannerRegistry()
	app, err := NewApp(parseProductionTemplate(t), registry, testLogger())
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	registry[BannerStandard] = map[rune][]string{'X': {"changed"}}
	delete(registry, BannerShadow)

	if _, ok := app.banners[BannerStandard][' ']; !ok {
		t.Fatal("replacing a caller registry entry changed the app registry")
	}
	if _, ok := app.banners[BannerShadow]; !ok {
		t.Fatal("deleting a caller registry entry changed the app registry")
	}
}

func TestRoutesPhaseZeroContract(t *testing.T) {
	app := newTestApp(t, parseProductionTemplate(t))
	handler := app.Routes()

	tests := []struct {
		name        string
		method      string
		path        string
		wantStatus  int
		wantAllowed string
	}{
		{name: "home", method: http.MethodGet, path: "/", wantStatus: http.StatusOK},
		{name: "unknown", method: http.MethodGet, path: "/unknown", wantStatus: http.StatusNotFound},
		{name: "wrong home method", method: http.MethodPost, path: "/", wantStatus: http.StatusMethodNotAllowed, wantAllowed: http.MethodGet},
		{name: "wrong generate method", method: http.MethodGet, path: "/ascii-art", wantStatus: http.StatusMethodNotAllowed, wantAllowed: http.MethodPost},
		{name: "wrong download method", method: http.MethodGet, path: "/ascii-art/download", wantStatus: http.StatusMethodNotAllowed, wantAllowed: http.MethodPost},
		{name: "generate without a body is rejected", method: http.MethodPost, path: "/ascii-art", wantStatus: http.StatusUnsupportedMediaType},
		{name: "download validates media type", method: http.MethodPost, path: "/ascii-art/download", wantStatus: http.StatusUnsupportedMediaType},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if test.wantAllowed != "" && !headerContainsToken(response.Header().Get("Allow"), test.wantAllowed) {
				t.Errorf("Allow = %q, want it to contain %q", response.Header().Get("Allow"), test.wantAllowed)
			}
		})
	}
}

func TestNotFoundReturnsMinimalHTMLPage(t *testing.T) {
	app := newTestApp(t, parseProductionTemplate(t))
	handler := app.Routes()

	for _, path := range []string{"/unknown", "/templates/404.html", "/static/missing.js"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", response.Code)
			}
			if got := response.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
				t.Errorf("Content-Type = %q, want text/html; charset=utf-8", got)
			}
			wantLength := strconv.Itoa(response.Body.Len())
			if got := response.Header().Get("Content-Length"); got != wantLength {
				t.Errorf("Content-Length = %q, want %q", got, wantLength)
			}

			body := response.Body.String()
			for _, fragment := range []string{
				"<!doctype html>",
				"Page not found",
				`href="/"`,
				`href="/static/style.css"`,
			} {
				if !strings.Contains(body, fragment) {
					t.Errorf("response body does not contain %q", fragment)
				}
			}
			if strings.Contains(body, path) {
				t.Errorf("response body exposes requested path %q", path)
			}
		})
	}
}

func TestStaticAssetRoutesServeOnlyKnownFiles(t *testing.T) {
	app := newTestApp(t, parseProductionTemplate(t))
	handler := app.Routes()
	chdirToRepoRoot(t)

	tests := []struct {
		name        string
		method      string
		path        string
		wantStatus  int
		wantAllowed string
		wantType    string
		wantBody    string
	}{
		{
			name:       "style css",
			method:     http.MethodGet,
			path:       "/static/style.css",
			wantStatus: http.StatusOK,
			wantType:   "text/css; charset=utf-8",
			wantBody:   `html[data-theme="dark"]`,
		},
		{
			name:       "app js",
			method:     http.MethodGet,
			path:       "/static/app.js",
			wantStatus: http.StatusOK,
			wantType:   "text/javascript; charset=utf-8",
			wantBody:   "data-theme-toggle",
		},
		{
			name:       "static directory not exposed",
			method:     http.MethodGet,
			path:       "/static/",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unexpected static file not exposed",
			method:     http.MethodGet,
			path:       "/static/missing.js",
			wantStatus: http.StatusNotFound,
		},
		{
			name:        "wrong static method",
			method:      http.MethodPost,
			path:        "/static/app.js",
			wantStatus:  http.StatusMethodNotAllowed,
			wantAllowed: http.MethodGet,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if test.wantAllowed != "" && !headerContainsToken(response.Header().Get("Allow"), test.wantAllowed) {
				t.Errorf("Allow = %q, want it to contain %q", response.Header().Get("Allow"), test.wantAllowed)
			}
			if test.wantType != "" && response.Header().Get("Content-Type") != test.wantType {
				t.Errorf("Content-Type = %q, want %q", response.Header().Get("Content-Type"), test.wantType)
			}
			if test.wantBody != "" && !strings.Contains(response.Body.String(), test.wantBody) {
				t.Errorf("response body does not contain %q", test.wantBody)
			}
		})
	}
}

func TestRequestLoggingMiddlewareRecordsSafeMetadata(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	app, err := NewApp(parseProductionTemplate(t), testBannerRegistry(), logger)
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	values := url.Values{
		"text":      {"secret text should not be logged"},
		"banner":    {BannerShadow},
		"substring": {"hidden substring"},
		"align":     {"left"},
		"width":     {"80"},
		"color":     {"#ff0000"},
	}
	request := httptest.NewRequest(http.MethodPost, "/ascii-art", strings.NewReader(values.Encode()))
	request.Header.Set("Content-Type", formMediaType)
	request.Header.Set("Cookie", "session=secret-cookie")
	response := httptest.NewRecorder()

	app.Routes().ServeHTTP(response, request)

	output := logs.String()
	for _, fragment := range []string{
		"msg=\"http request\"",
		"method=POST",
		"path=/ascii-art",
		"status=400",
		"duration=",
	} {
		if !strings.Contains(output, fragment) {
			t.Errorf("log output does not contain safe field %q: %s", fragment, output)
		}
	}
	for _, unsafe := range []string{
		"secret text should not be logged",
		"hidden substring",
		"secret-cookie",
		"session=",
		values.Encode(),
	} {
		if strings.Contains(output, unsafe) {
			t.Errorf("log output contains unsafe request data %q: %s", unsafe, output)
		}
	}
}

func chdirToRepoRoot(t *testing.T) {
	t.Helper()
	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("change working directory to repo root: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}

func TestHomeReturnsCompleteHTMLShell(t *testing.T) {
	app := newTestApp(t, parseProductionTemplate(t))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	app.Routes().ServeHTTP(response, request)

	if got := response.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html; charset=utf-8", got)
	}
	for _, fragment := range []string{
		"<!doctype html>",
		`action="/ascii-art"`,
		`href="/static/style.css"`,
		`src="/static/app.js"`,
	} {
		if !strings.Contains(response.Body.String(), fragment) {
			t.Errorf("response body does not contain %q", fragment)
		}
	}
}

func TestTemplateFailureReturnsCleanInternalServerError(t *testing.T) {
	failingTemplate := template.Must(template.New("index.html").Parse(failingPageTemplate))
	app := newTestApp(t, failingTemplate)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	app.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if strings.Contains(response.Body.String(), "partial page") {
		t.Fatalf("response leaked partial template output: %q", response.Body.String())
	}
	if response.Body.String() != http.StatusText(http.StatusInternalServerError)+"\n" {
		t.Errorf("body = %q, want a generic internal error", response.Body.String())
	}
}

func parseProductionTemplate(t *testing.T) *template.Template {
	t.Helper()
	pageTemplate, err := template.ParseFiles("../../templates/index.html", "../../templates/404.html")
	if err != nil {
		t.Fatalf("parse production template: %v", err)
	}
	return pageTemplate
}

func newTestApp(t *testing.T, pageTemplate *template.Template) *App {
	t.Helper()
	app, err := NewApp(pageTemplate, testBannerRegistry(), testLogger())
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	return app
}

func testBannerRegistry() BannerRegistry {
	return BannerRegistry{
		BannerStandard:   {' ': {"standard"}},
		BannerShadow:     {' ': {"shadow"}},
		BannerThinkertoy: {' ': {"thinkertoy"}},
	}
}

func registryWithout(name string) BannerRegistry {
	registry := testBannerRegistry()
	delete(registry, name)
	return registry
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func headerContainsToken(header string, token string) bool {
	for _, value := range strings.Split(header, ",") {
		if strings.TrimSpace(value) == token {
			return true
		}
	}
	return false
}

const failingPageTemplate = `
{{define "page"}}partial page {{.Missing}}{{end}}
{{define "input_controls"}}{{end}}
{{define "generate_action"}}{{end}}
{{define "output_box"}}{{end}}
{{define "align_controls"}}{{end}}
{{define "form_error"}}{{end}}
{{define "clear_action"}}{{end}}
{{define "download_action"}}{{end}}
{{define "color_controls"}}{{end}}
{{define "width_field"}}{{end}}
{{define "theme_control"}}{{end}}
{{define "not_found"}}{{end}}
`
