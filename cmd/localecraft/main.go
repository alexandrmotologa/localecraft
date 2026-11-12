package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configPath string
	version    = "1.0.0"

	rootCmd = &cobra.Command{
		Use:   "localecraft",
		Short: "Visual software localization engine and layout overflow sentinel",
		Long: `localecraft is a visual software localization engine and layout overflow sentinel.
It combines mathematical typography measurement (glyph advances, font metrics)
with autonomous language compression loops and strict ICU MessageFormat validation.`,
		Version: version,
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "localecraft.yaml", "path to localecraft.yaml configuration file")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
