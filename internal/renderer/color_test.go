// Package renderer tests ANSI color parsing and application.
package renderer

import "testing"

func TestGetANSIColorNamedColors(t *testing.T) {
	tests := []struct {
		name     string
		color    string
		expected string
	}{
		{name: "black", color: "black", expected: "\033[30m"},
		{name: "red", color: "red", expected: "\033[31m"},
		{name: "green", color: "green", expected: "\033[32m"},
		{name: "yellow", color: "yellow", expected: "\033[33m"},
		{name: "blue", color: "blue", expected: "\033[34m"},
		{name: "magenta", color: "magenta", expected: "\033[35m"},
		{name: "cyan", color: "cyan", expected: "\033[36m"},
		{name: "white", color: "white", expected: "\033[37m"},
		{name: "orange", color: "orange", expected: "\033[38;5;208m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getANSIColor(tt.color)
			if err != nil {
				t.Fatalf("getANSIColor(%q) returned unexpected error: %v", tt.color, err)
			}
			if result != tt.expected {
				t.Fatalf("getANSIColor(%q) = %q, expected %q", tt.color, result, tt.expected)
			}
		})
	}
}

func TestGetANSIColorHexColors(t *testing.T) {
	tests := []struct {
		name     string
		color    string
		expected string
	}{
		{name: "lowercase red", color: "#ff0000", expected: "\033[38;2;255;0;0m"},
		{name: "uppercase red", color: "#FF0000", expected: "\033[38;2;255;0;0m"},
		{name: "mixed case green blue", color: "#00Ff88", expected: "\033[38;2;0;255;136m"},
		{name: "black", color: "#000000", expected: "\033[38;2;0;0;0m"},
		{name: "white", color: "#ffffff", expected: "\033[38;2;255;255;255m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getANSIColor(tt.color)
			if err != nil {
				t.Fatalf("getANSIColor(%q) returned unexpected error: %v", tt.color, err)
			}
			if result != tt.expected {
				t.Fatalf("getANSIColor(%q) = %q, expected %q", tt.color, result, tt.expected)
			}
		})
	}
}

func TestGetANSIColorRGBColors(t *testing.T) {
	tests := []struct {
		name     string
		color    string
		expected string
	}{
		{name: "lowercase red", color: "rgb,255,0,0", expected: "\033[38;2;255;0;0m"},
		{name: "uppercase green", color: "RGB,0,255,0", expected: "\033[38;2;0;255;0m"},
		{name: "mixed case blue", color: "RgB,0,0,255", expected: "\033[38;2;0;0;255m"},
		{name: "spaces are accepted inside one argument", color: "rgb, 255, 170, 0", expected: "\033[38;2;255;170;0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getANSIColor(tt.color)
			if err != nil {
				t.Fatalf("getANSIColor(%q) returned unexpected error: %v", tt.color, err)
			}
			if result != tt.expected {
				t.Fatalf("getANSIColor(%q) = %q, expected %q", tt.color, result, tt.expected)
			}
		})
	}
}

func TestGetANSIColorHSLColors(t *testing.T) {
	tests := []struct {
		name     string
		color    string
		expected string
	}{
		{name: "lowercase red", color: "hsl,0,100%,50%", expected: "\033[38;2;255;0;0m"},
		{name: "uppercase green", color: "HSL,120,100%,50%", expected: "\033[38;2;0;255;0m"},
		{name: "mixed case blue", color: "HsL,240,100%,50%", expected: "\033[38;2;0;0;255m"},
		{name: "hue 360 equals red", color: "hsl,360,100%,50%", expected: "\033[38;2;255;0;0m"},
		{name: "spaces are accepted inside one argument", color: "hsl, 0, 0%, 50%", expected: "\033[38;2;128;128;128m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getANSIColor(tt.color)
			if err != nil {
				t.Fatalf("getANSIColor(%q) returned unexpected error: %v", tt.color, err)
			}
			if result != tt.expected {
				t.Fatalf("getANSIColor(%q) = %q, expected %q", tt.color, result, tt.expected)
			}
		})
	}
}

func TestGetANSIColorInvalidNames(t *testing.T) {
	tests := []string{
		"",
		"   ",
		"notacolor",
		"RED",
		"ansi:31",
		" #ff0000",
	}

	for _, color := range tests {
		t.Run(color, func(t *testing.T) {
			result, err := getANSIColor(color)
			if err == nil {
				t.Fatalf("getANSIColor(%q) expected error, got nil", color)
			}
			if err.Error() != invalidColorName {
				t.Fatalf("getANSIColor(%q) error = %q, expected %q", color, err.Error(), invalidColorName)
			}
			if result != "" {
				t.Fatalf("getANSIColor(%q) = %q, expected empty string", color, result)
			}
		})
	}
}

func TestGetANSIColorInvalidRGBCodes(t *testing.T) {
	tests := []string{
		"rgb",
		"rgbx,255,0,0",
		"rgb,255,0",
		"rgb,255,0,0,1",
		"rgb,255,,0",
		"rgb,255, ,0",
		"rgb,256,0,0",
		"rgb,-1,0,0",
		"rgb,255.5,0,0",
		"rgb,100%,0,0",
		"rgb,255,0,x",
		"rgb(255,0,0)",
		"rgb(255, 0, 0)",
		"rgb;255;0;0",
	}

	for _, color := range tests {
		t.Run(color, func(t *testing.T) {
			result, err := getANSIColor(color)
			if err == nil {
				t.Fatalf("getANSIColor(%q) expected error, got nil", color)
			}
			if err.Error() != invalidRGBColorCode {
				t.Fatalf("getANSIColor(%q) error = %q, expected %q", color, err.Error(), invalidRGBColorCode)
			}
			if result != "" {
				t.Fatalf("getANSIColor(%q) = %q, expected empty string", color, result)
			}
		})
	}
}

func TestGetANSIColorInvalidHSLCodes(t *testing.T) {
	tests := []string{
		"hsl",
		"hslx,0,100%,50%",
		"hsl,0,100%",
		"hsl,0,100%,50%,1",
		"hsl,0,,50%",
		"hsl,0, ,50%",
		"hsl,361,100%,50%",
		"hsl,720,100%,50%",
		"hsl,-1,100%,50%",
		"hsl,0,101%,50%",
		"hsl,0,100%,101%",
		"hsl,0,100,50%",
		"hsl,0,100%,50",
		"hsl,0,100.5%,50%",
		"hsl(0,100%,50%)",
		"hsl(0, 100%, 50%)",
		"hsl;0;100%;50%",
	}

	for _, color := range tests {
		t.Run(color, func(t *testing.T) {
			result, err := getANSIColor(color)
			if err == nil {
				t.Fatalf("getANSIColor(%q) expected error, got nil", color)
			}
			if err.Error() != invalidHSLColorCode {
				t.Fatalf("getANSIColor(%q) error = %q, expected %q", color, err.Error(), invalidHSLColorCode)
			}
			if result != "" {
				t.Fatalf("getANSIColor(%q) = %q, expected empty string", color, result)
			}
		})
	}
}

func TestGetANSIColorInvalidHexCodes(t *testing.T) {
	tests := []string{
		"#f00",
		"#ff000",
		"#ff00000",
		"#ff 000",
		"#gg0000",
		"#12345x",
		"#",
	}

	for _, color := range tests {
		t.Run(color, func(t *testing.T) {
			result, err := getANSIColor(color)
			if err == nil {
				t.Fatalf("getANSIColor(%q) expected error, got nil", color)
			}
			if err.Error() != invalidHexColorCode {
				t.Fatalf("getANSIColor(%q) error = %q, expected %q", color, err.Error(), invalidHexColorCode)
			}
			if result != "" {
				t.Fatalf("getANSIColor(%q) = %q, expected empty string", color, result)
			}
		})
	}
}

func TestColorizeWrapsTextWithColorAndReset(t *testing.T) {
	result := colorize("abc", "\033[31m")
	expected := "\033[31mabc\033[0m"

	if result != expected {
		t.Fatalf("colorize() = %q, expected %q", result, expected)
	}
}

func TestColorizeInvalidColorReturnsUnchangedText(t *testing.T) {
	result := colorize("abc", "")
	expected := "abc"

	if result != expected {
		t.Fatalf("colorize() = %q, expected %q", result, expected)
	}
}

func TestColorizeEmptyText(t *testing.T) {
	result := colorize("", "\033[31m")
	expected := "\033[31m\033[0m"

	if result != expected {
		t.Fatalf("colorize() = %q, expected %q", result, expected)
	}
}

func TestColorizeIncludesResetCode(t *testing.T) {
	result := colorize("abc", "\033[31m")

	if result[len(result)-len(ansiReset):] != ansiReset {
		t.Fatalf("colorize() = %q, expected reset suffix %q", result, ansiReset)
	}
}
