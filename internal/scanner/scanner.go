package scanner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/alexandrmotologa/localecraft/internal/config"
)

// InferredConstraint stores extracted UI limits and source code location.
type InferredConstraint struct {
	Key        string  `json:"key"`
	MaxWidthPX float64 `json:"max_width_px"`
	MaxLines   int     `json:"max_lines"`
	FontSizePX float64 `json:"font_size_px"`
	FontWeight int     `json:"font_weight"`
	Context    string  `json:"context"`
	FilePath   string  `json:"file_path"`
	LineNumber int     `json:"line_number"`
}

var (
	// Translation key extraction regexes
	tRegex             = regexp.MustCompile(`\bt\(\s*['"]([a-zA-Z0-9_.-]+)['"]\s*\)`)
	i18nTRegex         = regexp.MustCompile(`\bi18n\.t\(\s*['"]([a-zA-Z0-9_.-]+)['"]\s*\)`)
	formattedMsgRegex  = regexp.MustCompile(`<FormattedMessage[^>]*id=['"]([a-zA-Z0-9_.-]+)['"]`)
	transTagRegex      = regexp.MustCompile(`<Trans[^>]*i18nKey=['"]([a-zA-Z0-9_.-]+)['"]`)
	formatMsgCallRegex = regexp.MustCompile(`formatMessage\(\s*\{\s*id:\s*['"]([a-zA-Z0-9_.-]+)['"]`)

	// Tailwind width mappings
	tailwindWidths = map[string]float64{
		"w-12": 48, "w-16": 64, "w-20": 80, "w-24": 96, "w-28": 112,
		"w-32": 128, "w-36": 144, "w-40": 160, "w-44": 176, "w-48": 192,
		"w-52": 208, "w-56": 224, "w-60": 240, "w-64": 256, "w-72": 288,
		"w-80": 320, "w-96": 384,
		"max-w-xs": 320, "max-w-sm": 384, "max-w-md": 448, "max-w-lg": 512,
	}

	// Arbitrary Tailwind regex: max-w-[140px] or w-[120px]
	arbitraryWidthRegex = regexp.MustCompile(`(?:max-)?w-\[(\d+)px\]`)
	inlineStyleRegex    = regexp.MustCompile(`(?:max-?)?[wW]idth\s*[:=]\s*['"]?(\d+)px['"]?`)
)

// ScanDirectory walks a directory and inspects frontend source files.
func ScanDirectory(root string) ([]InferredConstraint, error) {
	var results []InferredConstraint
	seenKeys := make(map[string]bool)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "dist" || name == "build" || name == ".next" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".tsx" || ext == ".jsx" || ext == ".js" || ext == ".ts" || ext == ".vue" || ext == ".svelte" || ext == ".html" {
			extracted, err := scanFile(path)
			if err == nil {
				for _, c := range extracted {
					if !seenKeys[c.Key] {
						seenKeys[c.Key] = true
						results = append(results, c)
					}
				}
			}
		}

		return nil
	})

	return results, err
}

func scanFile(path string) ([]InferredConstraint, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var constraints []InferredConstraint
	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	for lineIdx, line := range lines {
		// Find translation keys in this line
		keys := extractKeys(line)
		if len(keys) == 0 {
			continue
		}

		// Priority 1: Current line containing the key
		currentLine := lines[lineIdx]
		widthPX := extractWidth(currentLine)
		fontSizePX := extractFontSize(currentLine)
		fontWeight := extractFontWeight(currentLine)
		maxLines := extractMaxLines(currentLine)

		// Priority 2: Walk backward line by line to find closest parent tag styles
		if widthPX == 0 || fontSizePX == 0 || fontWeight == 0 {
			startIdx := lineIdx - 3
			if startIdx < 0 {
				startIdx = 0
			}
			for back := lineIdx - 1; back >= startIdx; back-- {
				prevLine := lines[back]
				if widthPX == 0 {
					widthPX = extractWidth(prevLine)
				}
				if fontSizePX == 0 {
					fontSizePX = extractFontSize(prevLine)
				}
				if fontWeight == 0 {
					fontWeight = extractFontWeight(prevLine)
				}
				if maxLines == 1 {
					maxLines = extractMaxLines(prevLine)
				}
				// If we found both width and typography in this closest line, stop
				if widthPX > 0 && fontSizePX > 0 && fontWeight > 0 {
					break
				}
			}
		}

		if fontSizePX == 0 {
			fontSizePX = 14
		}
		if fontWeight == 0 {
			fontWeight = 500
		}

		for _, k := range keys {
			constraints = append(constraints, InferredConstraint{
				Key:        k,
				MaxWidthPX: widthPX,
				MaxLines:   maxLines,
				FontSizePX: fontSizePX,
				FontWeight: fontWeight,
				Context:    fmt.Sprintf("Extracted from %s:%d", filepath.Base(path), lineIdx+1),
				FilePath:   path,
				LineNumber: lineIdx + 1,
			})
		}
	}

	return constraints, nil
}

func extractFontSize(str string) float64 {
	if strings.Contains(str, "text-xs") {
		return 12
	}
	if strings.Contains(str, "text-sm") {
		return 14
	}
	if strings.Contains(str, "text-base") {
		return 16
	}
	if strings.Contains(str, "text-lg") {
		return 18
	}
	if strings.Contains(str, "text-xl") {
		return 20
	}
	return 0
}

func extractFontWeight(str string) int {
	if strings.Contains(str, "font-bold") {
		return 700
	}
	if strings.Contains(str, "font-semibold") {
		return 600
	}
	if strings.Contains(str, "font-medium") {
		return 500
	}
	if strings.Contains(str, "font-normal") {
		return 400
	}
	return 0
}

func extractMaxLines(str string) int {
	if strings.Contains(str, "line-clamp-3") {
		return 3
	}
	if strings.Contains(str, "line-clamp-2") {
		return 2
	}
	return 1
}

func extractKeys(line string) []string {
	var keys []string

	for _, re := range []*regexp.Regexp{tRegex, i18nTRegex, formattedMsgRegex, transTagRegex, formatMsgCallRegex} {
		matches := re.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			if len(m) > 1 {
				keys = append(keys, m[1])
			}
		}
	}

	return keys
}

func extractWidth(contextStr string) float64 {
	// 1. Check arbitrary width: w-[140px]
	if m := arbitraryWidthRegex.FindStringSubmatch(contextStr); len(m) > 1 {
		if val, err := strconv.ParseFloat(m[1], 64); err == nil {
			return val
		}
	}

	// 2. Check inline style: style={{ maxWidth: '140px' }}
	if m := inlineStyleRegex.FindStringSubmatch(contextStr); len(m) > 1 {
		if val, err := strconv.ParseFloat(m[1], 64); err == nil {
			return val
		}
	}

	// 3. Check standard Tailwind class
	cleanedStr := strings.ReplaceAll(contextStr, "=", " ")
	cleanedStr = strings.ReplaceAll(cleanedStr, "\"", " ")
	cleanedStr = strings.ReplaceAll(cleanedStr, "'", " ")
	words := strings.Fields(cleanedStr)
	for _, w := range words {
		if val, exists := tailwindWidths[w]; exists {
			return val
		}
	}

	return 0
}

// UpdateConfigWithInferred merges scanned constraints into an existing Config.
func UpdateConfigWithInferred(cfg *config.Config, constraints []InferredConstraint) int {
	if cfg.UIConstraints.Keys == nil {
		cfg.UIConstraints.Keys = make(map[string]config.KeyConstraint)
	}

	addedCount := 0
	for _, c := range constraints {
		existing, exists := cfg.UIConstraints.Keys[c.Key]
		if !exists {
			cfg.UIConstraints.Keys[c.Key] = config.KeyConstraint{
				MaxWidthPX: c.MaxWidthPX,
				MaxLines:   c.MaxLines,
				FontSizePX: c.FontSizePX,
				FontWeight: c.FontWeight,
				Context:    c.Context,
			}
			addedCount++
		} else {
			// Update max width if existing was unbounded
			if existing.MaxWidthPX <= 0 && c.MaxWidthPX > 0 {
				existing.MaxWidthPX = c.MaxWidthPX
				cfg.UIConstraints.Keys[c.Key] = existing
				addedCount++
			}
		}
	}

	return addedCount
}
