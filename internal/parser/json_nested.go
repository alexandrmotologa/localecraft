package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// JSONNestedParser handles nested and flat JSON localization catalogs.
type JSONNestedParser struct{}

// ParseBytes parses a JSON byte array into a flat dot-notated Catalog.
func (p *JSONNestedParser) ParseBytes(data []byte, locale string) (*Catalog, error) {
	var raw map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	catalog := NewCatalog("json_nested", locale, "")
	flattenJSON("", raw, catalog)
	return catalog, nil
}

func flattenJSON(prefix string, value any, catalog *Catalog) {
	switch v := value.(type) {
	case map[string]any:
		// Extract sorted keys for deterministic ordering if not already preserved
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			newKey := k
			if prefix != "" {
				newKey = prefix + "." + k
			}
			flattenJSON(newKey, v[k], catalog)
		}
	case string:
		catalog.Set(&Entry{
			Key:   prefix,
			Value: v,
		})
	case json.Number:
		catalog.Set(&Entry{
			Key:   prefix,
			Value: v.String(),
		})
	case bool:
		catalog.Set(&Entry{
			Key:   prefix,
			Value: fmt.Sprintf("%v", v),
		})
	default:
		if v != nil {
			catalog.Set(&Entry{
				Key:   prefix,
				Value: fmt.Sprintf("%v", v),
			})
		}
	}
}

// ParseFile reads a JSON file and parses it into a Catalog.
func (p *JSONNestedParser) ParseFile(path string, locale string) (*Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read json file %s: %w", path, err)
	}
	catalog, err := p.ParseBytes(data, locale)
	if err != nil {
		return nil, err
	}
	catalog.Path = path
	return catalog, nil
}

// Serialize reconstructs a nested JSON hierarchy and returns formatted JSON bytes.
func (p *JSONNestedParser) Serialize(catalog *Catalog) ([]byte, error) {
	nested := make(map[string]any)

	for _, key := range catalog.KeyOrder {
		entry, ok := catalog.Entries[key]
		if !ok {
			continue
		}
		parts := strings.Split(key, ".")
		insertNested(nested, parts, entry.Value)
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(nested); err != nil {
		return nil, fmt.Errorf("encode nested json: %w", err)
	}

	return buf.Bytes(), nil
}

func insertNested(current map[string]any, parts []string, value string) {
	if len(parts) == 0 {
		return
	}
	head := parts[0]
	if len(parts) == 1 {
		current[head] = value
		return
	}

	next, exists := current[head]
	var nextMap map[string]any
	if exists {
		var ok bool
		nextMap, ok = next.(map[string]any)
		if !ok {
			nextMap = make(map[string]any)
			current[head] = nextMap
		}
	} else {
		nextMap = make(map[string]any)
		current[head] = nextMap
	}

	insertNested(nextMap, parts[1:], value)
}

// WriteFile writes the serialized JSON to disk.
func (p *JSONNestedParser) WriteFile(catalog *Catalog, path string) error {
	data, err := p.Serialize(catalog)
	if err != nil {
		return err
	}
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
