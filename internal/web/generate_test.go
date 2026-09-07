package web

import (
	"strings"
	"testing"
)

// generateBanner builds a deterministic 8-row glyph map. Every glyph is two
// identical marker columns on all eight rows, so a rendered row is exactly two
// visible columns per input character. Alignment padding then becomes trivial
// to assert without depending on real banner artwork.
func generateBanner(marker string) map[rune][]string {
	glyphs := make(map[rune][]string)
	for _, ch := range " ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" {
		rows := make([]string, 8)
		for r := range rows {
			if ch == ' ' {
				rows[r] = "  "
			} else {
				rows[r] = marker + marker
			}
		}
		glyphs[ch] = rows
	}
	return glyphs
}

// generateTestApp builds an App with only the dependency Generate needs.
// Constructing App directly (rather than through NewApp) keeps the test focused:
// generation does not touch templates or the logger.
func generateTestApp() *App {
	return &App{
		banners: BannerRegistry{
			BannerStandard:   generateBanner("#"),
			BannerShadow:     generateBanner("$"),
			BannerThinkertoy: generateBanner("%"),
		},
	}
}

func baseInput() GenerationInput {
	return GenerationInput{
		Text:      "Hi",
		Banner:    BannerStandard,
		Alignment: "left",
		Width:     80,
	}
}

func TestGenerateUsesSelectedBanner(t *testing.T) {
	tests := []struct {
		banner string
		marker string
	}{
		{BannerStandard, "#"},
		{BannerShadow, "$"},
		{BannerThinkertoy, "%"},
	}

	app := generateTestApp()
	for _, test := range tests {
		t.Run(test.banner, func(t *testing.T) {
			in := baseInput()
			in.Banner = test.banner

			got, err := app.Generate(in)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if !strings.Contains(got.ANSIText, test.marker) {
				t.Errorf("output does not use the %s banner glyphs", test.banner)
			}
		})
	}
}

func TestGenerateRejectsUnknownBanner(t *testing.T) {
	in := baseInput()
	in.Banner = "gothic"

	if _, err := generateTestApp().Generate(in); err == nil {
		t.Fatal("Generate() error = nil, want an unknown-banner error")
	}
}

func TestGenerateRejectsUnsupportedCharacter(t *testing.T) {
	in := baseInput()
	in.Text = "Hi\u2603"

	if _, err := generateTestApp().Generate(in); err == nil {
		t.Fatal("Generate() error = nil, want an unsupported-character error")
	}
}

func TestGenerateRendersEachLogicalLineAsEightRows(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantRows int
	}{
		{name: "single line", text: "Hi", wantRows: 8},
		{name: "real newline", text: "Hi\nYo", wantRows: 16},
		{name: "literal newline", text: `Hi\nYo`, wantRows: 16},
		{name: "blank logical line", text: "Hi\n\nYo", wantRows: 17},
	}

	app := generateTestApp()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			in := baseInput()
			in.Text = test.text

			got, err := app.Generate(in)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			// WriteASCII terminates every row, so the trailing newline yields
			// one empty trailing field that is not a rendered row.
			rows := strings.Split(got.ANSIText, "\n")
			rows = rows[:len(rows)-1]
			if len(rows) != test.wantRows {
				t.Errorf("rows = %d, want %d", len(rows), test.wantRows)
			}
		})
	}
}

func TestGenerateAppliesAlignmentAtRequestedWidth(t *testing.T) {
	// "Hi" is 2 characters, each 2 columns wide, so every row is 4 columns.
	const rowWidth = 4
	const width = 20

	tests := []struct {
		alignment   string
		wantPadding int
	}{
		{alignment: "left", wantPadding: 0},
		{alignment: "center", wantPadding: (width - rowWidth) / 2},
		{alignment: "right", wantPadding: width - rowWidth},
	}

	app := generateTestApp()
	for _, test := range tests {
		t.Run(test.alignment, func(t *testing.T) {
			in := baseInput()
			in.Alignment = test.alignment
			in.Width = width

			got, err := app.Generate(in)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if got.Width != width {
				t.Errorf("Width = %d, want %d", got.Width, width)
			}

			firstRow := strings.SplitN(got.ANSIText, "\n", 2)[0]
			padding := len(firstRow) - len(strings.TrimLeft(firstRow, " "))
			if padding != test.wantPadding {
				t.Errorf("leading padding = %d, want %d", padding, test.wantPadding)
			}
		})
	}
}

func TestGenerateJustifyExpandsWordGaps(t *testing.T) {
	in := baseInput()
	in.Text = "Hi Yo"
	in.Alignment = "justify"
	in.Width = 40

	got, err := generateTestApp().Generate(in)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	firstRow := strings.SplitN(got.ANSIText, "\n", 2)[0]
	if len(firstRow) != in.Width {
		t.Errorf("justified row width = %d, want %d", len(firstRow), in.Width)
	}
}

func TestGenerateRejectsUnsupportedAlignment(t *testing.T) {
	in := baseInput()
	in.Alignment = "diagonal"

	if _, err := generateTestApp().Generate(in); err == nil {
		t.Fatal("Generate() error = nil, want an unsupported-alignment error")
	}
}

func TestGenerateColorIsOptIn(t *testing.T) {
	const ansiPrefix = "\033["

	tests := []struct {
		name      string
		color     SelectedColor
		substring string
		wantANSI  bool
	}{
		{
			name:     "disabled colour is ignored",
			color:    SelectedColor{Enabled: false, Hex: "#ff0000"},
			wantANSI: false,
		},
		{
			name:     "enabled colour colours the whole text",
			color:    SelectedColor{Enabled: true, Hex: "#ff0000"},
			wantANSI: true,
		},
		{
			name:      "substring is ignored while colour is disabled",
			color:     SelectedColor{Enabled: false, Hex: "#ff0000"},
			substring: "H",
			wantANSI:  false,
		},
		{
			name:      "enabled colour applies to the substring",
			color:     SelectedColor{Enabled: true, Hex: "#00ff88"},
			substring: "H",
			wantANSI:  true,
		},
	}

	app := generateTestApp()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			in := baseInput()
			in.Color = test.color
			in.Substring = test.substring

			got, err := app.Generate(in)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if strings.Contains(got.ANSIText, ansiPrefix) != test.wantANSI {
				t.Errorf("ANSI present = %t, want %t", !test.wantANSI, test.wantANSI)
			}
		})
	}
}

func TestGenerateRejectsInvalidColor(t *testing.T) {
	in := baseInput()
	in.Color = SelectedColor{Enabled: true, Hex: "#zzzzzz"}

	if _, err := generateTestApp().Generate(in); err == nil {
		t.Fatal("Generate() error = nil, want an invalid-colour error")
	}
}
