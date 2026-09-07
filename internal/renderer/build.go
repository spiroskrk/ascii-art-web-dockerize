// Package renderer converts normalized input text into structured ASCII-art rows.
package renderer

import "errors"

// BuildASCII renders normalized input lines into structured word and space segments.
func BuildASCII(lines []string, banner map[rune][]string, substring string, color string) (RenderedASCII, error) {
	var colorCode string
	if color != "" {
		var err error
		colorCode, err = getANSIColor(color)
		if err != nil {
			return nil, err
		}
	}

	var output RenderedASCII

	for _, n := range lines {
		if n == "" {
			output = append(output, RenderedLine{})
			continue
		}

		var positions map[int]bool
		if colorCode != "" {
			positions = coloredPositions(n, substring)
		}

		segments, err := renderLineSegments(n, banner, colorCode, positions)
		if err != nil {
			return nil, err
		}

		output = append(output, RenderedLine{Segments: segments})
	}

	return output, nil
}

func renderLineSegments(line string, banner map[rune][]string, colorCode string, positions map[int]bool) ([]RenderedSegment, error) {
	runes := []rune(line)
	var segments []RenderedSegment

	i := 0
	for i < len(runes) {
		// Group contiguous spaces separately so justify can expand only word gaps later.
		isSpace := runes[i] == ' '
		start := i
		for i < len(runes) && (runes[i] == ' ') == isSpace {
			i++
		}
		segRunes := runes[start:i]

		var rows [8]string
		for j, ch := range segRunes {
			charLines, ok := banner[ch]
			if !ok {
				return nil, errors.New("unsupported character")
			}
			pos := start + j
			for r := 0; r < 8; r++ {
				if colorCode != "" && positions[pos] {
					rows[r] += colorize(charLines[r], colorCode)
				} else {
					rows[r] += charLines[r]
				}
			}
		}

		kind := SegmentWord
		if isSpace {
			kind = SegmentSpace
		}

		segments = append(segments, RenderedSegment{
			Kind: kind,
			Text: string(segRunes),
			Rows: rows[:],
		})
	}

	return segments, nil
}

// Flatten reproduces the old flat row output from structured renderer data.
func Flatten(rendered RenderedASCII) []string {
	var output []string

	for _, line := range rendered {
		if len(line.Segments) == 0 {
			output = append(output, "")
			continue
		}

		var rows [8]string
		for _, seg := range line.Segments {
			for r := 0; r < 8; r++ {
				rows[r] += seg.Rows[r]
			}
		}
		for _, row := range rows {
			output = append(output, row)
		}
	}

	return output
}
