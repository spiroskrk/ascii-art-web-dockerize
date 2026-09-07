// Package renderer converts normalized input text into structured ASCII-art rows.
package renderer

// coloredPositions marks byte positions that should receive color in a single input line.
func coloredPositions(s string, substring string) map[int]bool {
	positions := make(map[int]bool)

	if substring == "" {
		for i := range s {
			positions[i] = true
		}
		return positions
	}
	for i := 0; i < len(substring); i++ {
		if substring[i] == '\n' {
			return positions
		}
	}
	for i := 0; i+len(substring) <= len(s); i++ {
		if substring == s[i:i+len(substring)] {
			for offset := 0; offset < len(substring); offset++ {
				positions[i+offset] = true
			}
		}
	}
	return positions
}
