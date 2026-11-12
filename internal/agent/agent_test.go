package agent_test

import (
	"context"
	"testing"

	"github.com/alexandrmotologa/localecraft/internal/agent"
	"github.com/alexandrmotologa/localecraft/internal/config"
	"github.com/alexandrmotologa/localecraft/internal/typography"
)

func TestHeuristicProvider(t *testing.T) {
	provider := agent.NewHeuristicProvider()
	ctx := context.Background()

	// 1. Unconstrained translation of "save" to German
	res1, err := provider.Translate(ctx, agent.TranslateRequest{
		Key:          "btn.save",
		SourceText:   "save",
		SourceLocale: "en",
		TargetLocale: "de",
	})
	if err != nil {
		t.Fatalf("translate failed: %v", err)
	}
	if res1.Translation != "Speichern und anwenden" {
		t.Fatalf("expected standard German 'Speichern und anwenden', got '%s'", res1.Translation)
	}

	// 2. Length-constrained translation (< 12 chars)
	res2, err := provider.Translate(ctx, agent.TranslateRequest{
		Key:             "btn.save",
		SourceText:      "save",
		SourceLocale:    "en",
		TargetLocale:    "de",
		TargetCharLimit: 10,
	})
	if err != nil {
		t.Fatalf("translate with limit failed: %v", err)
	}
	if res2.Translation != "Speichern" {
		t.Fatalf("expected shortened German 'Speichern', got '%s'", res2.Translation)
	}

	// 3. Ultra-compact constraint (< 8 chars)
	res3, err := provider.Translate(ctx, agent.TranslateRequest{
		Key:             "btn.save",
		SourceText:      "save",
		SourceLocale:    "en",
		TargetLocale:    "de",
		TargetCharLimit: 7,
	})
	if err != nil {
		t.Fatalf("translate compact failed: %v", err)
	}
	if res3.Translation != "Sichern" {
		t.Fatalf("expected compact 'Sichern', got '%s'", res3.Translation)
	}

	// 4. Glossary override
	resGlossary, err := provider.Translate(ctx, agent.TranslateRequest{
		Key:          "nav.workspace",
		SourceText:   "Workspace",
		SourceLocale: "en",
		TargetLocale: "de",
		Glossary: map[string]string{
			"Workspace": "Eigener Arbeitsbereich",
		},
	})
	if err != nil {
		t.Fatalf("glossary translation failed: %v", err)
	}
	if resGlossary.Translation != "Eigener Arbeitsbereich" {
		t.Fatalf("expected glossary translation 'Eigener Arbeitsbereich', got '%s'", resGlossary.Translation)
	}
}

func TestLayoutOptimizerLoop(t *testing.T) {
	provider := agent.NewHeuristicProvider()
	typoEngine, err := typography.NewEngine()
	if err != nil {
		t.Fatalf("failed to initialize typography engine: %v", err)
	}

	optimizer := agent.NewLayoutOptimizer(provider, typoEngine)
	ctx := context.Background()

	// Container of 90px with 12px padding each side -> 66px available
	// "Speichern und anwenden" (~145px) will overflow on iteration 1
	// Optimizer should autonomously compress to "Speichern" or "Sichern" which fits inside 66px!
	req := agent.OptimizeRequest{
		Key:          "btn.save",
		SourceText:   "save",
		SourceLocale: "en",
		TargetLocale: "de",
		Context:      "Primary action button in toolbar",
		Constraint: config.KeyConstraint{
			MaxWidthPX: 90,
			FontFamily: "Inter",
			FontSizePX: 14,
			FontWeight: 500,
		},
		MaxIterations: 4,
	}

	res, err := optimizer.Optimize(ctx, req)
	if err != nil {
		t.Fatalf("optimizer failed: %v", err)
	}

	if res.Status != "ACCEPTED" {
		t.Fatalf("expected status ACCEPTED, got %s (final translation: %s, width: %f, available: %f)",
			res.Status, res.FinalTranslation, res.FinalMeasure.RenderedWidthPX, res.FinalMeasure.AvailableWidthPX)
	}

	if res.FinalMeasure.IsOverflow {
		t.Fatalf("final measure still marked as overflow: %fpx > %fpx", res.FinalMeasure.RenderedWidthPX, res.FinalMeasure.AvailableWidthPX)
	}

	if res.IterationsUsed < 2 {
		t.Fatalf("expected at least 2 iterations for shortening, got %d", res.IterationsUsed)
	}
}
