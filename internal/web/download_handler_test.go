package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestDownloadReturnsRegeneratedPlainTextAttachment(t *testing.T) {
	app := newDownloadTestApp(t)

	tests := []struct {
		banner string
		marker string
	}{
		{banner: BannerStandard, marker: "#"},
		{banner: BannerShadow, marker: "$"},
		{banner: BannerThinkertoy, marker: "%"},
	}

	for _, test := range tests {
		t.Run(test.banner, func(t *testing.T) {
			values := validFormValues()
			values.Set("text", "Hi")
			values.Set("banner", test.banner)
			values.Set("use_color", "on")
			values.Set("color", "#12abef")
			values.Set("substring", "H")
			values.Set("align", "right")
			values.Set("width", "20")

			response := performDownload(app, values, formMediaType)
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body = %q", response.Code, response.Body.String())
			}
			if got := response.Header().Get("Content-Type"); got != downloadContentType {
				t.Errorf("Content-Type = %q, want %q", got, downloadContentType)
			}
			if got := response.Header().Get("Content-Disposition"); got != downloadDisposition {
				t.Errorf("Content-Disposition = %q, want %q", got, downloadDisposition)
			}
			wantLength := strconv.Itoa(response.Body.Len())
			if got := response.Header().Get("Content-Length"); got != wantLength {
				t.Errorf("Content-Length = %q, want %q", got, wantLength)
			}
			if strings.Contains(response.Body.String(), "\033[") {
				t.Error("download contains an ANSI escape sequence")
			}

			firstRow := strings.SplitN(response.Body.String(), "\n", 2)[0]
			wantRow := strings.Repeat(" ", 16) + strings.Repeat(test.marker, 4)
			if firstRow != wantRow {
				t.Errorf("first aligned row = %q, want %q", firstRow, wantRow)
			}
		})
	}
}

func TestDownloadUsesBodySnapshotInsteadOfQueryValues(t *testing.T) {
	app := newDownloadTestApp(t)
	values := validFormValues()
	values.Set("text", "Hi")

	request := httptest.NewRequest(
		http.MethodPost,
		"/ascii-art/download?text=%E2%98%83&banner=unknown",
		strings.NewReader(values.Encode()),
	)
	request.Header.Set("Content-Type", formMediaType)
	response := httptest.NewRecorder()
	app.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want body snapshot to succeed", response.Code)
	}
}

func TestDownloadPreservesMultilineSnapshot(t *testing.T) {
	values := validFormValues()
	values.Set("text", "Hi\nYo")

	response := performDownload(newDownloadTestApp(t), values, formMediaType)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if got := strings.Count(response.Body.String(), "\n"); got != 16 {
		t.Errorf("download row count = %d, want 16", got)
	}
}

func TestDownloadMissingWidthUsesFallback(t *testing.T) {
	values := validFormValues()
	values.Set("text", "Hi")
	values.Set("align", "right")
	values.Del("width")

	response := performDownload(newDownloadTestApp(t), values, formMediaType)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	firstRow := strings.SplitN(response.Body.String(), "\n", 2)[0]
	if want := strings.Repeat(" ", 76) + "####"; firstRow != want {
		t.Errorf("first fallback-width row = %q, want %q", firstRow, want)
	}
}

func TestDownloadValidationFailuresReturnHTMLWithoutAttachment(t *testing.T) {
	tooLargeBody := "text=" + strings.Repeat("a", maxFormBodySize)

	tests := []struct {
		name        string
		values      url.Values
		body        string
		contentType string
		wantStatus  int
		wantText    string
	}{
		{
			name:        "unsupported media type",
			body:        "text=Hello",
			contentType: "text/plain",
			wantStatus:  http.StatusUnsupportedMediaType,
			wantText:    "URL-encoded",
		},
		{
			name:        "oversized body",
			body:        tooLargeBody,
			contentType: formMediaType,
			wantStatus:  http.StatusRequestEntityTooLarge,
			wantText:    "too large",
		},
		{
			name:        "malformed encoding",
			body:        "text=%zz",
			contentType: formMediaType,
			wantStatus:  http.StatusBadRequest,
			wantText:    "malformed",
		},
		{
			name:        "empty text",
			values:      formValuesWith("text", ""),
			contentType: formMediaType,
			wantStatus:  http.StatusBadRequest,
			wantText:    "Text is required",
		},
		{
			name:        "unknown banner",
			values:      formValuesWith("banner", "../../banner.txt"),
			contentType: formMediaType,
			wantStatus:  http.StatusBadRequest,
			wantText:    "valid banner",
		},
		{
			name:        "invalid enabled color",
			values:      enabledColorValues("red; background:url(example)"),
			contentType: formMediaType,
			wantStatus:  http.StatusBadRequest,
			wantText:    "#RRGGBB",
		},
		{
			name:        "invalid alignment",
			values:      formValuesWith("align", "diagonal"),
			contentType: formMediaType,
			wantStatus:  http.StatusBadRequest,
			wantText:    "valid alignment",
		},
		{
			name:        "invalid width",
			values:      formValuesWith("width", "301"),
			contentType: formMediaType,
			wantStatus:  http.StatusBadRequest,
			wantText:    "20 through 300",
		},
		{
			name:        "unexpected filename",
			values:      formValuesWith("filename", "../../owned.txt"),
			contentType: formMediaType,
			wantStatus:  http.StatusBadRequest,
			wantText:    "unexpected field",
		},
		{
			name:        "duplicate field",
			values:      duplicateFormValues("text"),
			contentType: formMediaType,
			wantStatus:  http.StatusBadRequest,
			wantText:    "only once",
		},
		{
			name:        "unsupported character",
			values:      formValuesWith("text", "snowman: ☃"),
			contentType: formMediaType,
			wantStatus:  http.StatusBadRequest,
			wantText:    "unsupported character",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := test.body
			if test.values != nil {
				body = test.values.Encode()
			}
			request := httptest.NewRequest(http.MethodPost, "/ascii-art/download", strings.NewReader(body))
			request.Header.Set("Content-Type", test.contentType)
			response := httptest.NewRecorder()
			newDownloadTestApp(t).Routes().ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %q", response.Code, test.wantStatus, response.Body.String())
			}
			if got := response.Header().Get("Content-Disposition"); got != "" {
				t.Errorf("failed download Content-Disposition = %q, want empty", got)
			}
			if got := response.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
				t.Errorf("failed download Content-Type = %q, want HTML", got)
			}
			if !strings.Contains(response.Body.String(), test.wantText) {
				t.Errorf("error page does not contain %q", test.wantText)
			}
			if strings.Contains(response.Body.String(), `name="text"`) &&
				strings.Contains(response.Body.String(), `action="/ascii-art/download"`) {
				t.Error("failed download exposed a new download snapshot")
			}
		})
	}
}

func TestDownloadErrorPreservesEscapedFormState(t *testing.T) {
	values := validFormValues()
	values.Set("text", `<script>alert("kept")</script>`)
	values.Set("align", "invalid")

	response := performDownload(newDownloadTestApp(t), values, formMediaType)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
	body := response.Body.String()
	if !strings.Contains(body, `&lt;script&gt;alert`) {
		t.Error("error page did not preserve escaped submitted text")
	}
	if strings.Contains(body, "<script>") {
		t.Error("error page rendered submitted text as executable markup")
	}
	if strings.Contains(body, `action="/ascii-art/download"`) {
		t.Error("error page contains a download snapshot")
	}
}

func TestDownloadPresenterFailureReturnsSafeInternalError(t *testing.T) {
	app := newDownloadTestApp(t)
	badRows := make([]string, 8)
	for i := range badRows {
		badRows[i] = "\033[999mX"
	}
	app.banners[BannerStandard]['H'] = badRows

	values := validFormValues()
	values.Set("text", "H")
	response := performDownload(app, values, formMediaType)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
	if got := response.Header().Get("Content-Disposition"); got != "" {
		t.Errorf("failed download Content-Disposition = %q, want empty", got)
	}
	if response.Body.String() != http.StatusText(http.StatusInternalServerError)+"\n" {
		t.Errorf("body = %q, want generic internal error", response.Body.String())
	}
}

func newDownloadTestApp(t *testing.T) *App {
	t.Helper()
	registry := BannerRegistry{
		BannerStandard:   generateBanner("#"),
		BannerShadow:     generateBanner("$"),
		BannerThinkertoy: generateBanner("%"),
	}
	app, err := NewApp(parseProductionTemplate(t), registry, testLogger())
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	return app
}

func performDownload(app *App, values url.Values, contentType string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/ascii-art/download", strings.NewReader(values.Encode()))
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()
	app.Routes().ServeHTTP(response, request)
	return response
}

func formValuesWith(key string, value string) url.Values {
	values := validFormValues()
	values.Set(key, value)
	return values
}

func enabledColorValues(color string) url.Values {
	values := validFormValues()
	values.Set("use_color", "on")
	values.Set("color", color)
	return values
}

func duplicateFormValues(key string) url.Values {
	values := validFormValues()
	values[key] = append(values[key], "duplicate")
	return values
}
