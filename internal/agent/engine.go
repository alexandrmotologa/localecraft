package agent

import (
	"context"
)

// TranslateRequest encapsulates all data required for context-aware localization.
type TranslateRequest struct {
	Key                string            `json:"key"`
	SourceText         string            `json:"source_text"`
	SourceLocale       string            `json:"source_locale"`
	TargetLocale       string            `json:"target_locale"`
	Context            string            `json:"context"`
	Glossary           map[string]string `json:"glossary"`
	IsICU              bool              `json:"is_icu"`
	TargetCharLimit    int               `json:"target_char_limit,omitempty"`
	PreviousCandidates []string          `json:"previous_candidates,omitempty"`
}

// TranslateResult holds the translation output and metadata from an agent.
type TranslateResult struct {
	Translation string `json:"translation"`
	Explanation string `json:"explanation,omitempty"`
	Provider    string `json:"provider"`
}

// Provider defines the interface for language translation engines.
type Provider interface {
	Translate(ctx context.Context, req TranslateRequest) (TranslateResult, error)
	Name() string
}
