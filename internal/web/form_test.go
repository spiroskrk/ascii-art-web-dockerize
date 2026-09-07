package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestParsePostFormAcceptsURLEncodedBody(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
	}{
		{name: "media type", contentType: formMediaType},
		{name: "media type with parameter", contentType: formMediaType + "; charset=utf-8"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := newFormRequest("/?text=query", "text=body&banner=standard", test.contentType)
			values, err := parsePostForm(httptest.NewRecorder(), request)
			if err != nil {
				t.Fatalf("parsePostForm() error = %v", err)
			}
			if got := values.Get("text"); got != "body" {
				t.Errorf("PostForm text = %q, want body", got)
			}
			if got := values.Get("banner"); got != BannerStandard {
				t.Errorf("PostForm banner = %q, want %q", got, BannerStandard)
			}
			if got := values["text"]; len(got) != 1 {
				t.Errorf("PostForm text values = %q, want body value only", got)
			}
		})
	}
}

func TestParsePostFormRejectsUnsupportedMediaType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
	}{
		{name: "missing", contentType: ""},
		{name: "plain text", contentType: "text/plain"},
		{name: "multipart", contentType: "multipart/form-data; boundary=example"},
		{name: "malformed", contentType: `application/x-www-form-urlencoded; charset="`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := newFormRequest("/", "text=hello", test.contentType)
			_, err := parsePostForm(httptest.NewRecorder(), request)
			assertFormRequestStatus(t, err, http.StatusUnsupportedMediaType)
		})
	}
}

func TestParsePostFormEnforcesBodyLimitBeforeParsing(t *testing.T) {
	exactBody := "text=" + strings.Repeat("a", maxFormBodySize-len("text="))
	overBody := exactBody + "a"

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "exactly at limit", body: exactBody},
		{name: "one byte over limit", body: overBody, wantStatus: http.StatusRequestEntityTooLarge},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := newFormRequest("/", test.body, formMediaType)
			values, err := parsePostForm(httptest.NewRecorder(), request)
			if test.wantStatus != 0 {
				assertFormRequestStatus(t, err, test.wantStatus)
				return
			}
			if err != nil {
				t.Fatalf("parsePostForm() error = %v", err)
			}
			if got := len(values.Get("text")); got != maxFormBodySize-len("text=") {
				t.Errorf("parsed text length = %d, want %d", got, maxFormBodySize-len("text="))
			}
		})
	}
}

func TestParsePostFormRejectsMalformedEncoding(t *testing.T) {
	request := newFormRequest("/", "text=%zz", formMediaType)
	_, err := parsePostForm(httptest.NewRecorder(), request)
	assertFormRequestStatus(t, err, http.StatusBadRequest)
}

func TestDecodeFormCapturesRawValues(t *testing.T) {
	values := validFormValues()
	values.Set("text", "  Hello  ")
	values.Set("use_color", "on")
	values.Set("color", "#Aa10fF")
	values.Set("substring", " l ")

	got, err := decodeForm(values)
	if err != nil {
		t.Fatalf("decodeForm() error = %v", err)
	}

	want := FormState{
		Text:      "  Hello  ",
		Banner:    BannerStandard,
		UseColor:  true,
		Color:     "#Aa10fF",
		Substring: " l ",
		Alignment: "left",
		Width:     "80",
	}
	if got != want {
		t.Errorf("decodeForm() = %#v, want %#v", got, want)
	}
}

func TestDecodeFormUsesWidthFallbackOnlyWhenMissing(t *testing.T) {
	missing := validFormValues()
	missing.Del("width")
	state, err := decodeForm(missing)
	if err != nil {
		t.Fatalf("decodeForm() missing width error = %v", err)
	}
	if state.Width != "80" {
		t.Errorf("missing width = %q, want 80", state.Width)
	}

	explicitEmpty := validFormValues()
	explicitEmpty.Set("width", "")
	state, err = decodeForm(explicitEmpty)
	if err != nil {
		t.Fatalf("decodeForm() empty width error = %v", err)
	}
	if state.Width != "" {
		t.Errorf("explicit empty width = %q, want empty", state.Width)
	}
	if _, err := validateForm(state); err == nil {
		t.Fatal("validateForm() accepted an explicitly empty width")
	}
}

func TestDecodeFormRejectsUnexpectedAndDuplicateFields(t *testing.T) {
	t.Run("unexpected field", func(t *testing.T) {
		values := validFormValues()
		values.Set("filename", "../../output.txt")
		_, err := decodeForm(values)
		assertFormRequestStatus(t, err, http.StatusBadRequest)
	})

	for _, key := range []string{"text", "banner", "use_color", "color", "substring", "align", "width"} {
		t.Run("duplicate "+key, func(t *testing.T) {
			values := validFormValues()
			values[key] = []string{"first", "second"}
			_, err := decodeForm(values)
			assertFormRequestStatus(t, err, http.StatusBadRequest)
		})
	}
}

func TestDecodeFormValidatesCheckboxRepresentation(t *testing.T) {
	for _, value := range []string{"true", "1", "yes", ""} {
		t.Run(value, func(t *testing.T) {
			values := validFormValues()
			values.Set("use_color", value)
			_, err := decodeForm(values)
			assertFormRequestStatus(t, err, http.StatusBadRequest)
		})
	}

	values := validFormValues()
	state, err := decodeForm(values)
	if err != nil {
		t.Fatalf("decodeForm() absent checkbox error = %v", err)
	}
	if state.UseColor {
		t.Error("absent use_color enabled coloring")
	}
}

func TestValidateFormAcceptsValidValues(t *testing.T) {
	tests := []struct {
		name  string
		state FormState
	}{
		{name: "whitespace-only text", state: validFormState("   ")},
		{name: "minimum width", state: formStateWithWidth("20")},
		{name: "maximum width", state: formStateWithWidth("300")},
		{name: "shadow banner", state: formStateWithBanner(BannerShadow)},
		{name: "thinkertoy banner", state: formStateWithBanner(BannerThinkertoy)},
		{name: "center alignment", state: formStateWithAlignment("center")},
		{name: "right alignment", state: formStateWithAlignment("right")},
		{name: "justify alignment", state: formStateWithAlignment("justify")},
		{name: "4096 characters", state: validFormState(strings.Repeat("é", maxTextLength))},
		{name: "substring spaces preserved", state: formStateWithSubstring("  ell  ")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := validateForm(test.state)
			if err != nil {
				t.Fatalf("validateForm() error = %v", err)
			}
			if got.Text != test.state.Text || got.Substring != test.state.Substring {
				t.Errorf("validateForm() changed text or substring: %#v", got)
			}
		})
	}
}

func TestValidateFormColorBehavior(t *testing.T) {
	t.Run("disabled color is ignored", func(t *testing.T) {
		state := validFormState("Hello")
		state.Color = "not-a-color"

		got, err := validateForm(state)
		if err != nil {
			t.Fatalf("validateForm() error = %v", err)
		}
		if got.Color.Enabled || got.Color.Hex != "" {
			t.Errorf("disabled color = %#v, want disabled with no trusted value", got.Color)
		}
	})

	t.Run("enabled color retains validated hex", func(t *testing.T) {
		state := validFormState("Hello")
		state.UseColor = true
		state.Color = "#Aa10fF"

		got, err := validateForm(state)
		if err != nil {
			t.Fatalf("validateForm() error = %v", err)
		}
		if !got.Color.Enabled || got.Color.Hex != state.Color {
			t.Errorf("enabled color = %#v, want enabled hex %q", got.Color, state.Color)
		}
		if got.Color.Red != 170 || got.Color.Green != 16 || got.Color.Blue != 255 {
			t.Errorf("enabled color RGB = %d,%d,%d, want 170,16,255",
				got.Color.Red, got.Color.Green, got.Color.Blue)
		}
	})
}

func TestValidateFormRejectsInvalidValues(t *testing.T) {
	tooLongText := validFormState(strings.Repeat("a", maxTextLength+1))
	tooLongSubstring := validFormState("Hello")
	tooLongSubstring.Substring = strings.Repeat("a", maxTextLength+1)

	tests := []struct {
		name  string
		state FormState
	}{
		{name: "empty text", state: validFormState("")},
		{name: "text too long", state: tooLongText},
		{name: "missing banner", state: formStateWithBanner("")},
		{name: "unknown banner", state: formStateWithBanner("gothic")},
		{name: "substring too long", state: tooLongSubstring},
		{name: "missing alignment", state: formStateWithAlignment("")},
		{name: "unknown alignment", state: formStateWithAlignment("diagonal")},
		{name: "empty width", state: formStateWithWidth("")},
		{name: "non-numeric width", state: formStateWithWidth("wide")},
		{name: "negative width", state: formStateWithWidth("-20")},
		{name: "below minimum width", state: formStateWithWidth("19")},
		{name: "above maximum width", state: formStateWithWidth("301")},
	}

	invalidColors := []string{"", "ff0000", "#fff", "#gg0000", "#1234567", "#12345 "}
	for _, color := range invalidColors {
		state := validFormState("Hello")
		state.UseColor = true
		state.Color = color
		tests = append(tests, struct {
			name  string
			state FormState
		}{name: "invalid enabled color " + color, state: state})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := validateForm(test.state)
			assertFormRequestStatus(t, err, http.StatusBadRequest)
		})
	}
}

func newFormRequest(target string, body string, contentType string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	return request
}

func validFormValues() url.Values {
	return url.Values{
		"text":      {"Hello"},
		"banner":    {BannerStandard},
		"color":     {"#ff0000"},
		"substring": {""},
		"align":     {"left"},
		"width":     {"80"},
	}
}

func validFormState(text string) FormState {
	return FormState{
		Text:      text,
		Banner:    BannerStandard,
		Color:     "#ff0000",
		Alignment: "left",
		Width:     "80",
	}
}

func formStateWithWidth(width string) FormState {
	state := validFormState("Hello")
	state.Width = width
	return state
}

func formStateWithBanner(banner string) FormState {
	state := validFormState("Hello")
	state.Banner = banner
	return state
}

func formStateWithAlignment(alignment string) FormState {
	state := validFormState("Hello")
	state.Alignment = alignment
	return state
}

func formStateWithSubstring(substring string) FormState {
	state := validFormState("Hello")
	state.Substring = substring
	return state
}

func assertFormRequestStatus(t *testing.T, err error, want int) {
	t.Helper()
	if err == nil {
		t.Fatalf("parsePostForm() error = nil, want status %d", want)
	}

	var requestError *formRequestError
	if !errors.As(err, &requestError) {
		t.Fatalf("parsePostForm() error type = %T, want *formRequestError", err)
	}
	if requestError.status != want {
		t.Errorf("parsePostForm() status = %d, want %d", requestError.status, want)
	}
}
