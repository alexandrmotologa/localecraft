package store_test

import (
	"testing"

	"github.com/alexandrmotologa/localecraft/internal/config"
	"github.com/alexandrmotologa/localecraft/internal/parser"
	"github.com/alexandrmotologa/localecraft/internal/store"
	"github.com/alexandrmotologa/localecraft/internal/typography"
)

func TestStoreSyncAndQuery(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}
	defer db.Close()

	cfg := config.DefaultConfig()
	cfg.TargetLocales = []string{"de", "fr"}
	cfg.UIConstraints.Keys["btn.save"] = config.KeyConstraint{
		MaxWidthPX: 100,
		FontFamily: "Inter",
		FontSizePX: 14,
		FontWeight: 500,
		Context:    "Primary save button",
	}

	srcCat := parser.NewCatalog("json_nested", "en", "locales/en.json")
	srcCat.Set(&parser.Entry{Key: "btn.save", Value: "Save"})

	deCat := parser.NewCatalog("json_nested", "de", "locales/de.json")
	deCat.Set(&parser.Entry{Key: "btn.save", Value: "Speichern und anwenden"}) // Will overflow

	typo, err := typography.NewEngine()
	if err != nil {
		t.Fatalf("failed to init typography engine: %v", err)
	}

	targetCats := map[string]*parser.Catalog{
		"de": deCat,
		"fr": parser.NewCatalog("json_nested", "fr", "locales/fr.json"),
	}

	if err := db.SyncCatalogs(cfg, srcCat, targetCats, typo); err != nil {
		t.Fatalf("sync catalogs failed: %v", err)
	}

	keys, err := db.ListKeys("", "")
	if err != nil {
		t.Fatalf("list keys failed: %v", err)
	}

	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}

	k := keys[0]
	if k.KeyName != "btn.save" {
		t.Errorf("expected key 'btn.save', got '%s'", k.KeyName)
	}

	deTrans, ok := k.Translations["de"]
	if !ok {
		t.Fatalf("expected German translation record")
	}
	if deTrans.Status != "OVERFLOW" {
		t.Errorf("expected status 'OVERFLOW' for German, got '%s'", deTrans.Status)
	}

	frTrans, ok := k.Translations["fr"]
	if !ok {
		t.Fatalf("expected French translation record")
	}
	if frTrans.Status != "MISSING" {
		t.Errorf("expected status 'MISSING' for French, got '%s'", frTrans.Status)
	}

	sum, err := db.GetSummary()
	if err != nil {
		t.Fatalf("get summary failed: %v", err)
	}
	if sum.TotalKeys != 1 || sum.TotalOverflows != 1 || sum.TotalMissing != 1 {
		t.Errorf("unexpected summary: %+v", sum)
	}
}
