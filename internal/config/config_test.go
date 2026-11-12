package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/localecraft/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.SourceLocale != "en" {
		t.Fatalf("expected source locale 'en', got '%s'", cfg.SourceLocale)
	}
	if len(cfg.TargetLocales) == 0 {
		t.Fatalf("expected default target locales, got none")
	}
	if cfg.UIConstraints.FontDefaults.SizePX != 14 {
		t.Fatalf("expected default font size 14, got %f", cfg.UIConstraints.FontDefaults.SizePX)
	}
}

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "localecraft.yaml")

	yamlContent := `
version: "1"
source_locale: "en"
target_locales:
  - "de"
  - "fr"
files:
  - path: "locales/en.json"
    format: "json_nested"
    output_pattern: "locales/{{.Locale}}.json"
ui_constraints:
  font_defaults:
    family: "Inter"
    size_px: 16
    weight: 600
    letter_spacing_px: 0.5
    padding_x_px: 20
  keys:
    "btn.submit":
      max_width_px: 120
      max_lines: 1
      context: "Form submit button"
glossary:
  - term: "Dashboard"
    translations:
      de: "Übersicht"
      fr: "Tableau de bord"
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.SourceLocale != "en" {
		t.Errorf("expected en, got %s", cfg.SourceLocale)
	}
	if len(cfg.TargetLocales) != 2 || cfg.TargetLocales[0] != "de" {
		t.Errorf("unexpected target locales: %v", cfg.TargetLocales)
	}
	if len(cfg.Files) != 1 || cfg.Files[0].Format != "json_nested" {
		t.Errorf("unexpected files configuration: %v", cfg.Files)
	}

	kc := cfg.GetKeyConstraint("btn.submit")
	if kc.MaxWidthPX != 120 {
		t.Errorf("expected max width 120, got %f", kc.MaxWidthPX)
	}
	if kc.FontFamily != "Inter" || kc.FontSizePX != 16 || kc.FontWeight != 600 {
		t.Errorf("expected inherited font parameters, got family=%s size=%f weight=%d", kc.FontFamily, kc.FontSizePX, kc.FontWeight)
	}

	outPath := cfg.Files[0].ResolveOutputPath("de", tempDir)
	expectedOut := filepath.Join(tempDir, "locales/de.json")
	if filepath.Clean(outPath) != filepath.Clean(expectedOut) {
		t.Errorf("expected output path %s, got %s", expectedOut, outPath)
	}
}
