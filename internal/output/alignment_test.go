// Package output tests width detection and alignment behavior.
package output

import (
	"ascii-art/internal/renderer"
	"os"
	"strings"
	"testing"
)

func TestVisibleWidth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "plain text",
			input: "hello",
			want:  5,
		},
		{
			name:  "leading spaces",
			input: "  hi",
			want:  4,
		},
		{
			name:  "named ANSI color",
			input: "\033[31mhello\033[0m",
			want:  5,
		},
		{
			name:  "truecolor ANSI color",
			input: "\033[38;2;255;0;0mhello\033[0m",
			want:  5,
		},
		{
			name:  "colored leading spaces",
			input: "\033[36m  hi\033[0m",
			want:  4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := visibleWidth(tt.input)
			if got != tt.want {
				t.Fatalf("visibleWidth(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestTerminalWidth(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     int
	}{
		{
			name:     "valid columns",
			envValue: "120",
			want:     120,
		},
		{
			name:     "default width value",
			envValue: "80",
			want:     80,
		},
		{
			name:     "non numeric falls back",
			envValue: "abc",
			want:     80,
		},
		{
			name:     "zero falls back",
			envValue: "0",
			want:     80,
		},
		{
			name:     "negative falls back",
			envValue: "-5",
			want:     80,
		},
		{
			name:     "empty columns falls back",
			envValue: "",
			want:     80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("COLUMNS", tt.envValue)
			got := terminalWidth()
			if got != tt.want {
				t.Fatalf("terminalWidth() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTerminalWidthMissingColumnsFallsBack(t *testing.T) {
	oldValue, hadValue := os.LookupEnv("COLUMNS")
	t.Cleanup(func() {
		if hadValue {
			os.Setenv("COLUMNS", oldValue)
			return
		}
		os.Unsetenv("COLUMNS")
	})

	os.Unsetenv("COLUMNS")

	got := terminalWidth()
	if got != 80 {
		t.Fatalf("terminalWidth() = %d, want 80", got)
	}
}

func TestWriteAlignedASCIILeftPreservesFlattenedRows(t *testing.T) {
	var writer strings.Builder
	rendered := renderer.RenderedASCII{
		renderer.RenderedLine{
			Segments: []renderer.RenderedSegment{
				{
					Kind: renderer.SegmentWord,
					Text: "A",
					Rows: []string{"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7"},
				}, {
					Kind: renderer.SegmentSpace,
					Text: " ",
					Rows: []string{" ", " ", " ", " ", " ", " ", " ", " "},
				}, {
					Kind: renderer.SegmentWord,
					Text: "B",
					Rows: []string{"B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7"},
				},
			},
		},
	}
	err := WriteAlignedASCII(&writer, rendered, "left", 80)
	if err != nil {
		t.Fatalf("WriteAlignedASCII returned error: %v", err)
	}
	expected := "A0 B0\n" +
		"A1 B1\n" +
		"A2 B2\n" +
		"A3 B3\n" +
		"A4 B4\n" +
		"A5 B5\n" +
		"A6 B6\n" +
		"A7 B7\n"
	if writer.String() != expected {
		t.Fatalf("got %q, want %q", writer.String(), expected)
	}
}

func TestWriteAlignedASCIIDefaultAlignsLeft(t *testing.T) {
	var writer strings.Builder
	rendered := renderer.RenderedASCII{
		{
			Segments: []renderer.RenderedSegment{
				{
					Kind: renderer.SegmentWord,
					Text: "A",
					Rows: []string{"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7"},
				},
			},
		},
	}

	err := WriteAlignedASCII(&writer, rendered, "", 80)
	if err != nil {
		t.Fatalf("WriteAlignedASCII returned error: %v", err)
	}

	expected := "A0\n" +
		"A1\n" +
		"A2\n" +
		"A3\n" +
		"A4\n" +
		"A5\n" +
		"A6\n" +
		"A7\n"
	if writer.String() != expected {
		t.Fatalf("got %q, want %q", writer.String(), expected)
	}
}

func TestWriteAlignedASCIICenterPadsRows(t *testing.T) {
	var writer strings.Builder
	rendered := renderer.RenderedASCII{
		renderer.RenderedLine{
			Segments: []renderer.RenderedSegment{
				{
					Kind: renderer.SegmentWord,
					Text: "A",
					Rows: []string{"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7"},
				},
			},
		},
	}
	err := WriteAlignedASCII(&writer, rendered, "center", 6)
	if err != nil {
		t.Fatalf("WriteAlignedASCII returned error: %v", err)
	}
	expected := "  A0\n" +
		"  A1\n" +
		"  A2\n" +
		"  A3\n" +
		"  A4\n" +
		"  A5\n" +
		"  A6\n" +
		"  A7\n"

	if writer.String() != expected {
		t.Fatalf("got %q, want %q", writer.String(), expected)
	}

}

func TestWriteAlignedASCIIRightPadsRows(t *testing.T) {
	var writer strings.Builder
	rendered := renderer.RenderedASCII{
		renderer.RenderedLine{
			Segments: []renderer.RenderedSegment{
				{
					Kind: renderer.SegmentWord,
					Text: "A",
					Rows: []string{"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7"},
				},
			},
		},
	}
	err := WriteAlignedASCII(&writer, rendered, "right", 6)
	if err != nil {
		t.Fatalf("WriteAlignedASCII returned error: %v", err)
	}
	expected := "    A0\n" +
		"    A1\n" +
		"    A2\n" +
		"    A3\n" +
		"    A4\n" +
		"    A5\n" +
		"    A6\n" +
		"    A7\n"

	if writer.String() != expected {
		t.Fatalf("got %q, want %q", writer.String(), expected)
	}

}

func TestWriteAlignedASCIIDoesNotPadRowsWiderThanWidth(t *testing.T) {
	var writer strings.Builder
	rendered := renderer.RenderedASCII{
		{
			Segments: []renderer.RenderedSegment{
				{
					Kind: renderer.SegmentWord,
					Text: "A",
					Rows: []string{
						"ABCDE",
						"FGHIJ",
						"KLMNO",
						"PQRST",
						"UVWXY",
						"abcde",
						"fghij",
						"klmno",
					},
				},
			},
		},
	}
	err := WriteAlignedASCII(&writer, rendered, "right", 3)
	if err != nil {
		t.Fatalf("WriteAlignedASCII returned error: %v", err)
	}

	expected := "ABCDE\n" +
		"FGHIJ\n" +
		"KLMNO\n" +
		"PQRST\n" +
		"UVWXY\n" +
		"abcde\n" +
		"fghij\n" +
		"klmno\n"

	if writer.String() != expected {
		t.Fatalf("got %q, want %q", writer.String(), expected)
	}
}

func TestWriteAlignedASCIIJustifyDistributesExtraSpacesLeftToRight(t *testing.T) {
	var writer strings.Builder
	rendered := renderer.RenderedASCII{
		{
			Segments: []renderer.RenderedSegment{
				{
					Kind: renderer.SegmentWord,
					Text: "A",
					Rows: []string{"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7"},
				},
				{
					Kind: renderer.SegmentSpace,
					Text: " ",
					Rows: []string{" ", " ", " ", " ", " ", " ", " ", " "},
				},
				{
					Kind: renderer.SegmentWord,
					Text: "B",
					Rows: []string{"B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7"},
				},
				{
					Kind: renderer.SegmentSpace,
					Text: " ",
					Rows: []string{" ", " ", " ", " ", " ", " ", " ", " "},
				},
				{
					Kind: renderer.SegmentWord,
					Text: "C",
					Rows: []string{"C0", "C1", "C2", "C3", "C4", "C5", "C6", "C7"},
				},
			},
		},
	}

	err := WriteAlignedASCII(&writer, rendered, "justify", 11)
	if err != nil {
		t.Fatalf("WriteAlignedASCII returned error: %v", err)
	}

	expected := "A0   B0  C0\n" +
		"A1   B1  C1\n" +
		"A2   B2  C2\n" +
		"A3   B3  C3\n" +
		"A4   B4  C4\n" +
		"A5   B5  C5\n" +
		"A6   B6  C6\n" +
		"A7   B7  C7\n"
	if writer.String() != expected {
		t.Fatalf("got %q, want %q", writer.String(), expected)
	}
}

func TestWriteAlignedASCIIJustifyLeavesRowsWiderThanWidthUnchanged(t *testing.T) {
	var writer strings.Builder
	rendered := renderer.RenderedASCII{
		{
			Segments: []renderer.RenderedSegment{
				{
					Kind: renderer.SegmentWord,
					Text: "A",
					Rows: []string{"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7"},
				},
				{
					Kind: renderer.SegmentSpace,
					Text: " ",
					Rows: []string{" ", " ", " ", " ", " ", " ", " ", " "},
				},
				{
					Kind: renderer.SegmentWord,
					Text: "B",
					Rows: []string{"B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7"},
				},
			},
		},
	}

	err := WriteAlignedASCII(&writer, rendered, "justify", 4)
	if err != nil {
		t.Fatalf("WriteAlignedASCII returned error: %v", err)
	}

	expected := "A0 B0\n" +
		"A1 B1\n" +
		"A2 B2\n" +
		"A3 B3\n" +
		"A4 B4\n" +
		"A5 B5\n" +
		"A6 B6\n" +
		"A7 B7\n"
	if writer.String() != expected {
		t.Fatalf("got %q, want %q", writer.String(), expected)
	}
}

func TestWriteAlignedASCIIJustifyExpandsMultipleSpaceGaps(t *testing.T) {
	var writer strings.Builder
	rendered := renderer.RenderedASCII{
		{
			Segments: []renderer.RenderedSegment{
				{
					Kind: renderer.SegmentWord,
					Text: "A",
					Rows: []string{"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7"},
				},
				{
					Kind: renderer.SegmentSpace,
					Text: "  ",
					Rows: []string{"  ", "  ", "  ", "  ", "  ", "  ", "  ", "  "},
				},
				{
					Kind: renderer.SegmentWord,
					Text: "B",
					Rows: []string{"B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7"},
				},
			},
		},
	}

	err := WriteAlignedASCII(&writer, rendered, "justify", 8)
	if err != nil {
		t.Fatalf("WriteAlignedASCII returned error: %v", err)
	}

	expected := "A0    B0\n" +
		"A1    B1\n" +
		"A2    B2\n" +
		"A3    B3\n" +
		"A4    B4\n" +
		"A5    B5\n" +
		"A6    B6\n" +
		"A7    B7\n"
	if writer.String() != expected {
		t.Fatalf("got %q, want %q", writer.String(), expected)
	}
}

func TestWriteAlignedASCIIJustifyDoesNotExpandSpacesInsideWordRows(t *testing.T) {
	var writer strings.Builder
	rendered := renderer.RenderedASCII{
		{
			Segments: []renderer.RenderedSegment{
				{
					Kind: renderer.SegmentWord,
					Text: "A",
					Rows: []string{"A 0", "A 1", "A 2", "A 3", "A 4", "A 5", "A 6", "A 7"},
				},
				{
					Kind: renderer.SegmentSpace,
					Text: " ",
					Rows: []string{" ", " ", " ", " ", " ", " ", " ", " "},
				},
				{
					Kind: renderer.SegmentWord,
					Text: "B",
					Rows: []string{"B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7"},
				},
			},
		},
	}

	err := WriteAlignedASCII(&writer, rendered, "justify", 8)
	if err != nil {
		t.Fatalf("WriteAlignedASCII returned error: %v", err)
	}

	expected := "A 0   B0\n" +
		"A 1   B1\n" +
		"A 2   B2\n" +
		"A 3   B3\n" +
		"A 4   B4\n" +
		"A 5   B5\n" +
		"A 6   B6\n" +
		"A 7   B7\n"
	if writer.String() != expected {
		t.Fatalf("got %q, want %q", writer.String(), expected)
	}
}

func TestWriteAlignedASCIIUsesVisibleWidthForANSIColoredRows(t *testing.T) {
	var writer strings.Builder
	rendered := renderer.RenderedASCII{
		{
			Segments: []renderer.RenderedSegment{
				{
					Kind: renderer.SegmentWord,
					Text: "A",
					Rows: []string{
						"\033[31mA0\033[0m",
						"\033[31mA1\033[0m",
						"\033[31mA2\033[0m",
						"\033[31mA3\033[0m",
						"\033[31mA4\033[0m",
						"\033[31mA5\033[0m",
						"\033[31mA6\033[0m",
						"\033[31mA7\033[0m",
					},
				},
			},
		},
	}
	err := WriteAlignedASCII(&writer, rendered, "right", 6)
	if err != nil {
		t.Fatalf("WriteAlignedASCII returned error: %v", err)
	}

	expected := "    \033[31mA0\033[0m\n" +
		"    \033[31mA1\033[0m\n" +
		"    \033[31mA2\033[0m\n" +
		"    \033[31mA3\033[0m\n" +
		"    \033[31mA4\033[0m\n" +
		"    \033[31mA5\033[0m\n" +
		"    \033[31mA6\033[0m\n" +
		"    \033[31mA7\033[0m\n"

	if writer.String() != expected {
		t.Fatalf("got %q, want %q", writer.String(), expected)
	}
}

func TestWriteAlignedASCIIJustifyExpandsGapsBetweenWords(t *testing.T) {
	var writer strings.Builder
	rendered := renderer.RenderedASCII{
		{
			Segments: []renderer.RenderedSegment{
				{
					Kind: renderer.SegmentWord,
					Text: "A",
					Rows: []string{"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7"},
				}, {
					Kind: renderer.SegmentSpace,
					Text: " ",
					Rows: []string{" ", " ", " ", " ", " ", " ", " ", " "},
				}, {
					Kind: renderer.SegmentWord,
					Text: "B",
					Rows: []string{"B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7"},
				},
			},
		},
	}
	err := WriteAlignedASCII(&writer, rendered, "justify", 8)
	if err != nil {
		t.Fatalf("WriteAlignedASCII returned error: %v", err)
	}

	expected := "A0    B0\n" +
		"A1    B1\n" +
		"A2    B2\n" +
		"A3    B3\n" +
		"A4    B4\n" +
		"A5    B5\n" +
		"A6    B6\n" +
		"A7    B7\n"
	if writer.String() != expected {
		t.Fatalf("got %q, want %q", writer.String(), expected)
	}
}

func TestWriteAlignedASCIIJustifyLeavesOneWordLinesUnchanged(t *testing.T) {
	var writer strings.Builder
	rendered := renderer.RenderedASCII{
		{
			Segments: []renderer.RenderedSegment{
				{
					Kind: renderer.SegmentWord,
					Text: "A",
					Rows: []string{"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7"},
				},
			},
		},
	}
	err := WriteAlignedASCII(&writer, rendered, "justify", 8)
	if err != nil {
		t.Fatalf("WriteAlignedASCII returned error: %v", err)
	}
	expected := "A0\n" +
		"A1\n" +
		"A2\n" +
		"A3\n" +
		"A4\n" +
		"A5\n" +
		"A6\n" +
		"A7\n"
	if writer.String() != expected {
		t.Fatalf("got %q, want %q", writer.String(), expected)
	}
}

func TestWriteAlignedASCIIJustifyPreservesLeadingAndTrailingSpaces(t *testing.T) {
	var writer strings.Builder
	rendered := renderer.RenderedASCII{
		{
			Segments: []renderer.RenderedSegment{
				{
					Kind: renderer.SegmentSpace,
					Text: " ",
					Rows: []string{" ", " ", " ", " ", " ", " ", " ", " "},
				},
				{
					Kind: renderer.SegmentWord,
					Text: "A",
					Rows: []string{"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7"},
				},
				{
					Kind: renderer.SegmentSpace,
					Text: " ",
					Rows: []string{" ", " ", " ", " ", " ", " ", " ", " "},
				},
				{
					Kind: renderer.SegmentWord,
					Text: "B",
					Rows: []string{"B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7"},
				},
				{
					Kind: renderer.SegmentSpace,
					Text: " ",
					Rows: []string{" ", " ", " ", " ", " ", " ", " ", " "},
				},
			},
		},
	}

	err := WriteAlignedASCII(&writer, rendered, "justify", 10)
	if err != nil {
		t.Fatalf("WriteAlignedASCII returned error: %v", err)
	}

	expected := " A0    B0 \n" +
		" A1    B1 \n" +
		" A2    B2 \n" +
		" A3    B3 \n" +
		" A4    B4 \n" +
		" A5    B5 \n" +
		" A6    B6 \n" +
		" A7    B7 \n"

	if writer.String() != expected {
		t.Fatalf("got %q, want %q", writer.String(), expected)
	}
}
