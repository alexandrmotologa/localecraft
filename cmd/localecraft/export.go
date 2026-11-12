package main

import (
	"fmt"
	"path/filepath"

	"github.com/alexandrmotologa/localecraft/internal/config"
	"github.com/alexandrmotologa/localecraft/internal/parser"
	"github.com/spf13/cobra"
)

var (
	exportFormat string
	exportLocale string
	exportOutput string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export localization catalogs to CSV, XLIFF, or JSON for external translators",
	Example: `  localecraft export --format csv --locale de --output ./translations_de.csv
  localecraft export --format xliff --locale fr --output ./messages_fr.xlf`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if len(cfg.Files) == 0 {
			return fmt.Errorf("no localization files configured in %s", configPath)
		}

		fileCfg := cfg.Files[0]
		targetLocale := exportLocale
		if targetLocale == "" {
			targetLocale = cfg.TargetLocales[0]
		}

		tgtPath := fileCfg.ResolveOutputPath(targetLocale, cfg.WorkingDir)
		cat, err := parser.LoadFile(tgtPath, fileCfg.Format, targetLocale)
		if err != nil {
			return fmt.Errorf("load target catalog %s: %w", tgtPath, err)
		}

		fmtOut := exportFormat
		if fmtOut == "" {
			fmtOut = "csv"
		}

		outP, err := parser.GetParser(fmtOut)
		if err != nil {
			return fmt.Errorf("get parser for %s: %w", fmtOut, err)
		}

		destPath := exportOutput
		if destPath == "" {
			ext := ".csv"
			if fmtOut == "xliff" || fmtOut == "xlf" {
				ext = ".xlf"
			}
			destPath = filepath.Join(cfg.WorkingDir, fmt.Sprintf("export_%s%s", targetLocale, ext))
		}

		cat.Format = fmtOut
		if err := outP.WriteFile(cat, destPath); err != nil {
			return fmt.Errorf("write exported catalog: %w", err)
		}

		fmt.Printf("\nExported %d keys to %s (format: %s, locale: %s)\n\n", len(cat.Entries), destPath, fmtOut, targetLocale)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "csv", "export format (csv, xliff, json_nested)")
	exportCmd.Flags().StringVarP(&exportLocale, "locale", "l", "", "target locale to export")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "destination export file path")

	rootCmd.AddCommand(exportCmd)
}
