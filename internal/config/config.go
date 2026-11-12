package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileConfig maps a source file, format, and output destination pattern.
type FileConfig struct {
	Path          string `yaml:"path" json:"path"`
	Format        string `yaml:"format" json:"format"`
	OutputPattern string `yaml:"output_pattern" json:"output_pattern"`
}

// FontDefaults specifies global typography fallback values.
type FontDefaults struct {
	Family          string  `yaml:"family" json:"family"`
	SizePX          float64 `yaml:"size_px" json:"size_px"`
	Weight          int     `yaml:"weight" json:"weight"`
	LetterSpacingPX float64 `yaml:"letter_spacing_px" json:"letter_spacing_px"`
	PaddingXPX      float64 `yaml:"padding_x_px" json:"padding_x_px"`
	LineHeight      float64 `yaml:"line_height" json:"line_height"`
}

// KeyConstraint defines physical container bounds and semantic context for a key.
type KeyConstraint struct {
	MaxWidthPX         float64 `yaml:"max_width_px" json:"max_width_px"`
	MaxLines           int     `yaml:"max_lines" json:"max_lines"`
	Context            string  `yaml:"context" json:"context"`
	Format             string  `yaml:"format" json:"format"`
	AllowedOverflowPCT float64 `yaml:"allowed_overflow_pct" json:"allowed_overflow_pct"`
	FontFamily         string  `yaml:"font_family" json:"font_family"`
	FontSizePX         float64 `yaml:"font_size_px" json:"font_size_px"`
	FontWeight         int     `yaml:"font_weight" json:"font_weight"`
}

// UIConstraints contains default font specifications and per-key layout limits.
type UIConstraints struct {
	FontDefaults FontDefaults             `yaml:"font_defaults" json:"font_defaults"`
	Keys         map[string]KeyConstraint `yaml:"keys" json:"keys"`
}

// GlossaryTerm defines standard corporate terminology translations.
type GlossaryTerm struct {
	Term         string            `yaml:"term" json:"term"`
	Translations map[string]string `yaml:"translations" json:"translations"`
}

// AIConfig defines LLM provider settings for context translation and rewording.
type AIConfig struct {
	Provider string `yaml:"provider" json:"provider"` // "gemini", "openai", "heuristic"
	Model    string `yaml:"model" json:"model"`
	APIKey   string `yaml:"api_key" json:"api_key"`
	BaseURL  string `yaml:"base_url" json:"base_url"`
}

// Config represents the complete localecraft configuration manifest.
type Config struct {
	Version       string        `yaml:"version" json:"version"`
	SourceLocale  string        `yaml:"source_locale" json:"source_locale"`
	TargetLocales []string      `yaml:"target_locales" json:"target_locales"`
	Files         []FileConfig  `yaml:"files" json:"files"`
	UIConstraints UIConstraints `yaml:"ui_constraints" json:"ui_constraints"`
	Glossary      []GlossaryTerm`yaml:"glossary" json:"glossary"`
	AI            AIConfig      `yaml:"ai" json:"ai"`

	// WorkingDir keeps track of the root folder where config was loaded.
	WorkingDir string `yaml:"-" json:"-"`
}

// DefaultConfig returns a configuration struct populated with baseline values.
func DefaultConfig() *Config {
	return &Config{
		Version:      "1",
		SourceLocale: "en",
		TargetLocales: []string{"de", "fr", "es"},
		UIConstraints: UIConstraints{
			FontDefaults: FontDefaults{
				Family:          "Inter",
				SizePX:          14,
				Weight:          500,
				LetterSpacingPX: 0.0,
				PaddingXPX:      16,
				LineHeight:      1.4,
			},
			Keys: make(map[string]KeyConstraint),
		},
		Glossary: make([]GlossaryTerm, 0),
		AI: AIConfig{
			Provider: "heuristic",
		},
	}
}

// Load reads and parses a YAML configuration file from disk.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse yaml config: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err == nil {
		cfg.WorkingDir = filepath.Dir(absPath)
	} else {
		cfg.WorkingDir = "."
	}

	// Environment variable overrides for AI credentials
	if cfg.AI.APIKey == "" {
		if envKey := os.Getenv("LOCALECRAFT_AI_KEY"); envKey != "" {
			cfg.AI.APIKey = envKey
		} else if envGemini := os.Getenv("GEMINI_API_KEY"); envGemini != "" && (cfg.AI.Provider == "" || cfg.AI.Provider == "gemini") {
			cfg.AI.APIKey = envGemini
			if cfg.AI.Provider == "" {
				cfg.AI.Provider = "gemini"
			}
		} else if envOpenAI := os.Getenv("OPENAI_API_KEY"); envOpenAI != "" && cfg.AI.Provider == "openai" {
			cfg.AI.APIKey = envOpenAI
		}
	}

	if cfg.AI.Provider == "" {
		cfg.AI.Provider = "heuristic"
	}

	return cfg, nil
}

// ResolveOutputPath calculates the destination path for a given locale.
func (f *FileConfig) ResolveOutputPath(locale string, baseDir string) string {
	pattern := f.OutputPattern
	if pattern == "" {
		ext := filepath.Ext(f.Path)
		pattern = strings.TrimSuffix(f.Path, ext) + "." + locale + ext
	}

	out := strings.ReplaceAll(pattern, "{{.Locale}}", locale)
	out = strings.ReplaceAll(out, "{locale}", locale)

	if !filepath.IsAbs(out) && baseDir != "" {
		return filepath.Join(baseDir, out)
	}
	return out
}

// GetKeyConstraint returns the constraint for a specific key, falling back to defaults.
func (c *Config) GetKeyConstraint(key string) KeyConstraint {
	kc, ok := c.UIConstraints.Keys[key]
	if !ok {
		kc = KeyConstraint{}
	}

	if kc.FontFamily == "" {
		kc.FontFamily = c.UIConstraints.FontDefaults.Family
	}
	if kc.FontSizePX <= 0 {
		kc.FontSizePX = c.UIConstraints.FontDefaults.SizePX
	}
	if kc.FontWeight <= 0 {
		kc.FontWeight = c.UIConstraints.FontDefaults.Weight
	}

	return kc
}
