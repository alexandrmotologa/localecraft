package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/alexandrmotologa/localecraft/internal/config"
	"github.com/alexandrmotologa/localecraft/internal/icu"
	"github.com/alexandrmotologa/localecraft/internal/typography"
)

// OptimizeRequest defines inputs for layout-constrained translation.
type OptimizeRequest struct {
	Key           string
	SourceText    string
	SourceLocale  string
	TargetLocale  string
	Context       string
	Constraint    config.KeyConstraint
	Glossary      map[string]string
	MaxIterations int
}

// AttemptRecord tracks each candidate tested during optimization.
type AttemptRecord struct {
	Iteration        int                      `json:"iteration"`
	Translation      string                   `json:"translation"`
	Measure          typography.MeasureResult `json:"measure"`
	ICUValid         bool                     `json:"icu_valid"`
	ICUErrors        []string                 `json:"icu_errors,omitempty"`
	TargetCharLimit  int                      `json:"target_char_limit"`
	CompressionRatio float64                  `json:"compression_ratio"`
}

// OptimizeResult represents the final outcome of the optimization loop.
type OptimizeResult struct {
	Key              string                   `json:"key"`
	SourceText       string                   `json:"source_text"`
	FinalTranslation string                   `json:"final_translation"`
	Status           string                   `json:"status"` // "ACCEPTED", "BEST_EFFORT_OVERFLOW", "ICU_SYNTAX_ERROR"
	IterationsUsed   int                      `json:"iterations_used"`
	FinalMeasure     typography.MeasureResult `json:"final_measure"`
	Attempts         []AttemptRecord          `json:"attempts"`
	Provider         string                   `json:"provider"`
}

// LayoutOptimizer orchestrates translation and font measurement in a bounded loop.
type LayoutOptimizer struct {
	provider   Provider
	typography *typography.Engine
}

// NewLayoutOptimizer constructs a LayoutOptimizer.
func NewLayoutOptimizer(provider Provider, typoEngine *typography.Engine) *LayoutOptimizer {
	return &LayoutOptimizer{
		provider:   provider,
		typography: typoEngine,
	}
}

// Optimize executes the autonomous length-constrained rewording loop.
func (opt *LayoutOptimizer) Optimize(ctx context.Context, req OptimizeRequest) (*OptimizeResult, error) {
	if req.MaxIterations <= 0 {
		req.MaxIterations = 4
	}

	result := &OptimizeResult{
		Key:        req.Key,
		SourceText: req.SourceText,
		Attempts:   make([]AttemptRecord, 0),
		Provider:   opt.provider.Name(),
	}

	isICU := req.Constraint.Format == "icu_messageformat" || strings.Contains(req.SourceText, "{")

	var previousCandidates []string
	currentCharLimit := 0
	var bestAttempt *AttemptRecord

	for iter := 1; iter <= req.MaxIterations; iter++ {
		tReq := TranslateRequest{
			Key:                req.Key,
			SourceText:         req.SourceText,
			SourceLocale:       req.SourceLocale,
			TargetLocale:       req.TargetLocale,
			Context:            req.Context,
			Glossary:           req.Glossary,
			IsICU:              isICU,
			TargetCharLimit:    currentCharLimit,
			PreviousCandidates: previousCandidates,
		}

		transRes, err := opt.provider.Translate(ctx, tReq)
		if err != nil {
			return nil, fmt.Errorf("translate iteration %d: %w", iter, err)
		}

		candidate := transRes.Translation
		previousCandidates = append(previousCandidates, candidate)

		// 1. ICU Validation
		icuValid := true
		var icuErrors []string
		if isICU {
			valRes := icu.Validate(req.SourceText, candidate, req.TargetLocale)
			icuValid = valRes.Valid
			icuErrors = valRes.Errors
		}

		// 2. Typography Measurement
		measureReq := typography.MeasureRequest{
			Text:            candidate,
			Family:          req.Constraint.FontFamily,
			SizePX:          req.Constraint.FontSizePX,
			Weight:          req.Constraint.FontWeight,
			LetterSpacingPX: 0.0,
			PaddingXPX:      12, // Default horizontal padding for buttons/cards
			MaxWidthPX:      req.Constraint.MaxWidthPX,
			MaxLines:        req.Constraint.MaxLines,
		}
		measureRes := opt.typography.Measure(measureReq)

		attempt := AttemptRecord{
			Iteration:        iter,
			Translation:      candidate,
			Measure:          measureRes,
			ICUValid:         icuValid,
			ICUErrors:        icuErrors,
			TargetCharLimit:  currentCharLimit,
			CompressionRatio: measureRes.CompressionRatio,
		}
		result.Attempts = append(result.Attempts, attempt)

		if bestAttempt == nil || measureRes.RenderedWidthPX < bestAttempt.Measure.RenderedWidthPX {
			bestAttempt = &result.Attempts[len(result.Attempts)-1]
		}

		// Check if satisfied
		if !measureRes.IsOverflow && icuValid {
			result.FinalTranslation = candidate
			result.Status = "ACCEPTED"
			result.IterationsUsed = iter
			result.FinalMeasure = measureRes
			return result, nil
		}

		// If it overflows, update target character limit for next iteration
		if measureRes.IsOverflow && measureRes.TargetCharLimit > 0 {
			currentCharLimit = measureRes.TargetCharLimit
		} else {
			currentCharLimit = measureRes.GlyphCount - 2
			if currentCharLimit < 3 {
				currentCharLimit = 3
			}
		}
	}

	// If loop finished without satisfying full constraints, pick best attempt
	if bestAttempt != nil {
		result.FinalTranslation = bestAttempt.Translation
		result.IterationsUsed = len(result.Attempts)
		result.FinalMeasure = bestAttempt.Measure
		if !bestAttempt.ICUValid {
			result.Status = "ICU_SYNTAX_ERROR"
		} else {
			result.Status = "BEST_EFFORT_OVERFLOW"
		}
	}

	return result, nil
}
