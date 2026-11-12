package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// FlutterARBParser parses and serializes Flutter Application Resource Bundle (.arb) files.
type FlutterARBParser struct{}

// ParseBytes parses an ARB JSON byte array.
func (p *FlutterARBParser) ParseBytes(data []byte, locale string) (*Catalog, error) {
	var raw map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode arb json: %w", err)
	}

	detectedLocale := locale
	if l, ok := raw["@@locale"].(string); ok && l != "" {
		detectedLocale = l
	}

	catalog := NewCatalog("flutter_arb", detectedLocale, "")

	// Extract keys and metadata
	metaMap := make(map[string]map[string]any)
	regularKeys := make([]string, 0)

	for k, v := range raw {
		if strings.HasPrefix(k, "@@") {
			// Global metadata, keep in catalog raw meta
			continue
		}
		if strings.HasPrefix(k, "@") {
			targetKey := strings.TrimPrefix(k, "@")
			if m, ok := v.(map[string]any); ok {
				metaMap[targetKey] = m
			}
			continue
		}
		regularKeys = append(regularKeys, k)
	}

	sort.Strings(regularKeys)

	for _, k := range regularKeys {
		val := fmt.Sprintf("%v", raw[k])
		meta := metaMap[k]

		entry := &Entry{
			Key:   k,
			Value: val,
		}

		if meta != nil {
			entry.RawMeta = meta
			if desc, ok := meta["description"].(string); ok {
				entry.Comment = desc
			}
			if placeholders, ok := meta["placeholders"].(map[string]any); ok {
				entry.Placeholders = make(map[string]string)
				for pName := range placeholders {
					entry.Placeholders[pName] = fmt.Sprintf("%v", placeholders[pName])
				}
			}
		}

		catalog.Set(entry)
	}

	return catalog, nil
}

// ParseFile parses an ARB file from disk.
func (p *FlutterARBParser) ParseFile(path string, locale string) (*Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read arb file %s: %w", path, err)
	}
	cat, err := p.ParseBytes(data, locale)
	if err != nil {
		return nil, err
	}
	cat.Path = path
	return cat, nil
}

// Serialize converts a catalog back into Flutter ARB format.
func (p *FlutterARBParser) Serialize(catalog *Catalog) ([]byte, error) {
	arbMap := make(map[string]any)

	arbMap["@@locale"] = catalog.Locale

	for _, key := range catalog.KeyOrder {
		entry, ok := catalog.Entries[key]
		if !ok {
			continue
		}

		arbMap[entry.Key] = entry.Value

		if entry.Comment != "" || entry.RawMeta != nil || len(entry.Placeholders) > 0 {
			meta := make(map[string]any)
			if entry.RawMeta != nil {
				for mk, mv := range entry.RawMeta {
					meta[mk] = mv
				}
			}
			if entry.Comment != "" {
				meta["description"] = entry.Comment
			}
			if len(entry.Placeholders) > 0 && meta["placeholders"] == nil {
				phMap := make(map[string]any)
				for pk, pv := range entry.Placeholders {
					phMap[pk] = map[string]any{"description": pv}
				}
				meta["placeholders"] = phMap
			}
			arbMap["@"+entry.Key] = meta
		}
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(arbMap); err != nil {
		return nil, fmt.Errorf("encode arb json: %w", err)
	}

	return buf.Bytes(), nil
}

// WriteFile writes the serialized ARB content to disk.
func (p *FlutterARBParser) WriteFile(catalog *Catalog, path string) error {
	data, err := p.Serialize(catalog)
	if err != nil {
		return err
	}
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
