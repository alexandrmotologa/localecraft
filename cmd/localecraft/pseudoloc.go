package main

import (
	"fmt"
	"path/filepath"

	"github.com/alexandrmotologa/localecraft/internal/config"
	"github.com/alexandrmotologa/localecraft/internal/parser"
	"github.com/alexandrmotologa/localecraft/internal/pseudoloc"
	"github.com/spf13/cobra"
)

var (
	pseudoExpansion  int
	pseudoOutput     string
	pseudoNoBrackets bool
)

var pseudolocCmd = &cobra.Command{
	Use:   "pseudoloc",
	Short: "Generate pseudo-localized catalogs with accented characters and expansion padding",
	Example: `  localecraft pseudoloc --config ./localecraft.yaml
  localecraft pseudoloc --expansion 45 --output ./locales/pseudo.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if len(cfg.Files) == 0 {
			return fmt.Errorf("no files configured in %s", configPath)
		}

		fileCfg := cfg.Files[0]
		srcPath := fileCfg.Path
		if !filepath.IsAbs(srcPath) {
			srcPath = filepath.Join(cfg.WorkingDir, srcPath)
		}

		srcCat, err := parser.LoadFile(srcPath, fileCfg.Format, cfg.SourceLocale)
		if err != nil {
			return fmt.Errorf("load source catalog %s: %w", srcPath, err)
		}

		opts := pseudoloc.Options{
			ExpansionPCT: pseudoExpansion,
			Brackets:     !pseudoNoBrackets,
		}

		pseudoCat := pseudoloc.TransformCatalog(srcCat, opts)
		pseudoCat.Locale = "pseudo"

		outPath := pseudoOutput
		if outPath == "" {
			outPath = fileCfg.ResolveOutputPath("pseudo", cfg.WorkingDir)
		}

		if err := parser.SaveFile(pseudoCat, outPath); err != nil {
			return fmt.Errorf("write pseudo catalog %s: %w", outPath, err)
		}

		fmt.Printf("\nPseudo-localization catalog generated successfully!\n")
		fmt.Printf("Source:    %s (%s)\n", srcPath, cfg.SourceLocale)
		fmt.Printf("Output:    %s (pseudo)\n", outPath)
		fmt.Printf("Expansion: +%d%% length with accented characters\n", pseudoExpansion)
		fmt.Printf("Keys:      %d entries converted\n\n", len(pseudoCat.Entries))

		return nil
	},
}

func init() {
	pseudolocCmd.Flags().IntVar(&pseudoExpansion, "expansion", 35, "percentage to expand text length (e.g. 35 for +35%)")
	pseudolocCmd.Flags().StringVarP(&pseudoOutput, "output", "o", "", "destination path for pseudo catalog")
	pseudolocCmd.Flags().BoolVar(&pseudoNoBrackets, "no-brackets", false, "disable surrounding brackets [ ... ]")

	rootCmd.AddCommand(pseudolocCmd)
}
