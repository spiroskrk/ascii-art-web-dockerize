package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// getHome renders the default page through the real router and template.
func getHome(t *testing.T) *httptest.ResponseRecorder {
	t.Helper()
	app := newTestApp(t, parseProductionTemplate(t))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	app.Routes().ServeHTTP(response, request)
	return response
}

func TestHomeReturnsOKWithCompleteForm(t *testing.T) {
	response := getHome(t)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := response.Body.String()
	for _, fragment := range []string{
		`name="text"`,
		`name="banner"`,
		`value="standard"`,
		`value="shadow"`,
		`value="thinkertoy"`,
		`type="submit"`,
	} {
		if !strings.Contains(body, fragment) {
			t.Errorf("default page does not contain %q", fragment)
		}
	}
}

func TestHomeSelectsDefaultBanner(t *testing.T) {
	body := getHome(t).Body.String()

	standard := strings.Index(body, `value="standard"`)
	if standard < 0 {
		t.Fatal("standard banner option is missing")
	}
	// The checked attribute is rendered immediately after the value on the
	// selected radio, so a short window after it is enough to prove which
	// option the server marked as selected.
	window := body[standard:min(standard+80, len(body))]
	if !strings.Contains(window, "checked") {
		t.Error("standard banner is not selected by default")
	}

	shadow := strings.Index(body, `value="shadow"`)
	if shadow < 0 {
		t.Fatal("shadow banner option is missing")
	}
	bannerControls := body[standard:shadow]
	if strings.Count(bannerControls, "checked") != 1 {
		t.Errorf("banner controls contain %d checked attributes, want one",
			strings.Count(bannerControls, "checked"))
	}
}

func TestHomeContainsAlignmentAndClearControls(t *testing.T) {
	body := getHome(t).Body.String()

	for _, fragment := range []string{
		`name="align" value="left" required`,
		`name="align" value="center"`,
		`name="align" value="right"`,
		`name="align" value="justify"`,
		`class="button button-ghost" href="/">Clear</a>`,
	} {
		if !strings.Contains(body, fragment) {
			t.Errorf("default page does not contain %q", fragment)
		}
	}

	left := strings.Index(body, `name="align" value="left"`)
	center := strings.Index(body, `name="align" value="center"`)
	if left < 0 || center < 0 {
		t.Fatal("alignment controls are incomplete")
	}
	if got := strings.Count(body[left:center], "checked"); got != 1 {
		t.Errorf("left alignment section contains %d checked attributes, want one", got)
	}
}

func TestHomeContainsColorSubstringAndWidthControls(t *testing.T) {
	body := getHome(t).Body.String()

	for _, fragment := range []string{
		`type="checkbox" name="use_color"`,
		`id="color" type="color" name="color" value="#ff0000"`,
		`id="substring" class="sub-input" type="text" name="substring"`,
		`maxlength="4096"`,
		`type="hidden" name="width" value="80"`,
	} {
		if !strings.Contains(body, fragment) {
			t.Errorf("default page does not contain %q", fragment)
		}
	}

	useColor := strings.Index(body, `name="use_color"`)
	if useColor < 0 {
		t.Fatal("use_color checkbox is missing")
	}
	window := body[useColor:min(useColor+80, len(body))]
	if strings.Contains(window, "checked") {
		t.Error("use_color checkbox is checked by default")
	}
}

func TestHomeContainsThemeToggle(t *testing.T) {
	body := getHome(t).Body.String()

	for _, fragment := range []string{
		`class="theme-control"`,
		`id="theme-toggle-label"`,
		`Dark theme`,
		`type="button"`,
		`class="theme-toggle"`,
		`data-theme-toggle`,
		`aria-labelledby="theme-toggle-label"`,
		`aria-pressed="false"`,
		`class="theme-toggle-status"`,
		`data-theme-toggle-status`,
		`aria-hidden="true">OFF</span>`,
		`class="theme-toggle-thumb"`,
	} {
		if !strings.Contains(body, fragment) {
			t.Errorf("default page does not contain %q", fragment)
		}
	}
}

func TestHomeHasNoErrorOrDownloadForm(t *testing.T) {
	body := getHome(t).Body.String()

	if strings.Contains(body, `role="alert"`) {
		t.Error("default page contains an error message")
	}
	if strings.Contains(body, `action="/ascii-art/download"`) {
		t.Error("default page contains a download form before generation")
	}
}

func TestPagePreservesErrorStateSafely(t *testing.T) {
	data := PageData{
		Form: FormState{
			Text:      `<script>alert("text")</script>`,
			Banner:    BannerShadow,
			Color:     "#ff0000",
			Substring: `<img src=x onerror=alert(1)>`,
			Alignment: "right",
			Width:     "80",
		},
		Error: `Invalid <strong>input</strong>`,
	}
	body := renderPageForTest(t, data)

	for _, unsafe := range []string{"<script>", "<img", "<strong>"} {
		if strings.Contains(body, unsafe) {
			t.Errorf("rendered page contains unescaped markup %q", unsafe)
		}
	}
	for _, escaped := range []string{"&lt;script&gt;", "&lt;strong&gt;"} {
		if !strings.Contains(body, escaped) {
			t.Errorf("rendered page does not preserve escaped value %q", escaped)
		}
	}
	if !strings.Contains(body, `role="alert"`) {
		t.Error("rendered page does not expose the error as an alert")
	}
}

func TestPageRendersColoredRunsWithSafeRGBSpan(t *testing.T) {
	data := PageData{
		Color: SelectedColor{
			Enabled: true,
			Hex:     "#aa10ff",
			Red:     170,
			Green:   16,
			Blue:    255,
		},
		Output: []StyledLine{
			{Runs: []StyledRun{
				{Text: "<A>", Colored: true},
				{Text: "&B", Colored: false},
			}},
		},
		HasResult:     true,
		RenderedWidth: 80,
	}
	body := renderPageForTest(t, data)

	if !strings.Contains(body, `style="color: rgb(170, 16, 255)"`) {
		t.Errorf("rendered page does not contain the validated RGB style: %s", body)
	}
	if !strings.Contains(body, `&lt;A&gt;`) {
		t.Error("colored run text was not HTML-escaped")
	}
	if !strings.Contains(body, `&amp;B`) {
		t.Error("uncolored run text was not HTML-escaped")
	}
	if strings.Contains(body, `<A>`) {
		t.Error("colored run text rendered as raw HTML")
	}
}

func TestSuccessfulPageContainsSeparateDownloadSnapshot(t *testing.T) {
	data := PageData{
		Form: FormState{
			Text:      "Hello\nWorld",
			Banner:    BannerThinkertoy,
			UseColor:  true,
			Color:     "#12abEF",
			Substring: "lo",
			Alignment: "center",
			Width:     "80",
		},
		HasResult:     true,
		RenderedWidth: 120,
	}
	body := renderPageForTest(t, data)

	generationStart := strings.Index(body, `class="generation-form`)
	downloadStart := strings.Index(body, `action="/ascii-art/download"`)
	if generationStart < 0 || downloadStart < 0 {
		t.Fatal("page does not contain both expected forms")
	}
	relativeGenerationEnd := strings.Index(body[generationStart:], "</form>")
	if relativeGenerationEnd < 0 {
		t.Fatal("generation form has no closing tag")
	}
	generationEnd := generationStart + relativeGenerationEnd
	if downloadStart < generationEnd {
		t.Error("download form is nested inside the generation form")
	}

	for _, fragment := range []string{
		`<textarea name="text" hidden>Hello`,
		`World</textarea>`,
		`name="banner" value="thinkertoy"`,
		`name="use_color" value="on"`,
		`name="color" value="#12abEF"`,
		`name="substring" value="lo"`,
		`name="align" value="center"`,
		`name="width" value="120"`,
	} {
		if !strings.Contains(body[downloadStart:], fragment) {
			t.Errorf("download snapshot does not contain %q", fragment)
		}
	}
}

func TestHomeStartsWithEmptyTextAndNoResult(t *testing.T) {
	body := getHome(t).Body.String()

	if !strings.Contains(body, `placeholder="Type something"></textarea>`) {
		t.Error("text area is not empty on the default page")
	}
	if strings.Contains(body, "columns") {
		t.Error("default page reports a rendered width without a result")
	}
}

func TestDefaultFormStateMatchesContract(t *testing.T) {
	form := defaultFormState()

	tests := []struct {
		field string
		got   string
		want  string
	}{
		{"Text", form.Text, ""},
		{"Banner", form.Banner, BannerStandard},
		{"Color", form.Color, "#ff0000"},
		{"Substring", form.Substring, ""},
		{"Alignment", form.Alignment, "left"},
		{"Width", form.Width, "80"},
	}
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("default %s = %q, want %q", test.field, test.got, test.want)
		}
	}
	if form.UseColor {
		t.Error("colouring is enabled by default, want it off")
	}
}

func TestClearReturnsTheDefaultPage(t *testing.T) {
	// Clear is defined as a fresh GET of "/", so it must produce exactly the
	// same response as a first visit rather than an HTML form reset.
	first := getHome(t).Body.String()
	second := getHome(t).Body.String()

	if first != second {
		t.Error("clearing the page does not reproduce the default page")
	}
}

func renderPageForTest(t *testing.T, data PageData) string {
	t.Helper()
	app := newTestApp(t, parseProductionTemplate(t))
	response := httptest.NewRecorder()
	app.renderPage(response, http.StatusOK, data)
	if response.Code != http.StatusOK {
		t.Fatalf("renderPage() status = %d, want 200", response.Code)
	}
	return response.Body.String()
}
