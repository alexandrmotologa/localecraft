package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexandrmotologa/localecraft/internal/agent"
	"github.com/alexandrmotologa/localecraft/internal/api"
	"github.com/alexandrmotologa/localecraft/internal/config"
	"github.com/alexandrmotologa/localecraft/internal/parser"
	"github.com/alexandrmotologa/localecraft/internal/store"
	"github.com/alexandrmotologa/localecraft/internal/typography"
)

func setupTestServer(t *testing.T) (*api.Server, *store.DB) {
	cfg := config.DefaultConfig()
	cfg.TargetLocales = []string{"de", "fr"}
	cfg.UIConstraints.Keys["btn.save"] = config.KeyConstraint{
		MaxWidthPX: 100,
		FontFamily: "Inter",
		FontSizePX: 14,
		FontWeight: 500,
	}

	srcCat := parser.NewCatalog("json_nested", "en", "locales/en.json")
	srcCat.Set(&parser.Entry{Key: "btn.save", Value: "save"})

	deCat := parser.NewCatalog("json_nested", "de", "locales/de.json")
	deCat.Set(&parser.Entry{Key: "btn.save", Value: "Speichern und anwenden"})

	targetCats := map[string]*parser.Catalog{
		"de": deCat,
		"fr": parser.NewCatalog("json_nested", "fr", "locales/fr.json"),
	}

	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	typo, err := typography.NewEngine()
	if err != nil {
		t.Fatalf("init typo: %v", err)
	}

	if err := db.SyncCatalogs(cfg, srcCat, targetCats, typo); err != nil {
		t.Fatalf("sync db: %v", err)
	}

	provider := agent.NewHeuristicProvider()
	optimizer := agent.NewLayoutOptimizer(provider, typo)

	srv := api.NewServer(cfg, db, typo, optimizer, srcCat, targetCats, nil)
	return srv, db
}

func TestAPIEndpoints(t *testing.T) {
	srv, db := setupTestServer(t)
	defer db.Close()

	handler := srv.Router()

	// 1. GET /api/summary
	reqSummary := httptest.NewRequest(http.MethodGet, "/api/summary", nil)
	recSummary := httptest.NewRecorder()
	handler.ServeHTTP(recSummary, reqSummary)

	if recSummary.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recSummary.Code)
	}

	var sum store.StoreSummary
	if err := json.Unmarshal(recSummary.Body.Bytes(), &sum); err != nil {
		t.Fatalf("unmarshal summary: %v", err)
	}
	if sum.TotalKeys != 1 {
		t.Errorf("expected 1 key, got %d", sum.TotalKeys)
	}

	// 2. GET /api/keys
	reqKeys := httptest.NewRequest(http.MethodGet, "/api/keys", nil)
	recKeys := httptest.NewRecorder()
	handler.ServeHTTP(recKeys, reqKeys)

	if recKeys.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recKeys.Code)
	}

	var keys []store.KeyRecord
	if err := json.Unmarshal(recKeys.Body.Bytes(), &keys); err != nil {
		t.Fatalf("unmarshal keys: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected 1 key record, got %d", len(keys))
	}

	// 3. POST /api/measure
	measurePayload := []byte(`{"text":"Speichern","family":"Inter","size_px":14,"weight":500,"max_width_px":100,"padding_x_px":12}`)
	reqMeasure := httptest.NewRequest(http.MethodPost, "/api/measure", bytes.NewReader(measurePayload))
	recMeasure := httptest.NewRecorder()
	handler.ServeHTTP(recMeasure, reqMeasure)

	if recMeasure.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recMeasure.Code)
	}

	var mRes typography.MeasureResult
	if err := json.Unmarshal(recMeasure.Body.Bytes(), &mRes); err != nil {
		t.Fatalf("unmarshal measure: %v", err)
	}
	if mRes.IsOverflow {
		t.Errorf("expected 'Speichern' to fit in 76px available width, got overflow")
	}

	// 4. POST /api/translate with auto-reword
	transPayload := []byte(`{"key":"btn.save","target_locale":"de"}`)
	reqTrans := httptest.NewRequest(http.MethodPost, "/api/translate", bytes.NewReader(transPayload))
	recTrans := httptest.NewRecorder()
	handler.ServeHTTP(recTrans, reqTrans)

	if recTrans.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", recTrans.Code, recTrans.Body.String())
	}

	var optRes agent.OptimizeResult
	if err := json.Unmarshal(recTrans.Body.Bytes(), &optRes); err != nil {
		t.Fatalf("unmarshal opt res: %v", err)
	}
	if optRes.Status != "ACCEPTED" {
		t.Errorf("expected ACCEPTED, got %s", optRes.Status)
	}
}
