package agent

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

type termVariants struct {
	standard string
	short    string
	compact  string
}

// Built-in lexical database for common UI words across major locales.
var dictionary = map[string]map[string]termVariants{
	"save": {
		"de": {standard: "Speichern und anwenden", short: "Speichern", compact: "Sichern"},
		"fr": {standard: "Enregistrer les modifications", short: "Enregistrer", compact: "Sauver"},
		"es": {standard: "Guardar y aplicar", short: "Guardar", compact: "OK"},
		"ja": {standard: "保存して適用する", short: "保存する", compact: "保存"},
		"ro": {standard: "Salvează și aplică", short: "Salvează", compact: "Salvare"},
	},
	"save changes": {
		"de": {standard: "Änderungen speichern", short: "Speichern", compact: "Sichern"},
		"fr": {standard: "Enregistrer les modifications", short: "Enregistrer", compact: "Sauver"},
		"es": {standard: "Guardar cambios", short: "Guardar", compact: "OK"},
		"ja": {standard: "変更を保存する", short: "変更を保存", compact: "保存"},
		"ro": {standard: "Salvează modificările", short: "Salvează", compact: "Salvare"},
	},
	"checkout": {
		"de": {standard: "Zur Kasse gehen", short: "Zur Kasse", compact: "Kasse"},
		"fr": {standard: "Passer à la caisse", short: "Commander", compact: "Caisse"},
		"es": {standard: "Proceder al pago", short: "Finalizar", compact: "Pagar"},
		"ja": {standard: "チェックアウトに進む", short: "お会計へ", compact: "会計"},
		"ro": {standard: "Finalizează comanda", short: "La casă", compact: "Plată"},
	},
	"cancel": {
		"de": {standard: "Vorgang abbrechen", short: "Abbrechen", compact: "Stopp"},
		"fr": {standard: "Annuler l'opération", short: "Annuler", compact: "Stop"},
		"es": {standard: "Cancelar operación", short: "Cancelar", compact: "Fin"},
		"ja": {standard: "操作をキャンセル", short: "キャンセル", compact: "中止"},
		"ro": {standard: "Anulează operațiunea", short: "Anulează", compact: "Stop"},
	},
	"delete account": {
		"de": {standard: "Benutzerkonto endgültig löschen", short: "Konto löschen", compact: "Löschen"},
		"fr": {standard: "Supprimer définitivement le compte", short: "Supprimer le compte", compact: "Supprimer"},
		"es": {standard: "Eliminar cuenta permanentemente", short: "Eliminar cuenta", compact: "Borrar"},
		"ja": {standard: "アカウントを完全に削除する", short: "アカウント削除", compact: "削除"},
		"ro": {standard: "Șterge definitiv contul", short: "Șterge contul", compact: "Șterge"},
	},
	"submit": {
		"de": {standard: "Formular jetzt absenden", short: "Absenden", compact: "Senden"},
		"fr": {standard: "Soumettre le formulaire", short: "Soumettre", compact: "Envoyer"},
		"es": {standard: "Enviar formulario ahora", short: "Enviar", compact: "OK"},
		"ja": {standard: "フォームを送信する", short: "送信する", compact: "送信"},
		"ro": {standard: "Trimite formularul acum", short: "Trimite", compact: "OK"},
	},
	"workspace": {
		"de": {standard: "Zentraler Arbeitsbereich", short: "Arbeitsbereich", compact: "Bereich"},
		"fr": {standard: "Espace de travail principal", short: "Espace de travail", compact: "Espace"},
		"es": {standard: "Espacio de trabajo", short: "Espacio", compact: "Área"},
		"ja": {standard: "ワークスペース", short: "ワークスペース", compact: "作業場"},
		"ro": {standard: "Spațiu de lucru", short: "Spațiu lucru", compact: "Spațiu"},
	},
}

// HeuristicProvider provides deterministic translations for testing and offline environments.
type HeuristicProvider struct{}

// NewHeuristicProvider creates an instance of HeuristicProvider.
func NewHeuristicProvider() *HeuristicProvider {
	return &HeuristicProvider{}
}

func (h *HeuristicProvider) Name() string {
	return "heuristic"
}

func (h *HeuristicProvider) Translate(ctx context.Context, req TranslateRequest) (TranslateResult, error) {
	// 1. Check glossary override first
	if len(req.Glossary) > 0 {
		for term, trans := range req.Glossary {
			if strings.EqualFold(req.SourceText, term) {
				return TranslateResult{
					Translation: trans,
					Explanation: fmt.Sprintf("Applied exact glossary match for '%s'", term),
					Provider:    h.Name(),
				}, nil
			}
		}
	}

	lowerSrc := strings.ToLower(strings.TrimSpace(req.SourceText))
	locale := strings.ToLower(strings.Split(req.TargetLocale, "-")[0])

	// 2. Check dictionary
	variants, hasVariant := dictionary[lowerSrc][locale]
	if hasVariant {
		chosen := variants.standard
		charLimit := req.TargetCharLimit

		if charLimit > 0 {
			stdLen := utf8.RuneCountInString(variants.standard)
			shortLen := utf8.RuneCountInString(variants.short)
			compactLen := utf8.RuneCountInString(variants.compact)

			if stdLen <= charLimit {
				chosen = variants.standard
			} else if shortLen <= charLimit {
				chosen = variants.short
			} else if compactLen <= charLimit {
				chosen = variants.compact
			} else {
				// Trim to limit cleanly if even compact is too long
				runes := []rune(variants.compact)
				if len(runes) > charLimit && charLimit > 0 {
					chosen = string(runes[:charLimit])
				} else {
					chosen = variants.compact
				}
			}
		}

		return TranslateResult{
			Translation: chosen,
			Explanation: fmt.Sprintf("Synthesized dictionary variant (limit: %d chars)", req.TargetCharLimit),
			Provider:    h.Name(),
		}, nil
	}

	// 3. Fallback heuristic for generic phrases
	prefix := ""
	switch locale {
	case "de":
		prefix = "[DE] "
	case "fr":
		prefix = "[FR] "
	case "es":
		prefix = "[ES] "
	case "ja":
		prefix = "[JA] "
	default:
		prefix = fmt.Sprintf("[%s] ", strings.ToUpper(locale))
	}

	result := prefix + req.SourceText
	if req.TargetCharLimit > 0 {
		runes := []rune(result)
		if len(runes) > req.TargetCharLimit {
			// Trim without prefix if needed
			pureRunes := []rune(req.SourceText)
			if len(pureRunes) <= req.TargetCharLimit {
				result = req.SourceText
			} else if req.TargetCharLimit >= 4 {
				result = string(pureRunes[:req.TargetCharLimit-1]) + "…"
			} else {
				result = string(pureRunes[:req.TargetCharLimit])
			}
		}
	}

	return TranslateResult{
		Translation: result,
		Explanation: "Generated fallback translation with length constraint",
		Provider:    h.Name(),
	}, nil
}
