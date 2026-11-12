package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexandrmotologa/localecraft/internal/config"
	"github.com/alexandrmotologa/localecraft/internal/icu"
	"github.com/alexandrmotologa/localecraft/internal/parser"
	"github.com/alexandrmotologa/localecraft/internal/typography"
	"github.com/spf13/cobra"
)

var (
	auditStrict bool
	auditJSON   bool
)

type AuditKeyResult struct {
	Key             string  `json:"key"`
	Locale          string  `json:"locale"`
	SourceValue     string  `json:"source_value"`
	TargetValue     string  `json:"target_value"`
	Status          string  `json:"status"` // "PASSED", "OVERFLOW", "MISSING", "ICU_ERROR"
	RenderedWidthPX float64 `json:"rendered_width_px"`
	MaxWidthPX      float64 `json:"max_width_px"`
	OverflowPX      float64 `json:"overflow_px"`
	ErrorDetail     string  `json:"error_detail,omitempty"`
}

type AuditReport struct {
	TotalEvaluated int              `json:"total_evaluated"`
	PassedCount    int              `json:"passed_count"`
	OverflowCount  int              `json:"overflow_count"`
	MissingCount   int              `json:"missing_count"`
	ICUErrorCount  int              `json:"icu_error_count"`
	Results        []AuditKeyResult `json:"results"`
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit localization catalogs for layout overflows and ICU MessageFormat errors",
	Example: `  localecraft audit --config ./localecraft.yaml
  localecraft audit --config ./localecraft.yaml --strict`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if len(cfg.Files) == 0 {
			return fmt.Errorf("no localization files configured in %s", configPath)
		}

		fileCfg := cfg.Files[0]
		srcPath := fileCfg.Path
		if !filepath.IsAbs(srcPath) {
			srcPath = filepath.Join(cfg.WorkingDir, srcPath)
		}

		sourceCat, err := parser.LoadFile(srcPath, fileCfg.Format, cfg.SourceLocale)
		if err != nil {
			return fmt.Errorf("load source catalog %s: %w", srcPath, err)
		}

		targetCats := make(map[string]*parser.Catalog)
		for _, loc := range cfg.TargetLocales {
			tgtPath := fileCfg.ResolveOutputPath(loc, cfg.WorkingDir)
			cat, err := parser.LoadFile(tgtPath, fileCfg.Format, loc)
			if err != nil {
				// Initialize empty catalog if file doesn't exist yet
				cat = parser.NewCatalog(fileCfg.Format, loc, tgtPath)
			}
			targetCats[loc] = cat
		}

		typoEngine, err := typography.NewEngine()
		if err != nil {
			return fmt.Errorf("initialize typography: %w", err)
		}

		report := AuditReport{
			Results: make([]AuditKeyResult, 0),
		}

		for _, key := range sourceCat.KeyOrder {
			srcEntry, ok := sourceCat.Entries[key]
			if !ok {
				continue
			}

			constraint := cfg.GetKeyConstraint(key)
			isICU := constraint.Format == "icu_messageformat" || strings.Contains(srcEntry.Value, "{")

			for _, loc := range cfg.TargetLocales {
				report.TotalEvaluated++
				cat := targetCats[loc]
				tgtEntry, exists := cat.Entries[key]

				if !exists || strings.TrimSpace(tgtEntry.Value) == "" {
					report.MissingCount++
					report.Results = append(report.Results, AuditKeyResult{
						Key:         key,
						Locale:      loc,
						SourceValue: srcEntry.Value,
						Status:      "MISSING",
						MaxWidthPX:  constraint.MaxWidthPX,
					})
					continue
				}

				status := "PASSED"
				var errorDetail string

				// 1. ICU Validation
				if isICU {
					valRes := icu.Validate(srcEntry.Value, tgtEntry.Value, loc)
					if !valRes.Valid {
						status = "ICU_ERROR"
						errorDetail = strings.Join(valRes.Errors, "; ")
						report.ICUErrorCount++
					}
				}

				// 2. Typography Measurement
				measure := typoEngine.Measure(typography.MeasureRequest{
					Text:       tgtEntry.Value,
					Family:     constraint.FontFamily,
					SizePX:     constraint.FontSizePX,
					Weight:     constraint.FontWeight,
					PaddingXPX: 12,
					MaxWidthPX: constraint.MaxWidthPX,
					MaxLines:   constraint.MaxLines,
				})

				if status != "ICU_ERROR" {
					if measure.IsOverflow {
						status = "OVERFLOW"
						report.OverflowCount++
					} else {
						report.PassedCount++
					}
				}

				report.Results = append(report.Results, AuditKeyResult{
					Key:             key,
					Locale:          loc,
					SourceValue:     srcEntry.Value,
					TargetValue:     tgtEntry.Value,
					Status:          status,
					RenderedWidthPX: measure.RenderedWidthPX,
					MaxWidthPX:      constraint.MaxWidthPX,
					OverflowPX:      measure.OverflowPX,
					ErrorDetail:     errorDetail,
				})
			}
		}

		if auditJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(report)
		}

		// Print terminal output
		fmt.Printf("\nlocalecraft Localization & Layout Audit\n")
		fmt.Printf("Source Locale: %s | Target Locales: %s\n", cfg.SourceLocale, strings.Join(cfg.TargetLocales, ", "))
		fmt.Printf("----------------------------------------------------------------------------------------------------\n")
		fmt.Printf("%-28s %-6s %-12s %-10s %-10s %-12s\n", "KEY", "LOCALE", "STATUS", "WIDTH", "LIMIT", "OVERFLOW")
		fmt.Printf("----------------------------------------------------------------------------------------------------\n")

		for _, r := range report.Results {
			limitStr := fmt.Sprintf("%.0fpx", r.MaxWidthPX)
			if r.MaxWidthPX <= 0 {
				limitStr = "unbounded"
			}

			overflowStr := "-"
			if r.OverflowPX > 0 {
				overflowStr = fmt.Sprintf("+%.1fpx", r.OverflowPX)
			}

			fmt.Printf("%-28s %-6s %-12s %-10.1f %-10s %-12s\n",
				truncateString(r.Key, 27),
				r.Locale,
				r.Status,
				r.RenderedWidthPX,
				limitStr,
				overflowStr,
			)

			if r.ErrorDetail != "" {
				fmt.Printf("  └─ %s\n", r.ErrorDetail)
			}
		}

		fmt.Printf("----------------------------------------------------------------------------------------------------\n")
		fmt.Printf("Summary: Total: %d | Passed: %d | Overflows: %d | Missing: %d | ICU Errors: %d\n\n",
			report.TotalEvaluated, report.PassedCount, report.OverflowCount, report.MissingCount, report.ICUErrorCount)

		if auditStrict && (report.OverflowCount > 0 || report.ICUErrorCount > 0 || report.MissingCount > 0) {
			return fmt.Errorf("audit failed in strict mode with %d overflows, %d missing keys, and %d ICU errors",
				report.OverflowCount, report.MissingCount, report.ICUErrorCount)
		}

		return nil
	},
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

func init() {
	auditCmd.Flags().BoolVarP(&auditStrict, "strict", "s", false, "exit with status code 1 if any overflows or ICU errors are found")
	auditCmd.Flags().BoolVar(&auditJSON, "json", false, "output report as JSON")
	rootCmd.AddCommand(auditCmd)
}
