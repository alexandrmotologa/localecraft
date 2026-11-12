package parser

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// CSVParser handles tabular CSV translation exports and imports.
type CSVParser struct{}

// ParseBytes parses CSV bytes into a Catalog.
// Expected CSV header: Key, Target (or Source, Target, Context, Comment)
func (p *CSVParser) ParseBytes(data []byte, locale string) (*Catalog, error) {
	catalog := NewCatalog("csv", locale, "")
	r := csv.NewReader(bytes.NewReader(data))

	// Read header
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}

	keyIdx := -1
	targetIdx := -1
	sourceIdx := -1
	contextIdx := -1
	commentIdx := -1

	for i, col := range header {
		clean := strings.ToLower(strings.TrimSpace(col))
		switch clean {
		case "key", "id", "key_name":
			keyIdx = i
		case "target", "translation", "value", locale:
			targetIdx = i
		case "source", "source_value", "en":
			sourceIdx = i
		case "context":
			contextIdx = i
		case "comment", "note", "description":
			commentIdx = i
		}
	}

	if keyIdx == -1 {
		return nil, fmt.Errorf("csv header missing 'Key' column")
	}
	if targetIdx == -1 {
		if sourceIdx != -1 {
			targetIdx = sourceIdx
		} else {
			targetIdx = 1
		}
	}

	lineNum := 1
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read csv line %d: %w", lineNum, err)
		}
		lineNum++

		if len(record) <= keyIdx {
			continue
		}

		key := strings.TrimSpace(record[keyIdx])
		if key == "" {
			continue
		}

		val := ""
		if targetIdx < len(record) {
			val = record[targetIdx]
		}

		entry := &Entry{
			Key:        key,
			Value:      val,
			LineNumber: lineNum,
		}

		if contextIdx >= 0 && contextIdx < len(record) {
			entry.Context = record[contextIdx]
		}
		if commentIdx >= 0 && commentIdx < len(record) {
			entry.Comment = record[commentIdx]
		}

		catalog.Set(entry)
	}

	return catalog, nil
}

// ParseFile parses a CSV file from disk.
func (p *CSVParser) ParseFile(path string, locale string) (*Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read csv file %s: %w", path, err)
	}
	cat, err := p.ParseBytes(data, locale)
	if err != nil {
		return nil, err
	}
	cat.Path = path
	return cat, nil
}

// Serialize converts a catalog into standard CSV tabular format.
func (p *CSVParser) Serialize(catalog *Catalog) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Write header
	header := []string{"Key", "Target", "Context", "Comment"}
	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	for _, k := range catalog.KeyOrder {
		entry, ok := catalog.Entries[k]
		if !ok {
			continue
		}
		row := []string{entry.Key, entry.Value, entry.Context, entry.Comment}
		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("write csv row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}

	return buf.Bytes(), nil
}

// WriteFile writes the serialized CSV to disk.
func (p *CSVParser) WriteFile(catalog *Catalog, path string) error {
	data, err := p.Serialize(catalog)
	if err != nil {
		return err
	}
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
