package parser

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
)

// XLIFFParser handles XML Localization Interchange File Format (XLIFF 1.2).
type XLIFFParser struct{}

type xliffDoc struct {
	XMLName xml.Name  `xml:"xliff"`
	Version string    `xml:"version,attr"`
	File    xliffFile `xml:"file"`
}

type xliffFile struct {
	SourceLang string    `xml:"source-language,attr"`
	TargetLang string    `xml:"target-language,attr"`
	Datatype   string    `xml:"datatype,attr"`
	Original   string    `xml:"original,attr"`
	Body       xliffBody `xml:"body"`
}

type xliffBody struct {
	TransUnits []xliffTransUnit `xml:"trans-unit"`
}

type xliffTransUnit struct {
	ID     string `xml:"id,attr"`
	Source string `xml:"source"`
	Target string `xml:"target"`
	Note   string `xml:"note,omitempty"`
}

// ParseBytes parses an XLIFF 1.2 byte array.
func (p *XLIFFParser) ParseBytes(data []byte, locale string) (*Catalog, error) {
	var doc xliffDoc
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("unmarshal xliff xml: %w", err)
	}

	detectedLocale := locale
	if doc.File.TargetLang != "" {
		detectedLocale = doc.File.TargetLang
	}

	catalog := NewCatalog("xliff", detectedLocale, "")

	for _, u := range doc.File.Body.TransUnits {
		val := u.Target
		if val == "" {
			val = u.Source
		}

		entry := &Entry{
			Key:     u.ID,
			Value:   val,
			Comment: u.Note,
		}
		catalog.Set(entry)
	}

	return catalog, nil
}

// ParseFile reads an XLIFF file from disk.
func (p *XLIFFParser) ParseFile(path string, locale string) (*Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read xliff file %s: %w", path, err)
	}
	cat, err := p.ParseBytes(data, locale)
	if err != nil {
		return nil, err
	}
	cat.Path = path
	return cat, nil
}

// Serialize produces formatted XLIFF 1.2 XML.
func (p *XLIFFParser) Serialize(catalog *Catalog) ([]byte, error) {
	doc := xliffDoc{
		Version: "1.2",
		File: xliffFile{
			SourceLang: "en",
			TargetLang: catalog.Locale,
			Datatype:   "plaintext",
			Original:   catalog.Path,
			Body: xliffBody{
				TransUnits: make([]xliffTransUnit, 0, len(catalog.KeyOrder)),
			},
		},
	}

	for _, k := range catalog.KeyOrder {
		entry, ok := catalog.Entries[k]
		if !ok {
			continue
		}
		doc.File.Body.TransUnits = append(doc.File.Body.TransUnits, xliffTransUnit{
			ID:     entry.Key,
			Source: entry.Value,
			Target: entry.Value,
			Note:   entry.Comment,
		})
	}

	var buf bytes.Buffer
	buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(doc); err != nil {
		return nil, fmt.Errorf("encode xliff xml: %w", err)
	}
	buf.WriteString("\n")

	return buf.Bytes(), nil
}

// WriteFile writes the XLIFF XML to disk.
func (p *XLIFFParser) WriteFile(catalog *Catalog, path string) error {
	data, err := p.Serialize(catalog)
	if err != nil {
		return err
	}
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
