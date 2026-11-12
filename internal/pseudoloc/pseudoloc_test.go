package pseudoloc_test

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/localecraft/internal/parser"
	"github.com/alexandrmotologa/localecraft/internal/pseudoloc"
)

func TestTransformSimpleString(t *testing.T) {
	opts := pseudoloc.DefaultOptions()
	res := pseudoloc.Transform("Save", opts)

	if !strings.HasPrefix(res, "[") || !strings.HasSuffix(res, "]") {
		t.Fatalf("expected brackets around result, got '%s'", res)
	}

	// 'S' -> 'Š', 'a' -> 'å', 'v' -> 'ṽ', 'e' -> 'é'
	if !strings.Contains(res, "Šåṽé") {
		t.Fatalf("expected accented transformation 'Šåṽé', got '%s'", res)
	}

	// Should be longer than original due to +35% expansion
	if len([]rune(res)) <= len("Save") {
		t.Fatalf("expected expanded length, got %d <= %d", len([]rune(res)), len("Save"))
	}
}

func TestTransformPreservesICUTokens(t *testing.T) {
	opts := pseudoloc.DefaultOptions()
	input := "Welcome {userName}, you have {taskCount} items."
	res := pseudoloc.Transform(input, opts)

	if !strings.Contains(res, "{userName}") {
		t.Fatalf("expected {userName} to be preserved untouched, got '%s'", res)
	}
	if !strings.Contains(res, "{taskCount}") {
		t.Fatalf("expected {taskCount} to be preserved untouched, got '%s'", res)
	}
}

func TestTransformCatalog(t *testing.T) {
	cat := parser.NewCatalog("json_nested", "en", "en.json")
	cat.Set(&parser.Entry{Key: "btn.save", Value: "Save"})
	cat.Set(&parser.Entry{Key: "nav.title", Value: "Dashboard"})

	pseudoCat := pseudoloc.TransformCatalog(cat, pseudoloc.DefaultOptions())

	if len(pseudoCat.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(pseudoCat.Entries))
	}

	btn, ok := pseudoCat.Get("btn.save")
	if !ok || !strings.Contains(btn.Value, "Šåṽé") {
		t.Fatalf("unexpected pseudo value: %v", btn)
	}
}
