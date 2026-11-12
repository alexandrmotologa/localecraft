package parser_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexandrmotologa/localecraft/internal/parser"
)

func TestJSONNestedParser(t *testing.T) {
	rawJSON := `{
  "nav": {
    "title": "LocaleCraft Studio",
    "btn": {
      "submit": "Submit Changes",
      "cancel": "Cancel"
    }
  },
  "footer": {
    "copyright": "All rights reserved."
  }
}`

	p, err := parser.GetParser("json_nested")
	if err != nil {
		t.Fatalf("failed to get json parser: %v", err)
	}

	cat, err := p.ParseBytes([]byte(rawJSON), "en")
	if err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	if len(cat.Entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(cat.Entries))
	}

	e, ok := cat.Get("nav.btn.submit")
	if !ok || e.Value != "Submit Changes" {
		t.Fatalf("unexpected entry for nav.btn.submit: %v", e)
	}

	// Test serialization round-trip
	serialized, err := p.Serialize(cat)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	if !strings.Contains(string(serialized), `"submit": "Submit Changes"`) {
		t.Fatalf("serialized json missing submit button: %s", string(serialized))
	}
}

func TestGettextPOParser(t *testing.T) {
	rawPO := `
# Translator comment
#. Extracted comment
msgctxt "navigation"
msgid "Back"
msgstr "Zurück"

#. Plural counter
msgid "One task remaining"
msgid_plural "%d tasks remaining"
msgstr[0] "Eine Aufgabe übrig"
msgstr[1] "%d Aufgaben übrig"
`

	p, err := parser.GetParser("gettext_po")
	if err != nil {
		t.Fatalf("failed to get gettext parser: %v", err)
	}

	cat, err := p.ParseBytes([]byte(rawPO), "de")
	if err != nil {
		t.Fatalf("failed to parse PO: %v", err)
	}

	if len(cat.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(cat.Entries))
	}

	// Check context key
	e1, ok := cat.Get("navigation\x04Back")
	if !ok || e1.Value != "Zurück" {
		t.Fatalf("unexpected value for Back: %v", e1)
	}
	if e1.Context != "navigation" {
		t.Errorf("expected context 'navigation', got '%s'", e1.Context)
	}

	// Check plural entry
	e2, ok := cat.Get("One task remaining")
	if !ok || len(e2.PluralForms) != 2 {
		t.Fatalf("unexpected plural entry: %v", e2)
	}
	if e2.PluralForms[0] != "Eine Aufgabe übrig" || e2.PluralForms[1] != "%d Aufgaben übrig" {
		t.Errorf("unexpected plural forms: %v", e2.PluralForms)
	}

	// Serialize
	outBytes, err := p.Serialize(cat)
	if err != nil {
		t.Fatalf("failed to serialize PO: %v", err)
	}
	if !strings.Contains(string(outBytes), `msgstr[1] "%d Aufgaben übrig"`) {
		t.Errorf("serialized output missing plural line: %s", string(outBytes))
	}
}

func TestFlutterARBParser(t *testing.T) {
	rawARB := `{
  "@@locale": "en",
  "welcomeMessage": "Welcome to our application!",
  "@welcomeMessage": {
    "description": "Greeting on home view"
  },
  "cartItems": "{count, plural, =0{empty} one{1 item} other{{count} items}}",
  "@cartItems": {
    "description": "Item count in cart",
    "placeholders": {
      "count": {
        "type": "int"
      }
    }
  }
}`

	p, err := parser.GetParser("flutter_arb")
	if err != nil {
		t.Fatalf("failed to get arb parser: %v", err)
	}

	cat, err := p.ParseBytes([]byte(rawARB), "en")
	if err != nil {
		t.Fatalf("failed to parse ARB: %v", err)
	}

	if cat.Locale != "en" {
		t.Errorf("expected locale 'en', got '%s'", cat.Locale)
	}

	e, ok := cat.Get("cartItems")
	if !ok || e.Comment != "Item count in cart" {
		t.Fatalf("unexpected entry cartItems: %v", e)
	}
	if _, ok := e.Placeholders["count"]; !ok {
		t.Fatalf("expected count placeholder, got %v", e.Placeholders)
	}

	// Round trip serialization
	serialized, err := p.Serialize(cat)
	if err != nil {
		t.Fatalf("failed to serialize ARB: %v", err)
	}
	if !strings.Contains(string(serialized), `"@@locale": "en"`) || !strings.Contains(string(serialized), `"description": "Item count in cart"`) {
		t.Errorf("serialized ARB missing expected fields: %s", string(serialized))
	}
}

func TestCSVParser(t *testing.T) {
	csvData := `Key,Target,Context,Comment
nav.save,Speichern,toolbar,Save button
nav.cancel,Abbrechen,toolbar,Cancel button
`
	p, err := parser.GetParser("csv")
	if err != nil {
		t.Fatalf("get csv parser: %v", err)
	}

	cat, err := p.ParseBytes([]byte(csvData), "de")
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}

	if len(cat.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(cat.Entries))
	}

	e, ok := cat.Get("nav.save")
	if !ok || e.Value != "Speichern" || e.Context != "toolbar" {
		t.Fatalf("unexpected entry: %v", e)
	}

	serialized, err := p.Serialize(cat)
	if err != nil {
		t.Fatalf("serialize csv: %v", err)
	}
	if !strings.Contains(string(serialized), "nav.save,Speichern,toolbar,Save button") {
		t.Fatalf("serialized csv missing row: %s", string(serialized))
	}
}

func TestXLIFFParser(t *testing.T) {
	xliffData := `<?xml version="1.0" encoding="UTF-8"?>
<xliff version="1.2">
  <file source-language="en" target-language="de" datatype="plaintext" original="locales/en.json">
    <body>
      <trans-unit id="nav.checkout">
        <source>Checkout</source>
        <target>Zur Kasse</target>
        <note>Primary action</note>
      </trans-unit>
    </body>
  </file>
</xliff>`

	p, err := parser.GetParser("xliff")
	if err != nil {
		t.Fatalf("get xliff parser: %v", err)
	}

	cat, err := p.ParseBytes([]byte(xliffData), "de")
	if err != nil {
		t.Fatalf("parse xliff: %v", err)
	}

	if len(cat.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(cat.Entries))
	}

	e, ok := cat.Get("nav.checkout")
	if !ok || e.Value != "Zur Kasse" || e.Comment != "Primary action" {
		t.Fatalf("unexpected xliff entry: %v", e)
	}

	serialized, err := p.Serialize(cat)
	if err != nil {
		t.Fatalf("serialize xliff: %v", err)
	}
	if !strings.Contains(string(serialized), `<trans-unit id="nav.checkout">`) {
		t.Fatalf("serialized xliff missing trans-unit: %s", string(serialized))
	}
}

func TestDetectFormat(t *testing.T) {
	cases := []struct {
		path     string
		expected string
	}{
		{"locales/en.json", "json_nested"},
		{"messages.de.po", "gettext_po"},
		{"template.pot", "gettext_po"},
		{"app_en.arb", "flutter_arb"},
		{"something.unknown", "json_nested"},
	}

	for _, c := range cases {
		got := parser.DetectFormat(filepath.FromSlash(c.path))
		if got != c.expected {
			t.Errorf("for path %s expected format %s, got %s", c.path, c.expected, got)
		}
	}
}
