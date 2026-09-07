package web

import (
	"reflect"
	"strings"
	"testing"
)

func TestPresentPlainText(t *testing.T) {
	input := "ABC\n"

	lines, plain, err := Present(GeneratedASCII{ANSIText: input, Width: 80})
	if err != nil {
		t.Fatalf("Present() error = %v", err)
	}

	wantLines := []StyledLine{
		{Runs: []StyledRun{{Text: "ABC", Colored: false}}},
	}
	if !reflect.DeepEqual(lines, wantLines) {
		t.Errorf("lines = %#v, want %#v", lines, wantLines)
	}
	if plain != input {
		t.Errorf("plain = %q, want %q", plain, input)
	}
}

func TestPresentSplitsAndPreservesBlankLines(t *testing.T) {
	input := "A\n\nB\n"

	lines, plain, err := Present(GeneratedASCII{ANSIText: input, Width: 80})
	if err != nil {
		t.Fatalf("Present() error = %v", err)
	}

	wantLines := []StyledLine{
		{Runs: []StyledRun{{Text: "A", Colored: false}}},
		{},
		{Runs: []StyledRun{{Text: "B", Colored: false}}},
	}
	if !reflect.DeepEqual(lines, wantLines) {
		t.Errorf("lines = %#v, want %#v", lines, wantLines)
	}
	if plain != input {
		t.Errorf("plain = %q, want %q", plain, input)
	}
}

func TestPresentClassifiesColoredRunsAndStripsANSI(t *testing.T) {
	input := "A\033[31mB\033[0mC\n"

	lines, plain, err := Present(GeneratedASCII{ANSIText: input, Width: 80})
	if err != nil {
		t.Fatalf("Present() error = %v", err)
	}

	wantLines := []StyledLine{
		{Runs: []StyledRun{
			{Text: "A", Colored: false},
			{Text: "B", Colored: true},
			{Text: "C", Colored: false},
		}},
	}
	if !reflect.DeepEqual(lines, wantLines) {
		t.Errorf("lines = %#v, want %#v", lines, wantLines)
	}
	if plain != "ABC\n" {
		t.Errorf("plain = %q, want %q", plain, "ABC\n")
	}
	assertNoANSI(t, plain, lines)
}

func TestPresentAcceptsRendererColorSequences(t *testing.T) {
	codes := []string{
		"\033[30m",
		"\033[31m",
		"\033[32m",
		"\033[33m",
		"\033[34m",
		"\033[35m",
		"\033[36m",
		"\033[37m",
		"\033[38;5;208m",
		"\033[38;2;0;0;0m",
		"\033[38;2;255;128;64m",
		"\033[38;2;255;255;255m",
	}

	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			lines, plain, err := Present(GeneratedASCII{ANSIText: code + "X\033[0m\n", Width: 80})
			if err != nil {
				t.Fatalf("Present() error = %v", err)
			}

			wantLines := []StyledLine{
				{Runs: []StyledRun{{Text: "X", Colored: true}}},
			}
			if !reflect.DeepEqual(lines, wantLines) {
				t.Errorf("lines = %#v, want %#v", lines, wantLines)
			}
			if plain != "X\n" {
				t.Errorf("plain = %q, want X newline", plain)
			}
			assertNoANSI(t, plain, lines)
		})
	}
}

func TestPresentRejectsUnsupportedANSI(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "unsupported command", input: "A\033[999mB\n"},
		{name: "incomplete sequence", input: "A\033[31"},
		{name: "truecolor above range", input: "A\033[38;2;256;0;0mB\n"},
		{name: "truecolor below range", input: "A\033[38;2;-1;0;0mB\n"},
		{name: "truecolor missing component", input: "A\033[38;2;255;0mB\n"},
		{name: "truecolor empty component", input: "A\033[38;2;255;;0mB\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := Present(GeneratedASCII{ANSIText: test.input, Width: 80})
			if err == nil {
				t.Fatal("Present() error = nil, want an ANSI validation error")
			}
		})
	}
}

func assertNoANSI(t *testing.T, plain string, lines []StyledLine) {
	t.Helper()
	if strings.Contains(plain, "\033") {
		t.Errorf("plain contains ANSI escape sequence: %q", plain)
	}
	for _, line := range lines {
		for _, run := range line.Runs {
			if strings.Contains(run.Text, "\033") {
				t.Errorf("run text contains ANSI escape sequence: %q", run.Text)
			}
		}
	}
}
