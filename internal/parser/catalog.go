package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Entry represents a single localized string and its metadata.
type Entry struct {
	Key          string            `json:"key"`
	Value        string            `json:"value"`
	Comment      string            `json:"comment,omitempty"`
	Context      string            `json:"context,omitempty"`
	Placeholders map[string]string `json:"placeholders,omitempty"`
	PluralForms  []string          `json:"plural_forms,omitempty"`
	RawMeta      map[string]any    `json:"raw_meta,omitempty"`
	LineNumber   int               `json:"line_number,omitempty"`
}

// Catalog represents a collection of localized entries for a specific locale.
type Catalog struct {
	Format   string            `json:"format"`
	Locale   string            `json:"locale"`
	Path     string            `json:"path"`
	Entries  map[string]*Entry `json:"entries"`
	KeyOrder []string          `json:"key_order"`
}

// NewCatalog creates an empty Catalog initialized for a specific format and locale.
func NewCatalog(format string, locale string, path string) *Catalog {
	return &Catalog{
		Format:   format,
		Locale:   locale,
		Path:     path,
		Entries:  make(map[string]*Entry),
		KeyOrder: make([]string, 0),
	}
}

// Set adds or updates an entry, maintaining insertion order.
func (c *Catalog) Set(entry *Entry) {
	if _, exists := c.Entries[entry.Key]; !exists {
		c.KeyOrder = append(c.KeyOrder, entry.Key)
	}
	c.Entries[entry.Key] = entry
}

// Get retrieves an entry by key.
func (c *Catalog) Get(key string) (*Entry, bool) {
	e, ok := c.Entries[key]
	return e, ok
}

// Keys returns the ordered list of keys.
func (c *Catalog) Keys() []string {
	return c.KeyOrder
}

// Parser defines the interface for reading and writing localization files.
type Parser interface {
	ParseBytes(data []byte, locale string) (*Catalog, error)
	ParseFile(path string, locale string) (*Catalog, error)
	Serialize(catalog *Catalog) ([]byte, error)
	WriteFile(catalog *Catalog, path string) error
}

// DetectFormat inspects file path or extension to determine format.
func DetectFormat(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".po", ".pot":
		return "gettext_po"
	case ".arb":
		return "flutter_arb"
	case ".csv":
		return "csv"
	case ".xlf", ".xliff":
		return "xliff"
	case ".json":
		return "json_nested"
	default:
		return "json_nested"
	}
}

// GetParser returns the appropriate Parser implementation for a format string.
func GetParser(format string) (Parser, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json_nested", "json":
		return &JSONNestedParser{}, nil
	case "gettext_po", "po", "gettext":
		return &GettextPOParser{}, nil
	case "flutter_arb", "arb":
		return &FlutterARBParser{}, nil
	case "csv":
		return &CSVParser{}, nil
	case "xliff", "xlf":
		return &XLIFFParser{}, nil
	default:
		return nil, fmt.Errorf("unsupported localization format: %s", format)
	}
}

// LoadFile reads and parses a localization file using its detected or specified format.
func LoadFile(path string, format string, locale string) (*Catalog, error) {
	if format == "" {
		format = DetectFormat(path)
	}
	p, err := GetParser(format)
	if err != nil {
		return nil, err
	}
	return p.ParseFile(path, locale)
}

// SaveFile serializes and writes a catalog to disk, creating parent directories if needed.
func SaveFile(catalog *Catalog, path string) error {
	p, err := GetParser(catalog.Format)
	if err != nil {
		return err
	}
	return p.WriteFile(catalog, path)
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// SortedKeys returns the catalog keys sorted alphabetically.
func (c *Catalog) SortedKeys() []string {
	keys := make([]string, len(c.KeyOrder))
	copy(keys, c.KeyOrder)
	sort.Strings(keys)
	return keys
}
