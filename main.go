// Package main initializes and runs the ASCII Art Web server.
package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ascii-art/internal/banner"
	webapp "ascii-art/internal/web"
)

const (
	serverAddress        = ":8080"
	pageTemplatePath     = "templates/index.html"
	notFoundTemplatePath = "templates/404.html"
	shutdownTimeout      = 5 * time.Second
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := run(logger); err != nil {
		logger.Error("application stopped", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	pageTemplate, err := template.ParseFiles(pageTemplatePath, notFoundTemplatePath)
	if err != nil {
		return fmt.Errorf("parse page template: %w", err)
	}

	banners, err := loadBundledBanners()
	if err != nil {
		return err
	}

	app, err := webapp.NewApp(pageTemplate, banners, logger)
	if err != nil {
		return fmt.Errorf("construct web application: %w", err)
	}

	server := newHTTPServer(app.Routes())
	return serve(logger, server)
}

func loadBundledBanners() (webapp.BannerRegistry, error) {
	specs := [...]struct {
		name string
		path string
	}{
		{name: webapp.BannerStandard, path: "banners/standard.txt"},
		{name: webapp.BannerShadow, path: "banners/shadow.txt"},
		{name: webapp.BannerThinkertoy, path: "banners/thinkertoy.txt"},
	}

	registry := make(webapp.BannerRegistry, len(specs))
	for _, spec := range specs {
		glyphs, err := banner.LoadBanner(spec.path)
		if err != nil {
			return nil, fmt.Errorf("load %s banner: %w", spec.name, err)
		}
		registry[spec.name] = glyphs
	}
	return registry, nil
}

func newHTTPServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              serverAddress,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func serve(logger *slog.Logger, server *http.Server) error {
	signalContext, stopSignals := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stopSignals()

	listenErrors := make(chan error, 1)
	logger.Info("server starting", slog.String("address", server.Addr))
	go func() {
		listenErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-listenErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("listen and serve: %w", err)
	case <-signalContext.Done():
		logger.Info("shutdown signal received")
	}

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	if err := <-listenErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server stopped during shutdown: %w", err)
	}

	logger.Info("graceful shutdown complete")
	return nil
}
