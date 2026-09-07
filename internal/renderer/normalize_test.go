package renderer

import (
	"reflect"
	"testing"
)

func TestNormalizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple string",
			input:    "Hello",
			expected: []string{"Hello"},
		},
		{
			name:     "single newline",
			input:    "Hello\\nThere",
			expected: []string{"Hello", "There"},
		},
		{
			name:     "double newline",
			input:    "Hello\\n\\nThere",
			expected: []string{"Hello", "", "There"},
		},
		{
			name:     "only newline",
			input:    "\\n",
			expected: []string{""},
		},
		{
			name:     "trailing newline",
			input:    "Hello\\n",
			expected: []string{"Hello"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "real LF splits like literal backslash-n",
			input:    "Hello\nThere",
			expected: []string{"Hello", "There"},
		},
		{
			name:     "CRLF splits into logical lines",
			input:    "Hello\r\nThere",
			expected: []string{"Hello", "There"},
		},
		{
			name:     "lone CR splits into logical lines",
			input:    "Hello\rThere",
			expected: []string{"Hello", "There"},
		},
		{
			name:     "double CRLF produces an empty logical line",
			input:    "Hello\r\n\r\nThere",
			expected: []string{"Hello", "", "There"},
		},
		{
			name:     "trailing CRLF is dropped like a trailing literal newline",
			input:    "Hello\r\n",
			expected: []string{"Hello"},
		},
		{
			name:     "trailing lone CR is dropped like a trailing literal newline",
			input:    "Hello\r",
			expected: []string{"Hello"},
		},
		{
			name:     "mixed real newline and literal newline",
			input:    "Hello\n\\nThere",
			expected: []string{"Hello", "", "There"},
		},
		{
			name:     "mixed CRLF and literal newline",
			input:    "Hello\r\n\\nThere",
			expected: []string{"Hello", "", "There"},
		},
		{
			name:     "leading and repeated spaces are preserved",
			input:    "  Hello\r\nThere   World",
			expected: []string{"  Hello", "There   World"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeInput(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("NormalizeInput(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
