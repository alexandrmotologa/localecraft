package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initPath string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a starter localecraft configuration and example catalogs",
	Example: `  localecraft init
  localecraft init --path ./my-project`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := initPath
		if targetDir == "" {
			targetDir = "."
		}

		cfgFile := filepath.Join(targetDir, "localecraft.yaml")
		if _, err := os.Stat(cfgFile); err == nil {
			return fmt.Errorf("%s already exists", cfgFile)
		}

		yamlContent := `# localecraft configuration manifest
version: "1"
source_locale: "en"
target_locales:
  - "de"
  - "fr"
  - "es"
  - "ja"

files:
  - path: "locales/en.json"
    format: "json_nested"
    output_pattern: "locales/{{.Locale}}.json"

ui_constraints:
  font_defaults:
    family: "Inter"
    size_px: 14
    weight: 500
    letter_spacing_px: 0.0
    padding_x_px: 12

  keys:
    "nav.checkout_btn":
      max_width_px: 130
      max_lines: 1
      context: "Primary checkout CTA in navigation bar."

    "settings.delete_account_confirm":
      max_width_px: 220
      max_lines: 1
      context: "Destructive confirmation button in red danger zone."

    "inbox.task_counter":
      format: "icu_messageformat"
      context: "Task count counter on top toolbar."

glossary:
  - term: "Workspace"
    translations:
      de: "Arbeitsbereich"
      fr: "Espace de travail"
      ja: "ワークスペース"

ai:
  provider: "heuristic" # "gemini", "openai", "heuristic"
`

		localesDir := filepath.Join(targetDir, "locales")
		if err := os.MkdirAll(localesDir, 0755); err != nil {
			return fmt.Errorf("create locales dir: %w", err)
		}

		enJSON := `{
  "nav": {
    "checkout_btn": "Checkout"
  },
  "settings": {
    "delete_account_confirm": "Delete Account"
  },
  "inbox": {
    "task_counter": "{count, plural, =0{No tasks remaining} one{# task remaining} other{# tasks remaining}}"
  }
}
`

		if err := os.WriteFile(cfgFile, []byte(yamlContent), 0644); err != nil {
			return fmt.Errorf("write %s: %w", cfgFile, err)
		}

		enPath := filepath.Join(localesDir, "en.json")
		if err := os.WriteFile(enPath, []byte(enJSON), 0644); err != nil {
			return fmt.Errorf("write %s: %w", enPath, err)
		}

		fmt.Printf("\nInitialized localecraft workspace successfully!\n")
		fmt.Printf("Created: %s\n", cfgFile)
		fmt.Printf("Created: %s\n\n", enPath)
		fmt.Printf("Next steps:\n")
		fmt.Printf("  1. Audit current catalogs: localecraft audit\n")
		fmt.Printf("  2. Autonomously translate: localecraft translate\n")
		fmt.Printf("  3. Launch Visual Studio:   localecraft serve --open\n\n")

		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initPath, "path", ".", "target directory to initialize")
	rootCmd.AddCommand(initCmd)
}
