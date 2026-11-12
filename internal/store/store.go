package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexandrmotologa/localecraft/internal/config"
	"github.com/alexandrmotologa/localecraft/internal/icu"
	"github.com/alexandrmotologa/localecraft/internal/parser"
	"github.com/alexandrmotologa/localecraft/internal/typography"
	_ "modernc.org/sqlite"
)

// KeyRecord represents a localization key, its constraints, and all translations.
type KeyRecord struct {
	ID           int64                        `json:"id"`
	KeyName      string                       `json:"key_name"`
	SourceLocale string                       `json:"source_locale"`
	SourceValue  string                       `json:"source_value"`
	Context      string                       `json:"context"`
	MaxWidthPX   float64                      `json:"max_width_px"`
	FontFamily   string                       `json:"font_family"`
	FontSizePX   float64                      `json:"font_size_px"`
	FontWeight   int                          `json:"font_weight"`
	IsICU        bool                         `json:"is_icu"`
	Translations map[string]TranslationRecord `json:"translations"`
}

// TranslationRecord stores the locale translation and its layout measurement.
type TranslationRecord struct {
	ID              int64   `json:"id"`
	Locale          string  `json:"locale"`
	Value           string  `json:"value"`
	Status          string  `json:"status"` // "ACCEPTED", "OVERFLOW", "MISSING", "ICU_ERROR"
	RenderedWidthPX float64 `json:"rendered_width_px"`
	IsOverflow      bool    `json:"is_overflow"`
	OverflowPX      float64 `json:"overflow_px"`
	OverflowPCT     float64 `json:"overflow_pct"`
	IterationsUsed  int     `json:"iterations_used"`
	UpdatedAt       string  `json:"updated_at"`
}

// StoreSummary provides aggregate audit statistics for the UI dashboard.
type StoreSummary struct {
	TotalKeys       int `json:"total_keys"`
	TotalLocales    int `json:"total_locales"`
	TotalApproved   int `json:"total_approved"`
	TotalOverflows  int `json:"total_overflows"`
	TotalMissing    int `json:"total_missing"`
	TotalICUErrors  int `json:"total_icu_errors"`
}

// DB wraps SQLite operations for localecraft.
type DB struct {
	db *sql.DB
}

// Open opens or creates an SQLite database at the specified path.
func Open(path string) (*DB, error) {
	if path != ":memory:" {
		dir := filepath.Dir(path)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("create db directory: %w", err)
			}
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	s := &DB{db: db}
	if err := s.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return s, nil
}

// Close closes the underlying database connection.
func (s *DB) Close() error {
	return s.db.Close()
}

func (s *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key_name TEXT NOT NULL UNIQUE,
		source_locale TEXT NOT NULL,
		source_value TEXT NOT NULL,
		context TEXT DEFAULT '',
		max_width_px REAL DEFAULT 0,
		font_family TEXT DEFAULT 'Inter',
		font_size_px REAL DEFAULT 14,
		font_weight INTEGER DEFAULT 500,
		is_icu BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS translations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key_id INTEGER NOT NULL,
		locale TEXT NOT NULL,
		value TEXT NOT NULL,
		status TEXT NOT NULL,
		rendered_width_px REAL DEFAULT 0,
		is_overflow BOOLEAN DEFAULT FALSE,
		overflow_px REAL DEFAULT 0,
		overflow_pct REAL DEFAULT 0,
		iterations_used INTEGER DEFAULT 1,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(key_id) REFERENCES keys(id) ON DELETE CASCADE,
		UNIQUE(key_id, locale)
	);

	CREATE INDEX IF NOT EXISTS idx_trans_status ON translations(status);
	`
	_, err := s.db.Exec(schema)
	return err
}

// SyncCatalogs updates the SQLite state using the in-memory catalogs and config constraints.
func (s *DB) SyncCatalogs(cfg *config.Config, sourceCat *parser.Catalog, targetCats map[string]*parser.Catalog, typo *typography.Engine) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin sync tx: %w", err)
	}
	defer tx.Rollback()

	keyInsertStmt, err := tx.Prepare(`
		INSERT INTO keys (key_name, source_locale, source_value, context, max_width_px, font_family, font_size_px, font_weight, is_icu)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(key_name) DO UPDATE SET
			source_value = excluded.source_value,
			context = excluded.context,
			max_width_px = excluded.max_width_px,
			font_family = excluded.font_family,
			font_size_px = excluded.font_size_px,
			font_weight = excluded.font_weight,
			is_icu = excluded.is_icu
	`)
	if err != nil {
		return err
	}
	defer keyInsertStmt.Close()

	transInsertStmt, err := tx.Prepare(`
		INSERT INTO translations (key_id, locale, value, status, rendered_width_px, is_overflow, overflow_px, overflow_pct, iterations_used, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key_id, locale) DO UPDATE SET
			value = excluded.value,
			status = excluded.status,
			rendered_width_px = excluded.rendered_width_px,
			is_overflow = excluded.is_overflow,
			overflow_px = excluded.overflow_px,
			overflow_pct = excluded.overflow_pct,
			iterations_used = excluded.iterations_used,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer transInsertStmt.Close()

	for _, k := range sourceCat.KeyOrder {
		srcEntry, ok := sourceCat.Entries[k]
		if !ok {
			continue
		}

		constraint := cfg.GetKeyConstraint(k)
		isICU := constraint.Format == "icu_messageformat" || strings.Contains(srcEntry.Value, "{")

		contextVal := srcEntry.Comment
		if constraint.Context != "" {
			contextVal = constraint.Context
		}

		if _, err := keyInsertStmt.Exec(k, sourceCat.Locale, srcEntry.Value, contextVal, constraint.MaxWidthPX, constraint.FontFamily, constraint.FontSizePX, constraint.FontWeight, isICU); err != nil {
			return fmt.Errorf("insert key %s: %w", k, err)
		}

		var keyID int64
		if err := tx.QueryRow("SELECT id FROM keys WHERE key_name = ?", k).Scan(&keyID); err != nil {
			return fmt.Errorf("lookup key id for %s: %w", k, err)
		}

		// Insert or update target translations
		for _, targetLocale := range cfg.TargetLocales {
			var targetVal string
			var status string = "MISSING"
			var measure typography.MeasureResult

			if cat, ok := targetCats[targetLocale]; ok {
				if tgtEntry, ok := cat.Entries[k]; ok && tgtEntry.Value != "" {
					targetVal = tgtEntry.Value

					// ICU verification
					if isICU {
						valRes := icu.Validate(srcEntry.Value, targetVal, targetLocale)
						if !valRes.Valid {
							status = "ICU_ERROR"
						}
					}

					// Typography measurement
					measure = typo.Measure(typography.MeasureRequest{
						Text:            targetVal,
						Family:          constraint.FontFamily,
						SizePX:          constraint.FontSizePX,
						Weight:          constraint.FontWeight,
						LetterSpacingPX: 0,
						PaddingXPX:      12,
						MaxWidthPX:      constraint.MaxWidthPX,
						MaxLines:        constraint.MaxLines,
					})

					if status != "ICU_ERROR" {
						if measure.IsOverflow {
							status = "OVERFLOW"
						} else {
							status = "ACCEPTED"
						}
					}
				}
			}

			if _, err := transInsertStmt.Exec(keyID, targetLocale, targetVal, status, measure.RenderedWidthPX, measure.IsOverflow, measure.OverflowPX, measure.OverflowPCT, 1); err != nil {
				return fmt.Errorf("insert translation %s/%s: %w", k, targetLocale, err)
			}
		}
	}

	return tx.Commit()
}

// ListKeys returns all keys matching the query filter.
func (s *DB) ListKeys(search string, statusFilter string) ([]KeyRecord, error) {
	query := `
		SELECT id, key_name, source_locale, source_value, context, max_width_px, font_family, font_size_px, font_weight, is_icu
		FROM keys
		ORDER BY key_name ASC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query keys: %w", err)
	}
	defer rows.Close()

	var records []KeyRecord
	keyMap := make(map[int64]*KeyRecord)

	for rows.Next() {
		var rec KeyRecord
		rec.Translations = make(map[string]TranslationRecord)
		if err := rows.Scan(&rec.ID, &rec.KeyName, &rec.SourceLocale, &rec.SourceValue, &rec.Context, &rec.MaxWidthPX, &rec.FontFamily, &rec.FontSizePX, &rec.FontWeight, &rec.IsICU); err != nil {
			return nil, err
		}
		records = append(records, rec)
		keyMap[rec.ID] = &records[len(records)-1]
	}

	// Fetch translations
	transRows, err := s.db.Query(`SELECT id, key_id, locale, value, status, rendered_width_px, is_overflow, overflow_px, overflow_pct, iterations_used, updated_at FROM translations`)
	if err != nil {
		return nil, fmt.Errorf("query translations: %w", err)
	}
	defer transRows.Close()

	for transRows.Next() {
		var keyID int64
		var tr TranslationRecord
		var updatedAt time.Time
		if err := transRows.Scan(&tr.ID, &keyID, &tr.Locale, &tr.Value, &tr.Status, &tr.RenderedWidthPX, &tr.IsOverflow, &tr.OverflowPX, &tr.OverflowPCT, &tr.IterationsUsed, &updatedAt); err != nil {
			return nil, err
		}
		tr.UpdatedAt = updatedAt.Format(time.RFC3339)
		if parent, ok := keyMap[keyID]; ok {
			parent.Translations[tr.Locale] = tr
		}
	}

	// Apply filtering
	if search == "" && statusFilter == "" {
		return records, nil
	}

	searchLower := strings.ToLower(search)
	statusLower := strings.ToLower(statusFilter)

	var filtered []KeyRecord
	for _, r := range records {
		matchSearch := searchLower == "" ||
			strings.Contains(strings.ToLower(r.KeyName), searchLower) ||
			strings.Contains(strings.ToLower(r.SourceValue), searchLower)

		if !matchSearch {
			continue
		}

		if statusLower == "" {
			filtered = append(filtered, r)
			continue
		}

		matchStatus := false
		for _, t := range r.Translations {
			if strings.EqualFold(t.Status, statusLower) {
				matchStatus = true
				break
			}
		}

		if matchStatus {
			filtered = append(filtered, r)
		}
	}

	return filtered, nil
}

// SaveTranslation records an updated translation directly into the store.
func (s *DB) SaveTranslation(keyName string, locale string, value string, status string, measure typography.MeasureResult, iterations int) error {
	var keyID int64
	err := s.db.QueryRow("SELECT id FROM keys WHERE key_name = ?", keyName).Scan(&keyID)
	if err != nil {
		return fmt.Errorf("key '%s' not found: %w", keyName, err)
	}

	query := `
		INSERT INTO translations (key_id, locale, value, status, rendered_width_px, is_overflow, overflow_px, overflow_pct, iterations_used, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key_id, locale) DO UPDATE SET
			value = excluded.value,
			status = excluded.status,
			rendered_width_px = excluded.rendered_width_px,
			is_overflow = excluded.is_overflow,
			overflow_px = excluded.overflow_px,
			overflow_pct = excluded.overflow_pct,
			iterations_used = excluded.iterations_used,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err = s.db.Exec(query, keyID, locale, value, status, measure.RenderedWidthPX, measure.IsOverflow, measure.OverflowPX, measure.OverflowPCT, iterations)
	return err
}

// GetSummary calculates aggregate metrics across all keys and locales.
func (s *DB) GetSummary() (StoreSummary, error) {
	var sum StoreSummary

	err := s.db.QueryRow("SELECT count(*) FROM keys").Scan(&sum.TotalKeys)
	if err != nil {
		return sum, err
	}

	err = s.db.QueryRow("SELECT count(DISTINCT locale) FROM translations").Scan(&sum.TotalLocales)
	if err != nil {
		return sum, err
	}

	rows, err := s.db.Query("SELECT status, count(*) FROM translations GROUP BY status")
	if err != nil {
		return sum, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err == nil {
			switch strings.ToUpper(status) {
			case "ACCEPTED":
				sum.TotalApproved += count
			case "OVERFLOW":
				sum.TotalOverflows += count
			case "MISSING":
				sum.TotalMissing += count
			case "ICU_ERROR":
				sum.TotalICUErrors += count
			}
		}
	}

	return sum, nil
}
