// Package output tests flat row writing.
package output

import (
	"strings"
	"testing"
)

func TestWriteASCII_writesLinesWithNewlines(t *testing.T) {
	var buf strings.Builder
	lines := []string{"hello", "world"}

	err := WriteASCII(&buf, lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "hello\nworld\n"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestWriteASCII_emptyInput(t *testing.T) {
	var buf strings.Builder

	err := WriteASCII(&buf, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.String() != "" {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}
