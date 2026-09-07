// Package banner tests official banner file loading and validation.
package banner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadBannerMissingFile(t *testing.T) {
	banner, err := LoadBanner("does-not-exist.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if banner != nil {
		t.Fatal("expected a nil banner for a missing file")
	}

}
func TestLoadBannerInvalidFormat(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "banner.txt")
	err := os.WriteFile(filePath, []byte("hello"), 0644)
	if err != nil {
		t.Fatal("writing temp file failed")
	}
	banner, err := LoadBanner(filePath)
	if err == nil {
		t.Fatal("expected error for invalid banner format")
	}
	if banner != nil {
		t.Fatal("expected nil banner for invalid banner format")
	}
}

func TestLoadBannerCorrectFile(t *testing.T) {
	banner, err := LoadBanner("../../banners/standard.txt")
	if err != nil {
		t.Fatal("expected no error for valid banner file")
	}
	if banner == nil {
		t.Fatal("expected non-nil banner for valid banner file")
	}
	if len(banner['A']) != 8 {
		t.Fatal("expected 8 lines for 'A'")
	}
	if len(banner) != 95 {
		t.Fatal("expected 95 mapped characters")
	}
}

func TestLoadBannerOfficialFiles(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "standard", path: "../../banners/standard.txt"},
		{name: "shadow", path: "../../banners/shadow.txt"},
		{name: "thinkertoy", path: "../../banners/thinkertoy.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			banner, err := LoadBanner(tt.path)
			if err != nil {
				t.Fatalf("expected no error for %s banner: %v", tt.name, err)
			}

			if len(banner) != 95 {
				t.Fatalf("mapped characters = %d, want 95", len(banner))
			}

			for _, ch := range []rune{' ', '~'} {
				if len(banner[ch]) != 8 {
					t.Fatalf("lines for %q = %d, want 8", ch, len(banner[ch]))
				}
			}

			for ch, lines := range banner {
				for _, line := range lines {
					if strings.Contains(line, "\r") {
						t.Fatalf("line for %q contains carriage return", ch)
					}
				}
			}
		})
	}
}

func TestLoadBannerHandlesCRLFLineEndings(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "banner.txt")

	var content strings.Builder
	for ascii := 32; ascii <= 126; ascii++ {
		content.WriteString("\r\n")
		for line := 0; line < 8; line++ {
			content.WriteString("line\r\n")
		}
	}

	err := os.WriteFile(filePath, []byte(content.String()), 0644)
	if err != nil {
		t.Fatal("writing temp file failed")
	}

	banner, err := LoadBanner(filePath)
	if err != nil {
		t.Fatal("expected no error for CRLF banner file")
	}

	if strings.Contains(banner['A'][0], "\r") {
		t.Fatal("expected banner lines without carriage returns")
	}
}

func TestLoadBannerRejectsMissingCharacterBlock(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "banner.txt")

	var content strings.Builder
	for ascii := 32; ascii < 126; ascii++ {
		content.WriteString("\n")
		for line := 0; line < 8; line++ {
			content.WriteString("line\n")
		}
	}

	err := os.WriteFile(filePath, []byte(content.String()), 0644)
	if err != nil {
		t.Fatal("writing temp file failed")
	}

	banner, err := LoadBanner(filePath)
	if err == nil {
		t.Fatal("expected error for banner with missing character block")
	}
	if banner != nil {
		t.Fatal("expected nil banner for banner with missing character block")
	}
}
