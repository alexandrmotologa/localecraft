package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alexandrmotologa/localecraft/internal/agent"
	"github.com/alexandrmotologa/localecraft/internal/api"
	"github.com/alexandrmotologa/localecraft/internal/config"
	"github.com/alexandrmotologa/localecraft/internal/parser"
	"github.com/alexandrmotologa/localecraft/internal/store"
	"github.com/alexandrmotologa/localecraft/internal/typography"
	"github.com/alexandrmotologa/localecraft/web"
	"github.com/spf13/cobra"
)

var (
	servePort string
	serveDB   string
	serveOpen bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the visual localization studio web server",
	Example: `  localecraft serve --port 8080
  localecraft serve --config ./localecraft.yaml --open`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if len(cfg.Files) == 0 {
			return fmt.Errorf("no files defined in configuration")
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
		for _, loc := range cfg.TargetLocales {
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

		// Initialize SQLite Store
		dbPath := serveDB
		if dbPath == "" {
			dbPath = cfg.WorkingDir + "/localecraft.db"
		}
		db, err := store.Open(dbPath)
		if err != nil {
			return fmt.Errorf("initialize database %s: %w", dbPath, err)
		}
		defer db.Close()

		if err := db.SyncCatalogs(cfg, sourceCat, targetCats, typoEngine); err != nil {
			return fmt.Errorf("sync database with catalogs: %w", err)
		}

		// Initialize AI Provider
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

		// Get embedded frontend assets
		distFS, err := web.Dist()
		if err != nil {
			fmt.Printf("Notice: Embedded web assets not found: %v\n", err)
		}

		server := api.NewServer(cfg, db, typoEngine, optimizer, sourceCat, targetCats, distFS)
		addr := ":" + servePort

		fmt.Printf("\nLocaleCraft Visual Studio running at http://localhost:%s\n", servePort)
		fmt.Printf("Config: %s | Database: %s | AI Provider: %s\n", configPath, dbPath, aiProvider.Name())
		fmt.Printf("Press Ctrl+C to stop.\n\n")

		if serveOpen {
			go openBrowser("http://localhost:" + servePort)
		}

		return http.ListenAndServe(addr, server.Router())
	},
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func init() {
	serveCmd.Flags().StringVarP(&servePort, "port", "p", "8080", "HTTP server port")
	serveCmd.Flags().StringVar(&serveDB, "db", "", "SQLite database path (default: ./localecraft.db)")
	serveCmd.Flags().BoolVar(&serveOpen, "open", false, "automatically open browser")

	rootCmd.AddCommand(serveCmd)
}
