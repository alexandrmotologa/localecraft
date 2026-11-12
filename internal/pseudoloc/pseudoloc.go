package pseudoloc

import (
	"strings"
	"unicode/utf8"

	"github.com/alexandrmotologa/localecraft/internal/parser"
)

var charMap = map[rune]rune{
	'a': 'å', 'b': 'ƀ', 'c': 'ç', 'd': 'ð', 'e': 'é', 'f': 'ƒ', 'g': 'ĝ', 'h': 'ĥ',
	'i': 'î', 'j': 'ĵ', 'k': 'ƙ', 'l': 'ł', 'm': 'ɱ', 'n': 'ñ', 'o': 'ø', 'p': 'þ',
	'q': 'q', 'r': 'ř', 's': 'š', 't': 'ŧ', 'u': 'ü', 'v': 'ṽ', 'w': 'ŵ', 'x': 'ҳ',
	'y': 'ý', 'z': 'ž',
	'A': 'Å', 'B': 'Ɓ', 'C': 'Ç', 'D': 'Ð', 'E': 'É', 'F': 'Ƒ', 'G': 'Ĝ', 'H': 'Ĥ',
	'I': 'Î', 'J': 'Ĵ', 'K': 'Ќ', 'L': 'Ł', 'M': 'Ṁ', 'N': 'Ñ', 'O': 'Ø', 'P': 'Þ',
	'Q': 'Q', 'R': 'Ř', 'S': 'Š', 'T': 'Ŧ', 'U': 'Ü', 'V': 'Ṽ', 'W': 'Ŵ', 'X': 'Ҳ',
	'Y': 'Ý', 'Z': 'Ž',
}

// Options configures pseudo-localization transformation rules.
type Options struct {
	ExpansionPCT int  // percentage to expand (e.g., 35 for +35%)
	Brackets     bool // wrap with [ ... ]
}

// DefaultOptions returns baseline pseudo-localization options (+35% expansion with brackets).
func DefaultOptions() Options {
	return Options{
		ExpansionPCT: 35,
		Brackets:     true,
	}
}

// Transform converts an English string into an accented pseudo-localized string with expansion.
func Transform(text string, opt Options) string {
	if text == "" {
		return ""
	}

	var sb strings.Builder
	runes := []rune(text)
	inBrace := 0
	inTag := false

	charCount := 0

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// ICU Placeholder guard: {name} or {count, plural, ...}
		if r == '{' {
			inBrace++
			sb.WriteRune(r)
			continue
		}
		if r == '}' && inBrace > 0 {
			inBrace--
			sb.WriteRune(r)
			continue
		}

		// HTML Tag guard: <span class="...">
		if r == '<' {
			inTag = true
			sb.WriteRune(r)
			continue
		}
		if r == '>' && inTag {
			inTag = false
			sb.WriteRune(r)
			continue
		}

		if inBrace > 0 || inTag {
			sb.WriteRune(r)
			continue
		}

		charCount++
		if mapped, ok := charMap[r]; ok {
			sb.WriteRune(mapped)
		} else {
			sb.WriteRune(r)
		}
	}

	transformed := sb.String()

	// Calculate expansion padding
	if opt.ExpansionPCT > 0 && charCount > 0 {
		extraChars := (charCount * opt.ExpansionPCT) / 100
		if extraChars < 2 {
			extraChars = 2
		}

		padding := generateExpansionPadding(extraChars)
		transformed = transformed + " " + padding
	}

	if opt.Brackets {
		return "[" + transformed + "]"
	}

	return transformed
}

func generateExpansionPadding(length int) string {
	fillers := []string{"öñê", "ŧŵø", "ŧħřéé", "ƒøüř", "ƒîṽé", "šîҳ", "šéṽêñ"}
	var sb strings.Builder
	idx := 0

	for utf8.RuneCountInString(sb.String()) < length {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fillers[idx%len(fillers)])
		idx++
	}

	resRunes := []rune(sb.String())
	if len(resRunes) > length {
		return string(resRunes[:length])
	}
	return sb.String()
}

// TransformCatalog creates a pseudo-localized copy of a catalog.
func TransformCatalog(src *parser.Catalog, opt Options) *parser.Catalog {
	pseudo := parser.NewCatalog(src.Format, "pseudo", "")

	for _, key := range src.KeyOrder {
		entry, ok := src.Entries[key]
		if !ok {
			continue
		}

		pseudoEntry := &parser.Entry{
			Key:          entry.Key,
			Value:        Transform(entry.Value, opt),
			Comment:      entry.Comment,
			Context:      entry.Context,
			Placeholders: entry.Placeholders,
		}

		if len(entry.PluralForms) > 0 {
			pseudoEntry.PluralForms = make([]string, len(entry.PluralForms))
			for i, pf := range entry.PluralForms {
				pseudoEntry.PluralForms[i] = Transform(pf, opt)
			}
		}

		pseudo.Set(pseudoEntry)
	}

	return pseudo
}
