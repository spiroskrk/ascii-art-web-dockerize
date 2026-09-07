// Package renderer converts normalized input text into structured ASCII-art rows.
package renderer

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const ansiReset = "\033[0m"

const invalidHexColorCode = "Invalid hexadecimal color code"
const invalidRGBColorCode = "Invalid RGB color code"
const invalidHSLColorCode = "Invalid HSL color code"
const invalidColorName = "Invalid color name"

var ansiColors = map[string]string{
	"black":   "\033[30m",
	"red":     "\033[31m",
	"green":   "\033[32m",
	"yellow":  "\033[33m",
	"blue":    "\033[34m",
	"magenta": "\033[35m",
	"cyan":    "\033[36m",
	"white":   "\033[37m",
	"orange":  "\033[38;5;208m",
}

// getANSIColor resolves supported color notation to an ANSI foreground escape code.
func getANSIColor(color string) (string, error) {
	colorCode, ok := ansiColors[color]
	if ok {
		return colorCode, nil
	}

	if strings.HasPrefix(color, "#") {
		return getANSIColorFromHex(color)
	}

	lowerColor := strings.ToLower(color)
	if strings.HasPrefix(lowerColor, "rgb") {
		return getANSIColorFromRGB(color)
	}
	if strings.HasPrefix(lowerColor, "hsl") {
		return getANSIColorFromHSL(color)
	}

	return "", errors.New(invalidColorName)
}

// getANSIColorFromHex parses #RRGGBB notation.
func getANSIColorFromHex(color string) (string, error) {
	if len(color) != 7 {
		return "", errors.New(invalidHexColorCode)
	}

	red, err := parseHexColorPart(color[1:3])
	if err != nil {
		return "", errors.New(invalidHexColorCode)
	}
	green, err := parseHexColorPart(color[3:5])
	if err != nil {
		return "", errors.New(invalidHexColorCode)
	}
	blue, err := parseHexColorPart(color[5:7])
	if err != nil {
		return "", errors.New(invalidHexColorCode)
	}

	return formatTrueColor(int(red), int(green), int(blue)), nil
}

func parseHexColorPart(part string) (uint64, error) {
	return strconv.ParseUint(part, 16, 8)
}

// getANSIColorFromRGB parses rgb,<red>,<green>,<blue> notation.
func getANSIColorFromRGB(color string) (string, error) {
	parts := strings.Split(color, ",")
	if len(parts) != 4 || !strings.EqualFold(strings.TrimSpace(parts[0]), "rgb") {
		return "", errors.New(invalidRGBColorCode)
	}

	red, err := parseRGBColorPart(parts[1])
	if err != nil {
		return "", errors.New(invalidRGBColorCode)
	}
	green, err := parseRGBColorPart(parts[2])
	if err != nil {
		return "", errors.New(invalidRGBColorCode)
	}
	blue, err := parseRGBColorPart(parts[3])
	if err != nil {
		return "", errors.New(invalidRGBColorCode)
	}

	return formatTrueColor(red, green, blue), nil
}

func parseRGBColorPart(part string) (int, error) {
	part = strings.TrimSpace(part)
	if !isDigits(part) {
		return 0, errors.New(invalidRGBColorCode)
	}

	value, err := strconv.Atoi(part)
	if err != nil || value < 0 || value > 255 {
		return 0, errors.New(invalidRGBColorCode)
	}

	return value, nil
}

// getANSIColorFromHSL parses hsl,<hue>,<saturation>%,<lightness>% notation.
func getANSIColorFromHSL(color string) (string, error) {
	parts := strings.Split(color, ",")
	if len(parts) != 4 || !strings.EqualFold(strings.TrimSpace(parts[0]), "hsl") {
		return "", errors.New(invalidHSLColorCode)
	}

	hue, err := parseHSLHue(parts[1])
	if err != nil {
		return "", errors.New(invalidHSLColorCode)
	}
	saturation, err := parseHSLPercent(parts[2])
	if err != nil {
		return "", errors.New(invalidHSLColorCode)
	}
	lightness, err := parseHSLPercent(parts[3])
	if err != nil {
		return "", errors.New(invalidHSLColorCode)
	}

	red, green, blue := hslToRGB(hue, saturation, lightness)
	return formatTrueColor(red, green, blue), nil
}

func parseHSLHue(part string) (int, error) {
	part = strings.TrimSpace(part)
	if !isDigits(part) {
		return 0, errors.New(invalidHSLColorCode)
	}

	value, err := strconv.Atoi(part)
	if err != nil || value < 0 || value > 360 {
		return 0, errors.New(invalidHSLColorCode)
	}

	return value, nil
}

func parseHSLPercent(part string) (int, error) {
	part = strings.TrimSpace(part)
	if !strings.HasSuffix(part, "%") {
		return 0, errors.New(invalidHSLColorCode)
	}

	valuePart := strings.TrimSuffix(part, "%")
	if !isDigits(valuePart) {
		return 0, errors.New(invalidHSLColorCode)
	}

	value, err := strconv.Atoi(valuePart)
	if err != nil || value < 0 || value > 100 {
		return 0, errors.New(invalidHSLColorCode)
	}

	return value, nil
}

// hslToRGB converts integer HSL components into RGB components.
func hslToRGB(hue int, saturation int, lightness int) (int, int, int) {
	h := float64(hue)
	if hue == 360 {
		h = 0
	}
	s := float64(saturation) / 100
	l := float64(lightness) / 100

	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	var red, green, blue float64
	switch {
	case h < 60:
		red, green, blue = c, x, 0
	case h < 120:
		red, green, blue = x, c, 0
	case h < 180:
		red, green, blue = 0, c, x
	case h < 240:
		red, green, blue = 0, x, c
	case h < 300:
		red, green, blue = x, 0, c
	default:
		red, green, blue = c, 0, x
	}

	return roundColorPart(red + m), roundColorPart(green + m), roundColorPart(blue + m)
}

func roundColorPart(value float64) int {
	return int(math.Round(value * 255))
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// formatTrueColor formats RGB components as an ANSI truecolor escape sequence.
func formatTrueColor(red int, green int, blue int) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", red, green, blue)
}

// colorize wraps rendered ASCII text with the selected ANSI color and reset code.
func colorize(text string, colorCode string) string {
	if colorCode == "" {
		return text
	}

	return colorCode + text + ansiReset
}
