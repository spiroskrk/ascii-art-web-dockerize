// Package renderer tests structured ASCII-art rendering.
package renderer

import (
	"reflect"
	"strings"
	"testing"
)

func mockBanner() map[rune][]string {
	banner := make(map[rune][]string)
	for ch := rune(32); ch <= rune(126); ch++ {
		banner[ch] = []string{
			"line0",
			"line1",
			"line2",
			"line3",
			"line4",
			"line5",
			"line6",
			"line7",
		}
	}
	return banner
}

func TestBuildASCII(t *testing.T) {
	banner := mockBanner()

	t.Run("single line", func(t *testing.T) {
		result, err := BuildASCII([]string{"Hi"}, banner, "", "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(Flatten(result)) != 8 {
			t.Errorf("expected 8 lines, got %d", len(Flatten(result)))
		}
	})

	t.Run("empty line", func(t *testing.T) {
		result, err := BuildASCII([]string{""}, banner, "", "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(Flatten(result)) != 1 {
			t.Errorf("expected 1 line, got %d", len(Flatten(result)))
		}
	})

	t.Run("unsupported character", func(t *testing.T) {
		_, err := BuildASCII([]string{"Héllo"}, banner, "", "")
		if err == nil {
			t.Error("expected error for unsupported character")
		}
	})

	t.Run("multiple lines", func(t *testing.T) {
		result, err := BuildASCII([]string{"Hi", "Hi"}, banner, "", "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(Flatten(result)) != 16 {
			t.Errorf("expected 16 lines, got %d", len(Flatten(result)))
		}
	})
}

func TestBuildASCIICombinesCharactersHorizontally(t *testing.T) {
	banner := map[rune][]string{
		'A': {"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7"},
		'B': {"B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7"},
	}

	result, err := BuildASCII([]string{"AB"}, banner, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"A0B0",
		"A1B1",
		"A2B2",
		"A3B3",
		"A4B4",
		"A5B5",
		"A6B6",
		"A7B7",
	}

	if !reflect.DeepEqual(Flatten(result), expected) {
		t.Fatalf("Flatten result = %v, expected %v", Flatten(result), expected)
	}
}

func TestBuildASCIIKeepsEmptyInputLines(t *testing.T) {
	banner := map[rune][]string{
		'A': {"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7"},
	}

	result, err := BuildASCII([]string{"", "A"}, banner, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"",
		"A0",
		"A1",
		"A2",
		"A3",
		"A4",
		"A5",
		"A6",
		"A7",
	}

	if !reflect.DeepEqual(Flatten(result), expected) {
		t.Fatalf("Flatten result = %v, expected %v", Flatten(result), expected)
	}
}

func TestBuildASCIIProducesWordAndSpaceSegments(t *testing.T) {
	banner := mockBanner()

	result, err := BuildASCII([]string{"hi there"}, banner, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 line, got %d", len(result))
	}

	segments := result[0].Segments
	if len(segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(segments))
	}

	if segments[0].Kind != SegmentWord || segments[0].Text != "hi" {
		t.Errorf("segment 0 = %+v, expected word 'hi'", segments[0])
	}
	if segments[1].Kind != SegmentSpace || segments[1].Text != " " {
		t.Errorf("segment 1 = %+v, expected space ' '", segments[1])
	}
	if segments[2].Kind != SegmentWord || segments[2].Text != "there" {
		t.Errorf("segment 2 = %+v, expected word 'there'", segments[2])
	}
}

func TestBuildASCIIPreservesMultipleSpaces(t *testing.T) {
	banner := mockBanner()

	result, err := BuildASCII([]string{"a   b"}, banner, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	segments := result[0].Segments
	if len(segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(segments))
	}

	if segments[1].Kind != SegmentSpace || segments[1].Text != "   " {
		t.Errorf("segment 1 = %+v, expected space run of 3", segments[1])
	}
}

func TestBuildASCIIPreservesLeadingAndTrailingSpaces(t *testing.T) {
	banner := mockBanner()

	result, err := BuildASCII([]string{"  hi  "}, banner, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	segments := result[0].Segments
	if len(segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(segments))
	}

	if segments[0].Kind != SegmentSpace || segments[0].Text != "  " {
		t.Errorf("segment 0 = %+v, expected leading space run", segments[0])
	}
	if segments[2].Kind != SegmentSpace || segments[2].Text != "  " {
		t.Errorf("segment 2 = %+v, expected trailing space run", segments[2])
	}
}

func TestBuildASCIIUnsupportedColor(t *testing.T) {
	banner := mockBanner()
	_, err := BuildASCII([]string{"Hi"}, banner, "", "notacolor")
	if err == nil {
		t.Error("expected error for invalid color name")
	}
	if err.Error() != invalidColorName {
		t.Fatalf("error = %q, expected %q", err.Error(), invalidColorName)
	}
}

func TestBuildASCIIInvalidHexColor(t *testing.T) {
	banner := mockBanner()
	_, err := BuildASCII([]string{"Hi"}, banner, "", "#f00")
	if err == nil {
		t.Error("expected error for invalid hex color")
	}
	if err.Error() != invalidHexColorCode {
		t.Fatalf("error = %q, expected %q", err.Error(), invalidHexColorCode)
	}
}

func TestBuildASCIIInvalidRGBColor(t *testing.T) {
	banner := mockBanner()
	_, err := BuildASCII([]string{"Hi"}, banner, "", "rgb,256,0,0")
	if err == nil {
		t.Error("expected error for invalid RGB color")
	}
	if err.Error() != invalidRGBColorCode {
		t.Fatalf("error = %q, expected %q", err.Error(), invalidRGBColorCode)
	}
}

func TestBuildASCIIInvalidHSLColor(t *testing.T) {
	banner := mockBanner()
	_, err := BuildASCII([]string{"Hi"}, banner, "", "hsl,361,100%,50%")
	if err == nil {
		t.Error("expected error for invalid HSL color")
	}
	if err.Error() != invalidHSLColorCode {
		t.Fatalf("error = %q, expected %q", err.Error(), invalidHSLColorCode)
	}
}

func TestBuildASCIIColorWholeString(t *testing.T) {
	banner := map[rune][]string{
		'A': {"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7"},
		'B': {"B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7"},
	}

	result, err := BuildASCII([]string{"AB"}, banner, "", "#FF0000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	redCode := "\033[38;2;255;0;0m"

	for i, row := range Flatten(result) {
		if !strings.Contains(row, redCode) {
			t.Errorf("row %d missing hex ANSI color code: %q", i, row)
		}
		if !strings.Contains(row, "\033[0m") {
			t.Errorf("row %d missing reset code: %q", i, row)
		}
	}
}

func TestBuildASCIIColorSubstring(t *testing.T) {
	banner := map[rune][]string{
		'A': {"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7"},
		'B': {"B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7"},
		'C': {"C0", "C1", "C2", "C3", "C4", "C5", "C6", "C7"},
	}

	// Color only "B" inside "ABC"
	result, err := BuildASCII([]string{"ABC"}, banner, "B", "red")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	redCode := "\033[31m"
	reset := "\033[0m"

	for i, row := range Flatten(result) {
		if strings.Contains(row, redCode+"A") || strings.Contains(row, redCode+"C") {
			t.Errorf("row %d: A or C should not be colored: %q", i, row)
		}
		if !strings.Contains(row, redCode) || !strings.Contains(row, reset) {
			t.Errorf("row %d: expected B to be colored: %q", i, row)
		}
	}
}

func TestFlattenReproducesFlatOutput(t *testing.T) {
	banner := map[rune][]string{
		'A': {"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7"},
		'B': {"B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7"},
		' ': {"  ", "  ", "  ", "  ", "  ", "  ", "  ", "  "},
	}

	result, err := BuildASCII([]string{"A B"}, banner, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"A0  B0",
		"A1  B1",
		"A2  B2",
		"A3  B3",
		"A4  B4",
		"A5  B5",
		"A6  B6",
		"A7  B7",
	}

	if !reflect.DeepEqual(Flatten(result), expected) {
		t.Fatalf("Flatten result = %v, expected %v", Flatten(result), expected)
	}
}
