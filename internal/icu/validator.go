package icu

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidationResult summarizes the ICU syntax and parity checks between source and target.
type ValidationResult struct {
	Valid          bool     `json:"valid"`
	Errors         []string `json:"errors"`
	Warnings       []string `json:"warnings"`
	SourceVars     []string `json:"source_vars"`
	TargetVars     []string `json:"target_vars"`
	MissingVars    []string `json:"missing_vars"`
	UnexpectedVars []string `json:"unexpected_vars"`
}

// Allowed plural categories by locale according to Unicode CLDR rules.
var cldrPluralCategories = map[string][]string{
	"en": {"one", "other"},
	"de": {"one", "other"},
	"fr": {"one", "other"},
	"es": {"one", "other"},
	"it": {"one", "other"},
	"pt": {"one", "other"},
	"nl": {"one", "other"},
	"sv": {"one", "other"},
	"da": {"one", "other"},
	"no": {"one", "other"},
	"ja": {"other"},
	"zh": {"other"},
	"ko": {"other"},
	"vi": {"other"},
	"ru": {"one", "few", "many", "other"},
	"uk": {"one", "few", "many", "other"},
	"be": {"one", "few", "many", "other"},
	"pl": {"one", "few", "many", "other"},
	"cs": {"one", "few", "many", "other"},
	"sk": {"one", "few", "many", "other"},
	"ar": {"zero", "one", "two", "few", "many", "other"},
	"he": {"one", "two", "many", "other"},
	"ro": {"one", "few", "other"},
}

// Validate compares a source ICU string and its target translation for structural and variable parity.
func Validate(sourceText string, targetText string, targetLocale string) ValidationResult {
	res := ValidationResult{
		Valid:          true,
		Errors:         make([]string, 0),
		Warnings:       make([]string, 0),
		MissingVars:    make([]string, 0),
		UnexpectedVars: make([]string, 0),
	}

	srcMsg, srcErr := Parse(sourceText)
	if srcErr != nil {
		res.Valid = false
		res.Errors = append(res.Errors, fmt.Sprintf("Source ICU syntax error: %v", srcErr))
		return res
	}

	tgtMsg, tgtErr := Parse(targetText)
	if tgtErr != nil {
		res.Valid = false
		res.Errors = append(res.Errors, fmt.Sprintf("Target ICU syntax error: %v", tgtErr))
		return res
	}

	res.SourceVars = srcMsg.Variables()
	res.TargetVars = tgtMsg.Variables()

	// Check for missing variables
	srcVarMap := make(map[string]bool)
	for _, v := range res.SourceVars {
		srcVarMap[v] = true
	}

	tgtVarMap := make(map[string]bool)
	for _, v := range res.TargetVars {
		tgtVarMap[v] = true
	}

	for _, sv := range res.SourceVars {
		if !tgtVarMap[sv] {
			res.MissingVars = append(res.MissingVars, sv)
			res.Valid = false
			res.Errors = append(res.Errors, fmt.Sprintf("Missing variable '{%s}' in target translation", sv))
		}
	}

	for _, tv := range res.TargetVars {
		if !srcVarMap[tv] {
			res.UnexpectedVars = append(res.UnexpectedVars, tv)
			res.Valid = false
			res.Errors = append(res.Errors, fmt.Sprintf("Unexpected variable '{%s}' found in target translation", tv))
		}
	}

	// Validate plural nodes in target
	baseLocale := strings.ToLower(strings.Split(targetLocale, "-")[0])
	validateNodePlurals(tgtMsg.Nodes, baseLocale, &res)

	return res
}

func validateNodePlurals(nodes []Node, locale string, res *ValidationResult) {
	for _, n := range nodes {
		switch node := n.(type) {
		case *PluralNode:
			hasOther := false
			allowedCategories := cldrPluralCategories[locale]
			if len(allowedCategories) == 0 {
				allowedCategories = []string{"one", "other"}
			}

			allowedMap := make(map[string]bool)
			for _, cat := range allowedCategories {
				allowedMap[cat] = true
			}

			for _, c := range node.Cases {
				if c.Selector == "other" {
					hasOther = true
				}

				// Exact counts like =0, =1 are always valid
				if strings.HasPrefix(c.Selector, "=") {
					continue
				}

				if !allowedMap[c.Selector] {
					res.Warnings = append(res.Warnings, fmt.Sprintf("Plural selector '%s' is not standard for locale '%s'", c.Selector, locale))
				}

				if c.Message != nil {
					validateNodePlurals(c.Message.Nodes, locale, res)
				}
			}

			if !hasOther {
				res.Valid = false
				res.Errors = append(res.Errors, fmt.Sprintf("Plural argument '{%s}' is missing mandatory 'other' selector", node.Name))
			}

		case *SelectNode:
			hasOther := false
			for _, c := range node.Cases {
				if c.Selector == "other" {
					hasOther = true
				}
				if c.Message != nil {
					validateNodePlurals(c.Message.Nodes, locale, res)
				}
			}
			if !hasOther {
				res.Valid = false
				res.Errors = append(res.Errors, fmt.Sprintf("Select argument '{%s}' is missing mandatory 'other' selector", node.Name))
			}
		}
	}
}

// Evaluate renders an ICU message AST with concrete parameters and basic CLDR plural rules.
func Evaluate(msg *Message, params map[string]any, locale string) (string, error) {
	var sb strings.Builder
	baseLocale := strings.ToLower(strings.Split(locale, "-")[0])

	for _, n := range msg.Nodes {
		switch node := n.(type) {
		case *TextNode:
			sb.WriteString(node.Text)

		case *ArgumentNode:
			val, ok := params[node.Name]
			if !ok {
				sb.WriteString("{" + node.Name + "}")
			} else {
				sb.WriteString(fmt.Sprintf("%v", val))
			}

		case *PluralNode:
			numVal := 0.0
			rawVal, ok := params[node.Name]
			if ok {
				switch v := rawVal.(type) {
				case int:
					numVal = float64(v)
				case int64:
					numVal = float64(v)
				case float64:
					numVal = v
				case string:
					parsed, _ := strconv.ParseFloat(v, 64)
					numVal = parsed
				}
			}

			effectiveVal := numVal - float64(node.Offset)
			cat := getPluralCategory(effectiveVal, baseLocale)

			// 1. Try exact match (=0, =1)
			exactSelector := fmt.Sprintf("=%d", int(effectiveVal))
			var matchedCase *CaseClause
			for i := range node.Cases {
				if node.Cases[i].Selector == exactSelector {
					matchedCase = &node.Cases[i]
					break
				}
			}

			// 2. Try CLDR category match (one, few, many)
			if matchedCase == nil {
				for i := range node.Cases {
					if node.Cases[i].Selector == cat {
						matchedCase = &node.Cases[i]
						break
					}
				}
			}

			// 3. Fallback to 'other'
			if matchedCase == nil {
				for i := range node.Cases {
					if node.Cases[i].Selector == "other" {
						matchedCase = &node.Cases[i]
						break
					}
				}
			}

			if matchedCase != nil && matchedCase.Message != nil {
				subRendered, err := Evaluate(matchedCase.Message, params, locale)
				if err != nil {
					return "", err
				}
				// Replace '#' with formatted number
				numStr := fmt.Sprintf("%v", effectiveVal)
				if effectiveVal == float64(int64(effectiveVal)) {
					numStr = fmt.Sprintf("%d", int64(effectiveVal))
				}
				subRendered = strings.ReplaceAll(subRendered, "#", numStr)
				sb.WriteString(subRendered)
			}

		case *SelectNode:
			valStr := ""
			if v, ok := params[node.Name]; ok {
				valStr = fmt.Sprintf("%v", v)
			}

			var matchedCase *CaseClause
			for i := range node.Cases {
				if node.Cases[i].Selector == valStr {
					matchedCase = &node.Cases[i]
					break
				}
			}
			if matchedCase == nil {
				for i := range node.Cases {
					if node.Cases[i].Selector == "other" {
						matchedCase = &node.Cases[i]
						break
					}
				}
			}

			if matchedCase != nil && matchedCase.Message != nil {
				subRendered, err := Evaluate(matchedCase.Message, params, locale)
				if err != nil {
					return "", err
				}
				sb.WriteString(subRendered)
			}
		}
	}

	return sb.String(), nil
}

func getPluralCategory(n float64, locale string) string {
	intPart := int(n)
	isInteger := n == float64(intPart)

	switch locale {
	case "ru", "uk", "be":
		if isInteger {
			mod10 := intPart % 10
			mod100 := intPart % 100
			if mod10 == 1 && mod100 != 11 {
				return "one"
			}
			if (mod10 >= 2 && mod10 <= 4) && (mod100 < 12 || mod100 > 14) {
				return "few"
			}
			if mod10 == 0 || (mod10 >= 5 && mod10 <= 9) || (mod100 >= 11 && mod100 <= 14) {
				return "many"
			}
		}
		return "other"

	case "ar":
		if isInteger {
			if intPart == 0 {
				return "zero"
			}
			if intPart == 1 {
				return "one"
			}
			if intPart == 2 {
				return "two"
			}
			mod100 := intPart % 100
			if mod100 >= 3 && mod100 <= 10 {
				return "few"
			}
			if mod100 >= 11 && mod100 <= 99 {
				return "many"
			}
		}
		return "other"

	case "ro":
		if isInteger {
			if intPart == 1 {
				return "one"
			}
			mod100 := intPart % 100
			if intPart == 0 || (mod100 >= 1 && mod100 <= 19) {
				return "few"
			}
		}
		return "other"

	default: // en, de, es, fr, it
		if n == 1 {
			return "one"
		}
		return "other"
	}
}

// CheckSyntax runs a fast parse check on a single string and returns true if it is valid ICU.
func CheckSyntax(text string) (bool, error) {
	_, err := Parse(text)
	if err != nil {
		return false, err
	}
	return true, nil
}
