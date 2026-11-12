package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexandrmotologa/localecraft/internal/config"
	"github.com/alexandrmotologa/localecraft/internal/scanner"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	scanDir          string
	scanUpdateConfig bool
	scanJSON         bool
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan frontend source files (TSX, JSX, Vue) to extract i18n keys and container width limits",
	Example: `  localecraft scan --dir ./src
  localecraft scan --dir ./web/src --update-config`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := scanDir
		if targetDir == "" {
			targetDir = "."
		}

		fmt.Printf("\nScanning codebase at %s for i18n translation keys and container boundaries...\n", targetDir)
		constraints, err := scanner.ScanDirectory(targetDir)
		if err != nil {
			return fmt.Errorf("scan directory: %w", err)
		}

		if scanJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(constraints)
		}

		fmt.Printf("Found %d localization keys in UI components:\n", len(constraints))
		fmt.Printf("----------------------------------------------------------------------------------------------------\n")
		fmt.Printf("%-32s %-12s %-10s %-10s %-30s\n", "KEY", "MAX WIDTH", "FONT SIZE", "WEIGHT", "LOCATION")
		fmt.Printf("----------------------------------------------------------------------------------------------------\n")

		for _, c := range constraints {
			widthStr := "unbounded"
			if c.MaxWidthPX > 0 {
				widthStr = fmt.Sprintf("%.0fpx", c.MaxWidthPX)
			}

			fmt.Printf("%-32s %-12s %-10.0f %-10d %-30s\n",
				truncateString(c.Key, 31),
				widthStr,
				c.FontSizePX,
				c.FontWeight,
				truncateString(fmt.Sprintf("%s:%d", filepath.Base(c.FilePath), c.LineNumber), 30),
			)
		}
		fmt.Printf("----------------------------------------------------------------------------------------------------\n")

		if scanUpdateConfig {
			cfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			added := scanner.UpdateConfigWithInferred(cfg, constraints)
			yamlBytes, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("marshal updated config: %w", err)
			}

			if err := os.WriteFile(configPath, yamlBytes, 0644); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Printf("Successfully updated %s with %d inferred UI constraints.\n\n", configPath, added)
		} else {
			fmt.Printf("Tip: Run with --update-config to persist these constraints into %s\n\n", configPath)
		}

		return nil
	},
}

func init() {
	scanCmd.Flags().StringVarP(&scanDir, "dir", "d", ".", "root directory to scan for components")
	scanCmd.Flags().BoolVarP(&scanUpdateConfig, "update-config", "u", false, "update localecraft.yaml with extracted constraints")
	scanCmd.Flags().BoolVar(&scanJSON, "json", false, "output detected constraints as JSON")

	rootCmd.AddCommand(scanCmd)
}
