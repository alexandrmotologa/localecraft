package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// GettextPOParser parses and serializes GNU gettext .po and .pot files.
type GettextPOParser struct{}

// ParseBytes parses GNU gettext PO format byte slice.
func (p *GettextPOParser) ParseBytes(data []byte, locale string) (*Catalog, error) {
	catalog := NewCatalog("gettext_po", locale, "")
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var currentComments []string
	var currentContext string
	var currentMsgID string
	var currentMsgStr string
	var currentPluralID string
	var pluralStrs []string
	var inState string // "msgctxt", "msgid", "msgid_plural", "msgstr", "msgstr_plural"
	var currentPluralIdx int

	commitCurrent := func() {
		if currentMsgID != "" {
			entryKey := currentMsgID
			if currentContext != "" {
				entryKey = currentContext + "\x04" + currentMsgID
			}

			entry := &Entry{
				Key:         entryKey,
				Value:       currentMsgStr,
				Context:     currentContext,
				Comment:     strings.Join(currentComments, "\n"),
				PluralForms: pluralStrs,
			}
			if currentPluralID != "" {
				if entry.RawMeta == nil {
					entry.RawMeta = make(map[string]any)
				}
				entry.RawMeta["msgid_plural"] = currentPluralID
			}
			catalog.Set(entry)
		}
		currentComments = nil
		currentContext = ""
		currentMsgID = ""
		currentMsgStr = ""
		currentPluralID = ""
		pluralStrs = nil
		inState = ""
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			commitCurrent()
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			currentComments = append(currentComments, trimmed)
			continue
		}

		if strings.HasPrefix(trimmed, "msgctxt ") {
			inState = "msgctxt"
			currentContext = parseQuotedString(strings.TrimPrefix(trimmed, "msgctxt "))
			continue
		}

		if strings.HasPrefix(trimmed, "msgid_plural ") {
			inState = "msgid_plural"
			currentPluralID = parseQuotedString(strings.TrimPrefix(trimmed, "msgid_plural "))
			continue
		}

		if strings.HasPrefix(trimmed, "msgid ") {
			inState = "msgid"
			currentMsgID = parseQuotedString(strings.TrimPrefix(trimmed, "msgid "))
			continue
		}

		if strings.HasPrefix(trimmed, "msgstr ") {
			inState = "msgstr"
			currentMsgStr = parseQuotedString(strings.TrimPrefix(trimmed, "msgstr "))
			continue
		}

		if strings.HasPrefix(trimmed, "msgstr[") {
			inState = "msgstr_plural"
			closeIdx := strings.Index(trimmed, "]")
			if closeIdx > 7 {
				idxStr := trimmed[7:closeIdx]
				idx, err := strconv.Atoi(idxStr)
				if err == nil {
					currentPluralIdx = idx
					for len(pluralStrs) <= currentPluralIdx {
						pluralStrs = append(pluralStrs, "")
					}
					rest := strings.TrimSpace(trimmed[closeIdx+1:])
					pluralStrs[currentPluralIdx] = parseQuotedString(rest)
				}
			}
			continue
		}

		if strings.HasPrefix(trimmed, "\"") && inState != "" {
			quotedPart := parseQuotedString(trimmed)
			switch inState {
			case "msgctxt":
				currentContext += quotedPart
			case "msgid":
				currentMsgID += quotedPart
			case "msgid_plural":
				currentPluralID += quotedPart
			case "msgstr":
				currentMsgStr += quotedPart
			case "msgstr_plural":
				if currentPluralIdx < len(pluralStrs) {
					pluralStrs[currentPluralIdx] += quotedPart
				}
			}
			continue
		}
	}

	commitCurrent()
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan po file: %w", err)
	}

	return catalog, nil
}

func parseQuotedString(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "\"") {
		return s
	}
	unquoted, err := strconv.Unquote(s)
	if err != nil {
		// Fallback manual trim if strconv fails on edge cases
		if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") && len(s) >= 2 {
			inner := s[1 : len(s)-1]
			inner = strings.ReplaceAll(inner, `\"`, `"`)
			inner = strings.ReplaceAll(inner, `\n`, "\n")
			inner = strings.ReplaceAll(inner, `\t`, "\t")
			inner = strings.ReplaceAll(inner, `\\`, `\`)
			return inner
		}
		return s
	}
	return unquoted
}

func formatPOString(s string) string {
	if !strings.Contains(s, "\n") {
		return strconv.Quote(s)
	}
	lines := strings.Split(s, "\n")
	var buf strings.Builder
	buf.WriteString("\"\"\n")
	for i, l := range lines {
		if i < len(lines)-1 {
			buf.WriteString(fmt.Sprintf("%s\n", strconv.Quote(l+"\n")))
		} else if l != "" {
			buf.WriteString(fmt.Sprintf("%s\n", strconv.Quote(l)))
		}
	}
	return strings.TrimSuffix(buf.String(), "\n")
}

// ParseFile parses a .po file from disk.
func (p *GettextPOParser) ParseFile(path string, locale string) (*Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read po file %s: %w", path, err)
	}
	cat, err := p.ParseBytes(data, locale)
	if err != nil {
		return nil, err
	}
	cat.Path = path
	return cat, nil
}

// Serialize converts a catalog into standard gettext PO file text.
func (p *GettextPOParser) Serialize(catalog *Catalog) ([]byte, error) {
	var buf strings.Builder

	// Header comments
	buf.WriteString("# Generated by localecraft\n")
	buf.WriteString(fmt.Sprintf("# Target locale: %s\n", catalog.Locale))
	buf.WriteString(`msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"Language: ` + catalog.Locale + `\n"

`)

	for _, key := range catalog.KeyOrder {
		entry, ok := catalog.Entries[key]
		if !ok || entry.Key == "" {
			continue
		}

		if entry.Comment != "" {
			for _, commentLine := range strings.Split(entry.Comment, "\n") {
				if !strings.HasPrefix(commentLine, "#") {
					buf.WriteString("#. " + commentLine + "\n")
				} else {
					buf.WriteString(commentLine + "\n")
				}
			}
		}

		if entry.Context != "" {
			buf.WriteString(fmt.Sprintf("msgctxt %s\n", formatPOString(entry.Context)))
		}

		cleanMsgID := entry.Key
		if entry.Context != "" && strings.Contains(cleanMsgID, "\x04") {
			parts := strings.SplitN(cleanMsgID, "\x04", 2)
			cleanMsgID = parts[1]
		}

		buf.WriteString(fmt.Sprintf("msgid %s\n", formatPOString(cleanMsgID)))

		if len(entry.PluralForms) > 0 {
			pluralID := cleanMsgID
			if entry.RawMeta != nil {
				if pid, ok := entry.RawMeta["msgid_plural"].(string); ok && pid != "" {
					pluralID = pid
				}
			}
			buf.WriteString(fmt.Sprintf("msgid_plural %s\n", formatPOString(pluralID)))
			for idx, pf := range entry.PluralForms {
				buf.WriteString(fmt.Sprintf("msgstr[%d] %s\n", idx, formatPOString(pf)))
			}
		} else {
			buf.WriteString(fmt.Sprintf("msgstr %s\n", formatPOString(entry.Value)))
		}

		buf.WriteString("\n")
	}

	return []byte(buf.String()), nil
}

// WriteFile writes the serialized PO content to a file.
func (p *GettextPOParser) WriteFile(catalog *Catalog, path string) error {
	data, err := p.Serialize(catalog)
	if err != nil {
		return err
	}
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
