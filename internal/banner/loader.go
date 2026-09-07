// Package banner loads the official ASCII-art banner files.
package banner

import (
	"errors"
	"os"
	"strings"
)

// LoadBanner reads a banner file and maps printable ASCII runes to their 8-line art blocks.
func LoadBanner(path string) (map[rune][]string, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}
	m := make(map[rune][]string)
	content := string(data)
	content = strings.ReplaceAll(content, "\r\n", "\n")
	trimmedContent := strings.TrimSuffix(content, "\n")
	lines := strings.Split(trimmedContent, "\n")
	printableChCount := (126 - 32) + 1
	linesPerBlock := 9
	visibleArtLines := 8
	expectedLines := printableChCount * linesPerBlock
	if len(lines) != expectedLines {
		return nil, errors.New("invalid banner format")
	}
	for ascii := 32; ascii <= 126; ascii++ {
		blockStart := (ascii - 32) * linesPerBlock
		artStart := blockStart + 1
		eightLines := lines[artStart : artStart+visibleArtLines]
		a := rune(ascii)
		m[a] = eightLines

	}

	return m, nil
}
