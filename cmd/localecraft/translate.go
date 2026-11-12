package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alexandrmotologa/localecraft/internal/agent"
	"github.com/alexandrmotologa/localecraft/internal/config"
	"github.com/alexandrmotologa/localecraft/internal/parser"
	"github.com/alexandrmotologa/localecraft/internal/typography"
	"github.com/spf13/cobra"
)

var (
	transProvider string
	transModel    string
	transAPIKey   string
	transLocale   string
	transKey      string
	transNoSave   bool
)

var translateCmd = &cobra.Command{
	Use:   "translate",
	Short: "Autonomously translate and rephrase localization catalogs under layout constraints",
	Example: `  localecraft translate --config ./localecraft.yaml
  localecraft translate --config ./localecraft.yaml --provider gemini --locale de`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		// Apply CLI overrides
		if transProvider != "" {
			cfg.AI.Provider = transProvider
		}
		if transModel != "" {
			cfg.AI.Model = transModel
		}
		if transAPIKey != "" {
			cfg.AI.APIKey = transAPIKey
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
			return fmt.Errorf("load source catalog: %w", err)
		}

		targetCats := make(map[string]*parser.Catalog)
		localesToTranslate := cfg.TargetLocales
		if transLocale != "" {
			localesToTranslate = []string{transLocale}
		}

		for _, loc := range localesToTranslate {
			tgtPath := fileCfg.ResolveOutputPath(loc, cfg.WorkingDir)
			cat, err := parser.LoadFile(tgtPath, fileCfg.Format, loc)
			if err != nil {
				cat = parser.NewCatalog(fileCfg.Format, loc, tgtPath)
			}
			targetCats[loc] = cat
		}

		typoEngine, err := typography.NewEngine()
		if err != nil {
			return fmt.Errorf("initialize typography: %w", err)
		}

		// Instantiate AI provider
		var aiProvider agent.Provider
		switch strings.ToLower(cfg.AI.Provider) {
		case "gemini":
			aiProvider = agent.NewGeminiProvider(cfg.AI.APIKey, cfg.AI.Model)
		case "openai":
			aiProvider = agent.NewOpenAIProvider(cfg.AI.APIKey, cfg.AI.BaseURL, cfg.AI.Model)
		default:
			aiProvider = agent.NewHeuristicProvider()
		}

		optimizer := agent.NewLayoutOptimizer(aiProvider, typoEngine)
		ctx := context.Background()

		fmt.Printf("\nlocalecraft Autonomous Translation & Layout Optimization\n")
		fmt.Printf("Provider: %s | Source: %s\n", aiProvider.Name(), cfg.SourceLocale)
		fmt.Printf("----------------------------------------------------------------------------------------------------\n")

		updatedCount := 0

		for _, key := range sourceCat.KeyOrder {
			if transKey != "" && key != transKey {
				continue
			}

			srcEntry := sourceCat.Entries[key]
			constraint := cfg.GetKeyConstraint(key)

			for _, loc := range localesToTranslate {
				cat := targetCats[loc]
				existing, exists := cat.Entries[key]

				// Check if translation is needed (missing or overflowing)
				needsWork := !exists || strings.TrimSpace(existing.Value) == ""
				if !needsWork && constraint.MaxWidthPX > 0 {
					m := typoEngine.Measure(typography.MeasureRequest{
						Text:       existing.Value,
						Family:     constraint.FontFamily,
						SizePX:     constraint.FontSizePX,
						Weight:     constraint.FontWeight,
						PaddingXPX: 12,
						MaxWidthPX: constraint.MaxWidthPX,
					})
					if m.IsOverflow {
						needsWork = true
					}
				}

				if !needsWork {
					continue
				}

				// Build glossary for target locale
				glossary := make(map[string]string)
				for _, gt := range cfg.Glossary {
					if v, ok := gt.Translations[loc]; ok {
						glossary[gt.Term] = v
					}
				}

				res, err := optimizer.Optimize(ctx, agent.OptimizeRequest{
					Key:           key,
					SourceText:    srcEntry.Value,
					SourceLocale:  cfg.SourceLocale,
					TargetLocale:  loc,
					Context:       constraint.Context,
					Constraint:    constraint,
					Glossary:      glossary,
					MaxIterations: 4,
				})
				if err != nil {
					fmt.Printf("ERROR [%s/%s]: %v\n", key, loc, err)
					continue
				}

				cat.Set(&parser.Entry{
					Key:   key,
					Value: res.FinalTranslation,
				})
				updatedCount++

				iterNotice := ""
				if res.IterationsUsed > 1 {
					iterNotice = fmt.Sprintf("(compressed in %d iterations)", res.IterationsUsed)
				}

				fmt.Printf("[%s] %-24s -> \"%s\" (%.1fpx / %.0fpx) %s\n",
					loc,
					truncateString(key, 24),
					res.FinalTranslation,
					res.FinalMeasure.RenderedWidthPX,
					constraint.MaxWidthPX,
					iterNotice,
				)
			}
		}

		fmt.Printf("----------------------------------------------------------------------------------------------------\n")
		fmt.Printf("Optimization complete. Processed %d translations.\n", updatedCount)

		// Save updated catalogs to disk
		if !transNoSave && updatedCount > 0 {
			for loc, cat := range targetCats {
				tgtPath := fileCfg.ResolveOutputPath(loc, cfg.WorkingDir)
				if err := parser.SaveFile(cat, tgtPath); err != nil {
					return fmt.Errorf("save catalog %s: %w", tgtPath, err)
				}
				fmt.Printf("Saved updated locale catalog: %s\n", tgtPath)
			}
		}

		fmt.Printf("\n")
		return nil
	},
}

func init() {
	translateCmd.Flags().StringVar(&transProvider, "provider", "", "AI translation provider (gemini, openai, heuristic)")
	translateCmd.Flags().StringVar(&transModel, "model", "", "model name")
	translateCmd.Flags().StringVar(&transAPIKey, "api-key", "", "API key override")
	translateCmd.Flags().StringVarP(&transLocale, "locale", "l", "", "target locale to translate (default: all)")
	translateCmd.Flags().StringVarP(&transKey, "key", "k", "", "translate only a specific key")
	translateCmd.Flags().BoolVar(&transNoSave, "no-save", false, "dry run without saving files")

	rootCmd.AddCommand(translateCmd)
}
