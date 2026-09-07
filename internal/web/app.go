package web

import (
	"fmt"
	"html/template"
	"log/slog"
)

// App holds dependencies that are initialized once and read by every request.
type App struct {
	pageTemplate *template.Template
	banners      BannerRegistry
	logger       *slog.Logger
}

// NewApp validates and stores the application's immutable dependencies.
func NewApp(pageTemplate *template.Template, banners BannerRegistry, logger *slog.Logger) (*App, error) {
	if pageTemplate == nil {
		return nil, fmt.Errorf("page template is required")
	}
	templateNames := [...]string{
		"page",
		"input_controls",
		"generate_action",
		"output_box",
		"align_controls",
		"form_error",
		"clear_action",
		"download_action",
		"color_controls",
		"width_field",
		"theme_control",
		"not_found",
	}
	for _, name := range templateNames {
		if pageTemplate.Lookup(name) == nil {
			return nil, fmt.Errorf("missing template definition %q", name)
		}
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	registry := make(BannerRegistry, len(banners))
	for name, glyphs := range banners {
		registry[name] = glyphs
	}
	bannerNames := [...]string{BannerStandard, BannerShadow, BannerThinkertoy}
	for _, name := range bannerNames {
		if len(registry[name]) == 0 {
			return nil, fmt.Errorf("banner %q is required", name)
		}
	}

	return &App{
		pageTemplate: pageTemplate,
		banners:      registry,
		logger:       logger,
	}, nil
}
