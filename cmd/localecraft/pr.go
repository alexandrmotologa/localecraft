package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/alexandrmotologa/localecraft/internal/config"
	"github.com/spf13/cobra"
)

var (
	prBranch string
	prTitle  string
	prBody   string
	prRepo   string
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create a branch, commit updated localization files, and open a GitHub Pull Request",
	Example: `  localecraft pr --branch "i18n-layout-audit" --title "i18n: fix text overflows"
  localecraft pr --config ./localecraft.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if prBranch == "" {
			prBranch = fmt.Sprintf("localecraft/update-%d", time.Now().Unix())
		}
		if prTitle == "" {
			prTitle = "i18n: visual software localization and overflow prevention"
		}
		if prBody == "" {
			prBody = "Autonomous localization updates with verified container boundaries and ICU parity."
		}

		fmt.Printf("Creating branch: %s\n", prBranch)
		cmdBranch := exec.Command("git", "checkout", "-b", prBranch)
		cmdBranch.Dir = cfg.WorkingDir
		if out, err := cmdBranch.CombinedOutput(); err != nil {
			fmt.Printf("Notice: %s\n", string(out))
		}

		cmdAdd := exec.Command("git", "add", ".")
		cmdAdd.Dir = cfg.WorkingDir
		if out, err := cmdAdd.CombinedOutput(); err != nil {
			return fmt.Errorf("git add failed: %s (%w)", string(out), err)
		}

		cmdCommit := exec.Command("git", "commit", "-m", "chore(i18n): update localecraft translations")
		cmdCommit.Dir = cfg.WorkingDir
		if out, err := cmdCommit.CombinedOutput(); err != nil {
			fmt.Printf("Notice on commit: %s\n", string(out))
		}

		fmt.Printf("Pushing branch to origin...\n")
		cmdPush := exec.Command("git", "push", "-u", "origin", prBranch)
		cmdPush.Dir = cfg.WorkingDir
		if out, err := cmdPush.CombinedOutput(); err != nil {
			return fmt.Errorf("git push failed: %s (%w)", string(out), err)
		}

		fmt.Printf("Opening GitHub Pull Request via gh CLI...\n")
		ghArgs := []string{"pr", "create", "--title", prTitle, "--body", prBody, "--head", prBranch}
		if prRepo != "" {
			ghArgs = append(ghArgs, "--repo", prRepo)
		}

		cmdPR := exec.Command("gh", ghArgs...)
		cmdPR.Dir = cfg.WorkingDir
		prOut, err := cmdPR.CombinedOutput()
		if err != nil {
			return fmt.Errorf("gh pr create failed: %s (%w)", string(prOut), err)
		}

		fmt.Printf("\nPull Request successfully opened!\n%s\n\n", strings.TrimSpace(string(prOut)))
		return nil
	},
}

func init() {
	prCmd.Flags().StringVar(&prBranch, "branch", "", "git branch name")
	prCmd.Flags().StringVar(&prTitle, "title", "", "pull request title")
	prCmd.Flags().StringVar(&prBody, "body", "", "pull request description")
	prCmd.Flags().StringVar(&prRepo, "repo", "", "target GitHub repository (e.g. owner/repo)")

	rootCmd.AddCommand(prCmd)
}
