package output

import (
	"os"
	"testing"
)

func TestTerminalWidthFromRegularFileIsUnavailable(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "not-a-terminal-*")
	if err != nil {
		t.Fatalf("create temporary file: %v", err)
	}
	defer file.Close()

	width, ok := terminalWidthFromFile(file)
	if ok || width != 0 {
		t.Fatalf("expected ok: false and width: 0")
	}
}
