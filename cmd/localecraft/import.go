package main

import (
	"fmt"
	"path/filepath"

	"github.com/alexandrmotologa/localecraft/internal/config"
	"github.com/alexandrmotologa/localecraft/internal/parser"
	"github.com/alexandrmotologa/localecraft/internal/typography"
	"github.com/spf13/cobra"
)

var (
	importFormat string
	importLocale string
	importInput  string
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import external CSV or XLIFF translations into project catalogs with layout verification",
	Example: `  localecraft import --input ./translations_de.csv --locale de
  localecraft import --input ./messages_fr.xlf --locale fr`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if importInput == "" {
			return fmt.Errorf("--input file path is required")
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if len(cfg.Files) == 0 {
			return fmt.Errorf("no localization files configured")
		}

		targetLocale := importLocale
		if targetLocale == "" {
			targetLocale = cfg.TargetLocales[0]
		}

		fmtIn := importFormat
		if fmtIn == "" {
			fmtIn = parser.DetectFormat(importInput)
		}

		inP, err := parser.GetParser(fmtIn)
		if err != nil {
			return fmt.Errorf("get parser: %w", err)
		}

		importedCat, err := inP.ParseFile(importInput, targetLocale)
		if err != nil {
			return fmt.Errorf("read import file %s: %w", importInput, err)
		}

		fileCfg := cfg.Files[0]
		targetPath := fileCfg.ResolveOutputPath(targetLocale, cfg.WorkingDir)

		targetCat, err := parser.LoadFile(targetPath, fileCfg.Format, targetLocale)
		if err != nil {
			targetCat = parser.NewCatalog(fileCfg.Format, targetLocale, targetPath)
		}

		typoEngine, err := typography.NewEngine()
		if err != nil {
			return fmt.Errorf("init typography: %w", err)
		}

		overflowCount := 0
		appliedCount := 0

		fmt.Printf("\nImporting translations from %s into %s...\n", filepath.Base(importInput), targetPath)

		for _, k := range importedCat.KeyOrder {
			importedEntry := importedCat.Entries[k]
			if importedEntry.Value == "" {
				continue
			}

			constraint := cfg.GetKeyConstraint(k)

			// Check container layout
			if constraint.MaxWidthPX > 0 {
				m := typoEngine.Measure(typography.MeasureRequest{
					Text:       importedEntry.Value,
					Family:     constraint.FontFamily,
					SizePX:     constraint.FontSizePX,
					Weight:     constraint.FontWeight,
					PaddingXPX: 12,
					MaxWidthPX: constraint.MaxWidthPX,
				})
				if m.IsOverflow {
					overflowCount++
					fmt.Printf("  WARNING: Key '%s' overflows container (+%.1fpx)\n", k, m.OverflowPX)
				}
			}

			targetCat.Set(&parser.Entry{
				Key:   k,
				Value: importedEntry.Value,
			})
			appliedCount++
		}

		if err := parser.SaveFile(targetCat, targetPath); err != nil {
			return fmt.Errorf("save updated catalog %s: %w", targetPath, err)
		}

		fmt.Printf("\nSuccessfully imported %d keys (%d overflow warnings).\n\n", appliedCount, overflowCount)
		return nil
	},
}

func init() {
	importCmd.Flags().StringVarP(&importInput, "input", "i", "", "path to CSV or XLIFF file to import")
	importCmd.Flags().StringVarP(&importLocale, "locale", "l", "", "target locale for imported translations")
	importCmd.Flags().StringVarP(&importFormat, "format", "f", "", "input format (csv, xliff, auto-detected)")

	rootCmd.AddCommand(importCmd)
}
