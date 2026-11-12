package typography_test

import (
	"testing"

	"github.com/alexandrmotologa/localecraft/internal/typography"
)

func TestTypographyEngine(t *testing.T) {
	engine, err := typography.NewEngine()
	if err != nil {
		t.Fatalf("failed to create typography engine: %v", err)
	}

	// Test 1: English short string fits inside 100px container
	resEn := engine.Measure(typography.MeasureRequest{
		Text:            "Save",
		Family:          "Inter",
		SizePX:          14,
		Weight:          500,
		LetterSpacingPX: 0,
		PaddingXPX:      12,
		MaxWidthPX:      100, // available = 100 - 24 = 76px
		MaxLines:        1,
	})

	if resEn.IsOverflow {
		t.Errorf("expected 'Save' to fit without overflow, got width=%f available=%f", resEn.RenderedWidthPX, resEn.AvailableWidthPX)
	}
	if resEn.RenderedWidthPX <= 0 {
		t.Errorf("expected positive rendered width, got %f", resEn.RenderedWidthPX)
	}

	// Test 2: German expanded string overflows 100px container
	resDe := engine.Measure(typography.MeasureRequest{
		Text:            "Speichern und anwenden",
		Family:          "Inter",
		SizePX:          14,
		Weight:          500,
		LetterSpacingPX: 0,
		PaddingXPX:      12,
		MaxWidthPX:      100, // available = 76px
		MaxLines:        1,
	})

	if !resDe.IsOverflow {
		t.Errorf("expected 'Speichern und anwenden' to overflow 76px available width, got width=%f", resDe.RenderedWidthPX)
	}
	if resDe.OverflowPX <= 0 {
		t.Errorf("expected positive overflow in px, got %f", resDe.OverflowPX)
	}
	if resDe.TargetCharLimit >= len("Speichern und anwenden") {
		t.Errorf("expected target char limit to be constrained lower than %d, got %d", len("Speichern und anwenden"), resDe.TargetCharLimit)
	}
	if resDe.CompressionRatio >= 1.0 {
		t.Errorf("expected compression ratio < 1.0, got %f", resDe.CompressionRatio)
	}

	// Test 3: Weight scaling (Bold vs Regular)
	resRegular := engine.Measure(typography.MeasureRequest{
		Text:       "Account Settings",
		SizePX:     14,
		Weight:     400,
		MaxWidthPX: 200,
	})
	resBold := engine.Measure(typography.MeasureRequest{
		Text:       "Account Settings",
		SizePX:     14,
		Weight:     700,
		MaxWidthPX: 200,
	})

	if resBold.RenderedWidthPX <= resRegular.RenderedWidthPX {
		t.Errorf("expected bold width (%f) to exceed regular width (%f)", resBold.RenderedWidthPX, resRegular.RenderedWidthPX)
	}

	// Test 4: Letter spacing increases rendered width
	resSpaced := engine.Measure(typography.MeasureRequest{
		Text:            "Account Settings",
		SizePX:          14,
		Weight:          400,
		LetterSpacingPX: 2.0,
		MaxWidthPX:      200,
	})
	if resSpaced.RenderedWidthPX <= resRegular.RenderedWidthPX {
		t.Errorf("expected letter-spaced text to be wider, got %f vs %f", resSpaced.RenderedWidthPX, resRegular.RenderedWidthPX)
	}
}
