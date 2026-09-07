// Package output owns terminal width detection, alignment, and writing rendered ASCII.
package output

import (
	"fmt"
	"io"
)

// WriteASCII writes each flat ASCII row followed by one newline.
func WriteASCII(w io.Writer, lines []string) error {
	for _, line := range lines {
		_, err := fmt.Fprintln(w, line)
		if err != nil {
			return err
		}
	}
	return nil
}
