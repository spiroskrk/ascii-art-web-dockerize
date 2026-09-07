package web

import (
	"errors"
	"strconv"
	"strings"
)

// Present converts renderer-controlled ANSI text into safe browser runs and
// ANSI-free download text. Package C implements it after Phase 0.
func Present(g GeneratedASCII) (lines []StyledLine, plain string, err error) {
	// Keep browser runs and download text in sync: visible bytes go to both,
	// while ANSI control bytes are interpreted and skipped.
	var plainBuilder strings.Builder
	var runBuilder strings.Builder
	var currentLine StyledLine
	colored := false

	flushRun := func() {
		if runBuilder.Len() == 0 {
			return
		}
		// Flush before any color-state change so the completed run keeps the
		// state it was collected under.
		currentLine.Runs = append(currentLine.Runs, StyledRun{
			Text:    runBuilder.String(),
			Colored: colored,
		})
		runBuilder.Reset()
	}
	flushLine := func() {
		flushRun()
		lines = append(lines, currentLine)
		currentLine = StyledLine{}
	}

	text := g.ANSIText
	for i := 0; i < len(text); {
		ch := text[i]
		if ch == '\n' {
			plainBuilder.WriteByte(ch)

			flushLine()
			i++
			continue
		}
		if ch == '\033' {
			// ANSI is renderer-owned control data. It toggles Colored state,
			// but must never leak into HTML text or download output.
			end := i + 1
			for end < len(text) && text[end] != 'm' {
				end++
			}
			if end >= len(text) {
				plain = plainBuilder.String()
				return lines, plain, errors.New("incomplete ANSI escape sequence")
			}
			sequence := text[i : end+1]
			flushRun()
			if sequence == "\033[0m" {
				colored = false
			} else if isSupportedANSIColor(sequence) {
				colored = true
			} else {
				plain = plainBuilder.String()
				return lines, plain, errors.New("unsupported ANSI escape sequence")
			}
			i = end + 1
			continue

		}
		plainBuilder.WriteByte(ch)
		runBuilder.WriteByte(ch)
		i++

	}
	if runBuilder.Len() > 0 || len(currentLine.Runs) > 0 {
		flushLine()
	}
	plain = plainBuilder.String()
	return
}

func isSupportedANSIColor(sequence string) bool {
	// Accept only the exact foreground sequences the renderer can emit.
	switch sequence {
	case "\033[30m", "\033[31m", "\033[32m", "\033[33m":
		return true
	case "\033[34m", "\033[35m", "\033[36m", "\033[37m":
		return true
	case "\033[38;5;208m":
		return true
	default:
		return isSupportedTrueColorSequence(sequence)
	}
}

func isSupportedTrueColorSequence(sequence string) bool {
	prefix := "\033[38;2;"
	if !strings.HasPrefix(sequence, prefix) {
		return false
	}
	// Truecolor uses three decimal RGB components after the fixed prefix.
	body := strings.TrimPrefix(sequence, prefix)
	body = strings.TrimSuffix(body, "m")
	parts := strings.Split(body, ";")
	if len(parts) != 3 {
		return false
	}
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		value, err := strconv.Atoi(part)
		if err != nil {
			return false
		}
		if value < 0 || value > 255 {
			return false
		}
	}
	return true
}

func selectedColorFromHex(hex string) (SelectedColor, error) {
	if !isWebHexColor(hex) {
		return SelectedColor{}, errors.New("unsupported hex color")
	}
	// Each #RRGGBB pair is parsed as one 8-bit channel before the values are
	// used in inline CSS.
	red, err := strconv.ParseUint(hex[1:3], 16, 8)
	if err != nil {
		return SelectedColor{}, errors.New("unsupported hex color")
	}
	green, err := strconv.ParseUint(hex[3:5], 16, 8)
	if err != nil {
		return SelectedColor{}, errors.New("unsupported hex color")
	}
	blue, err := strconv.ParseUint(hex[5:7], 16, 8)
	if err != nil {
		return SelectedColor{}, errors.New("unsupported hex color")
	}

	return SelectedColor{
		Enabled: true,
		Hex:     hex,
		Red:     int(red),
		Green:   int(green),
		Blue:    int(blue),
	}, nil
}
