package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alexandrmotologa/localecraft/internal/typography"
	"github.com/spf13/cobra"
)

var (
	measureText          string
	measureFamily        string
	measureSize          float64
	measureWeight        int
	measureLetterSpacing float64
	measurePadding       float64
	measureMaxWidth      float64
	measureJSON          bool
)

var measureCmd = &cobra.Command{
	Use:   "measure",
	Short: "Calculate exact rendered pixel dimensions of a string using font metrics",
	Example: `  localecraft measure --text "Speichern und anwenden" --max-width 120
  localecraft measure --text "Delete account permanently" --family "Inter" --size 16 --weight 600`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if measureText == "" {
			return fmt.Errorf("--text flag is required")
		}

		typoEngine, err := typography.NewEngine()
		if err != nil {
			return fmt.Errorf("initialize typography engine: %w", err)
		}

		res := typoEngine.Measure(typography.MeasureRequest{
			Text:            measureText,
			Family:          measureFamily,
			SizePX:          measureSize,
			Weight:          measureWeight,
			LetterSpacingPX: measureLetterSpacing,
			PaddingXPX:      measurePadding,
			MaxWidthPX:      measureMaxWidth,
			MaxLines:        1,
		})

		if measureJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(res)
		}

		fmt.Printf("\nTypography Measurement for: \"%s\"\n", res.Text)
		fmt.Printf("--------------------------------------------------\n")
		fmt.Printf("Font:              %s %.1fpx (weight %d)\n", measureFamily, measureSize, measureWeight)
		fmt.Printf("Rendered Width:    %.2f px\n", res.RenderedWidthPX)
		fmt.Printf("Glyph Count:       %d\n", res.GlyphCount)

		if measureMaxWidth > 0 {
			fmt.Printf("Max Container:     %.2f px (Available: %.2f px, Padding: 2x%.1fpx)\n", measureMaxWidth, res.AvailableWidthPX, measurePadding)
			if res.IsOverflow {
				fmt.Printf("Status:            OVERFLOW (+%.2f px / +%.1f%%)\n", res.OverflowPX, res.OverflowPCT)
				fmt.Printf("Compression Ratio: %.3f\n", res.CompressionRatio)
				fmt.Printf("Target Char Limit: %d characters\n", res.TargetCharLimit)
			} else {
				fmt.Printf("Status:            PASSED (Fits inside container)\n")
			}
		}
		fmt.Printf("--------------------------------------------------\n\n")

		return nil
	},
}

func init() {
	measureCmd.Flags().StringVarP(&measureText, "text", "t", "", "string to measure")
	measureCmd.Flags().StringVar(&measureFamily, "family", "Inter", "font family (Inter, Roboto, default, mono)")
	measureCmd.Flags().Float64Var(&measureSize, "size", 14.0, "font size in pixels")
	measureCmd.Flags().IntVar(&measureWeight, "weight", 500, "font weight (400, 500, 600, 700)")
	measureCmd.Flags().Float64Var(&measureLetterSpacing, "letter-spacing", 0.0, "letter spacing in pixels")
	measureCmd.Flags().Float64Var(&measurePadding, "padding", 12.0, "horizontal container padding in pixels")
	measureCmd.Flags().Float64Var(&measureMaxWidth, "max-width", 0.0, "maximum container width in pixels")
	measureCmd.Flags().BoolVar(&measureJSON, "json", false, "output result as JSON")

	rootCmd.AddCommand(measureCmd)
}
