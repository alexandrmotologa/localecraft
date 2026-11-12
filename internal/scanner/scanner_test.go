package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/localecraft/internal/config"
	"github.com/alexandrmotologa/localecraft/internal/scanner"
)

func TestScanDirectory(t *testing.T) {
	tempDir := t.TempDir()

	tsxContent := `
import React from 'react';
import { useTranslation } from 'react-i18next';

export const CheckoutHeader = () => {
  const { t } = useTranslation();
  return (
    <div className="flex items-center justify-between">
      <h1 className="text-lg font-bold">{t('nav.title')}</h1>
      <button className="w-32 px-3 py-1 text-sm font-semibold truncate">
        {t('nav.checkout_btn')}
      </button>
      <div className="max-w-[150px]">
        <span>{t('settings.danger_zone')}</span>
      </div>
    </div>
  );
};
`
	testFile := filepath.Join(tempDir, "CheckoutHeader.tsx")
	if err := os.WriteFile(testFile, []byte(tsxContent), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	constraints, err := scanner.ScanDirectory(tempDir)
	if err != nil {
		t.Fatalf("scan directory failed: %v", err)
	}

	if len(constraints) != 3 {
		t.Fatalf("expected 3 constraints, got %d", len(constraints))
	}

	constraintMap := make(map[string]scanner.InferredConstraint)
	for _, c := range constraints {
		constraintMap[c.Key] = c
	}

	// 1. Check nav.checkout_btn with w-32 -> 128px
	btnConstraint, ok := constraintMap["nav.checkout_btn"]
	if !ok {
		t.Fatalf("missing nav.checkout_btn constraint")
	}
	if btnConstraint.MaxWidthPX != 128 {
		t.Errorf("expected 128px for w-32, got %f", btnConstraint.MaxWidthPX)
	}
	if btnConstraint.FontSizePX != 14 {
		t.Errorf("expected 14px for text-sm, got %f", btnConstraint.FontSizePX)
	}
	if btnConstraint.FontWeight != 600 {
		t.Errorf("expected 600 for font-semibold, got %d", btnConstraint.FontWeight)
	}

	// 2. Check settings.danger_zone with max-w-[150px] -> 150px
	dangerConstraint, ok := constraintMap["settings.danger_zone"]
	if !ok {
		t.Fatalf("missing settings.danger_zone constraint")
	}
	if dangerConstraint.MaxWidthPX != 150 {
		t.Errorf("expected 150px for max-w-[150px], got %f", dangerConstraint.MaxWidthPX)
	}

	// 3. Test merging into Config
	cfg := config.DefaultConfig()
	added := scanner.UpdateConfigWithInferred(cfg, constraints)
	if added != 3 {
		t.Errorf("expected 3 added constraints, got %d", added)
	}
	if cfg.UIConstraints.Keys["nav.checkout_btn"].MaxWidthPX != 128 {
		t.Errorf("expected merged config to have 128px limit")
	}
}
