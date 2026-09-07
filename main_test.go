// Package main contains tests for web-server startup wiring.
package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	webapp "ascii-art/internal/web"
)

func TestNewHTTPServerConfiguration(t *testing.T) {
	handler := http.NewServeMux()
	server := newHTTPServer(handler)

	if server.Addr != ":8080" {
		t.Fatalf("Addr = %q, want :8080", server.Addr)
	}
	if server.Handler != handler {
		t.Fatal("server did not retain the supplied handler")
	}

	timeouts := map[string]struct {
		got  time.Duration
		want time.Duration
	}{
		"ReadHeaderTimeout": {server.ReadHeaderTimeout, 5 * time.Second},
		"ReadTimeout":       {server.ReadTimeout, 15 * time.Second},
		"WriteTimeout":      {server.WriteTimeout, 30 * time.Second},
		"IdleTimeout":       {server.IdleTimeout, 60 * time.Second},
	}
	for name, timeout := range timeouts {
		if timeout.got != timeout.want {
			t.Errorf("%s = %s, want %s", name, timeout.got, timeout.want)
		}
	}
	if shutdownTimeout != 5*time.Second {
		t.Errorf("shutdownTimeout = %s, want 5s", shutdownTimeout)
	}
}

func TestLoadBundledBanners(t *testing.T) {
	registry, err := loadBundledBanners()
	if err != nil {
		t.Fatalf("loadBundledBanners() error = %v", err)
	}

	for _, name := range []string{
		webapp.BannerStandard,
		webapp.BannerShadow,
		webapp.BannerThinkertoy,
	} {
		glyphs, ok := registry[name]
		if !ok {
			t.Errorf("registry is missing %q", name)
			continue
		}
		if len(glyphs) != 95 {
			t.Errorf("registry[%q] has %d glyphs, want 95", name, len(glyphs))
		}
	}
}

func TestRunFailsBeforeListeningWhenStartupResourcesAreMissing(t *testing.T) {
	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	err = run(logger)
	if err == nil {
		t.Fatal("run() error = nil, want missing startup resource error")
	}
	if !strings.Contains(err.Error(), "parse page template") {
		t.Errorf("run() error = %q, want template initialization context", err)
	}
}
