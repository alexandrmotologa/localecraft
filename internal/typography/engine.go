package typography

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gomedium"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// FontKey identifies a font family and weight combination.
type FontKey struct {
	Family string
	Weight int
}

// MeasureRequest defines parameters for a string bounding box measurement.
type MeasureRequest struct {
	Text            string  `json:"text"`
	Family          string  `json:"family"`
	SizePX          float64 `json:"size_px"`
	Weight          int     `json:"weight"`
	LetterSpacingPX float64 `json:"letter_spacing_px"`
	PaddingXPX      float64 `json:"padding_x_px"`
	MaxWidthPX      float64 `json:"max_width_px"`
	MaxLines        int     `json:"max_lines"`
}

// MeasureResult contains computed pixel dimensions and constraint status.
type MeasureResult struct {
	Text             string  `json:"text"`
	RenderedWidthPX  float64 `json:"rendered_width_px"`
	RenderedHeightPX float64 `json:"rendered_height_px"`
	AvailableWidthPX float64 `json:"available_width_px"`
	IsOverflow       bool    `json:"is_overflow"`
	OverflowPX       float64 `json:"overflow_px"`
	OverflowPCT      float64 `json:"overflow_pct"`
	TargetCharLimit  int     `json:"target_char_limit"`
	GlyphCount       int     `json:"glyph_count"`
	LinesUsed        int     `json:"lines_used"`
	CompressionRatio float64 `json:"compression_ratio"`
}

// Engine manages font parsing, face caching, and pixel-precise layout measurements.
type Engine struct {
	mu     sync.RWMutex
	fonts  map[FontKey]*opentype.Font
	faces  map[string]font.Face
	dflt   *opentype.Font
	bold   *opentype.Font
	medium *opentype.Font
	mono   *opentype.Font
}

// NewEngine creates and initializes a typography measurement engine with bundled fonts.
func NewEngine() (*Engine, error) {
	e := &Engine{
		fonts: make(map[FontKey]*opentype.Font),
		faces: make(map[string]font.Face),
	}

	regularFont, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, fmt.Errorf("parse bundled regular font: %w", err)
	}
	boldFont, err := opentype.Parse(gobold.TTF)
	if err != nil {
		return nil, fmt.Errorf("parse bundled bold font: %w", err)
	}
	mediumFont, err := opentype.Parse(gomedium.TTF)
	if err != nil {
		return nil, fmt.Errorf("parse bundled medium font: %w", err)
	}
	monoFont, err := opentype.Parse(gomono.TTF)
	if err != nil {
		return nil, fmt.Errorf("parse bundled mono font: %w", err)
	}

	e.dflt = regularFont
	e.bold = boldFont
	e.medium = mediumFont
	e.mono = monoFont

	// Register built-in default aliases
	e.RegisterFont("default", 400, regularFont)
	e.RegisterFont("default", 500, mediumFont)
	e.RegisterFont("default", 600, boldFont)
	e.RegisterFont("default", 700, boldFont)

	e.RegisterFont("inter", 400, regularFont)
	e.RegisterFont("inter", 500, mediumFont)
	e.RegisterFont("inter", 600, boldFont)
	e.RegisterFont("inter", 700, boldFont)

	e.RegisterFont("roboto", 400, regularFont)
	e.RegisterFont("roboto", 700, boldFont)

	e.RegisterFont("mono", 400, monoFont)
	e.RegisterFont("monospace", 400, monoFont)

	return e, nil
}

// RegisterFont associates a parsed font with a family name and weight.
func (e *Engine) RegisterFont(family string, weight int, f *opentype.Font) {
	e.mu.Lock()
	defer e.mu.Unlock()
	key := FontKey{
		Family: strings.ToLower(strings.TrimSpace(family)),
		Weight: normalizeWeight(weight),
	}
	e.fonts[key] = f
}

// LoadFontFile reads a TTF or OTF file from disk and adds it to the font catalog.
func (e *Engine) LoadFontFile(path string, family string, weight int) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read font file: %w", err)
	}

	parsed, err := opentype.Parse(data)
	if err != nil {
		return fmt.Errorf("parse font file %s: %w", filepath.Base(path), err)
	}

	e.RegisterFont(family, weight, parsed)
	return nil
}

func normalizeWeight(w int) int {
	if w <= 0 {
		return 400
	}
	if w < 450 {
		return 400
	}
	if w < 600 {
		return 500
	}
	if w < 650 {
		return 600
	}
	return 700
}

func (e *Engine) getFace(family string, sizePX float64, weight int) (font.Face, error) {
	normWeight := normalizeWeight(weight)
	famKey := strings.ToLower(strings.TrimSpace(family))
	if famKey == "" {
		famKey = "default"
	}

	cacheKey := fmt.Sprintf("%s:%d:%.2f", famKey, normWeight, sizePX)

	e.mu.RLock()
	if face, exists := e.faces[cacheKey]; exists {
		e.mu.RUnlock()
		return face, nil
	}
	e.mu.RUnlock()

	e.mu.Lock()
	defer e.mu.Unlock()

	// Double check after write lock
	if face, exists := e.faces[cacheKey]; exists {
		return face, nil
	}

	fontObj, exists := e.fonts[FontKey{Family: famKey, Weight: normWeight}]
	if !exists {
		// Try generic fallback
		switch {
		case normWeight >= 650:
			fontObj = e.bold
		case normWeight >= 500:
			fontObj = e.medium
		default:
			fontObj = e.dflt
		}
	}

	face, err := opentype.NewFace(fontObj, &opentype.FaceOptions{
		Size:    sizePX,
		DPI:     72, // 1 pt = 1 px at 72 DPI
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("create font face: %w", err)
	}

	e.faces[cacheKey] = face
	return face, nil
}

// Measure calculates the rendered pixel width and checks against container limits.
func (e *Engine) Measure(req MeasureRequest) MeasureResult {
	runeCount := utf8.RuneCountInString(req.Text)
	if runeCount == 0 {
		avail := req.MaxWidthPX - (req.PaddingXPX * 2)
		if avail < 0 {
			avail = 0
		}
		return MeasureResult{
			Text:             req.Text,
			RenderedWidthPX:  0,
			RenderedHeightPX: req.SizePX,
			AvailableWidthPX: avail,
			IsOverflow:       false,
			OverflowPX:       0,
			OverflowPCT:      0,
			TargetCharLimit:  0,
			GlyphCount:       0,
			LinesUsed:        1,
			CompressionRatio: 1.0,
		}
	}

	if req.SizePX <= 0 {
		req.SizePX = 14
	}

	face, err := e.getFace(req.Family, req.SizePX, req.Weight)
	if err != nil {
		// Fallback to basic estimation if font face fails
		estimatedWidth := float64(runeCount) * req.SizePX * 0.58
		avail := req.MaxWidthPX - (req.PaddingXPX * 2)
		isOver := req.MaxWidthPX > 0 && estimatedWidth > avail
		overPX := 0.0
		overPCT := 0.0
		ratio := 1.0
		targetChars := runeCount
		if isOver && avail > 0 {
			overPX = estimatedWidth - avail
			overPCT = (overPX / avail) * 100
			ratio = avail / estimatedWidth
			targetChars = int(math.Floor(float64(runeCount) * ratio * 0.95))
			if targetChars < 3 {
				targetChars = 3
			}
		}
		return MeasureResult{
			Text:             req.Text,
			RenderedWidthPX:  math.Round(estimatedWidth*100) / 100,
			RenderedHeightPX: req.SizePX,
			AvailableWidthPX: avail,
			IsOverflow:       isOver,
			OverflowPX:       math.Round(overPX*100) / 100,
			OverflowPCT:      math.Round(overPCT*10) / 10,
			TargetCharLimit:  targetChars,
			GlyphCount:       runeCount,
			LinesUsed:        1,
			CompressionRatio: math.Round(ratio*1000) / 1000,
		}
	}

	lines := strings.Split(req.Text, "\n")
	var maxWidth fixed.Int26_6
	linesUsed := len(lines)

	for _, line := range lines {
		w := font.MeasureString(face, line)
		if w > maxWidth {
			maxWidth = w
		}
	}

	// 1 unit in 26.6 fixed-point = 1/64 pixel
	baseWidthPX := float64(maxWidth) / 64.0

	// Add letter spacing: (runeCount - 1) * letterSpacingPX for the longest line
	if req.LetterSpacingPX > 0 && runeCount > 1 {
		baseWidthPX += float64(runeCount-1) * req.LetterSpacingPX
	}

	availableWidthPX := req.MaxWidthPX
	if req.PaddingXPX > 0 {
		availableWidthPX -= (req.PaddingXPX * 2)
	}
	if availableWidthPX < 0 {
		availableWidthPX = 0
	}

	isOverflow := false
	overflowPX := 0.0
	overflowPCT := 0.0
	compressionRatio := 1.0
	targetCharLimit := runeCount

	if req.MaxWidthPX > 0 && baseWidthPX > availableWidthPX {
		isOverflow = true
		overflowPX = baseWidthPX - availableWidthPX
		if availableWidthPX > 0 {
			overflowPCT = (overflowPX / availableWidthPX) * 100
			compressionRatio = availableWidthPX / baseWidthPX
			targetCharLimit = int(math.Floor(float64(runeCount) * compressionRatio * 0.96))
			if targetCharLimit < 3 {
				targetCharLimit = 3
			}
		}
	}

	lineHeight := req.SizePX * 1.35
	renderedHeight := lineHeight * float64(linesUsed)

	return MeasureResult{
		Text:             req.Text,
		RenderedWidthPX:  math.Round(baseWidthPX*100) / 100,
		RenderedHeightPX: math.Round(renderedHeight*100) / 100,
		AvailableWidthPX: math.Round(availableWidthPX*100) / 100,
		IsOverflow:       isOverflow,
		OverflowPX:       math.Round(overflowPX*100) / 100,
		OverflowPCT:      math.Round(overflowPCT*10) / 10,
		TargetCharLimit:  targetCharLimit,
		GlyphCount:       runeCount,
		LinesUsed:        linesUsed,
		CompressionRatio: math.Round(compressionRatio*1000) / 1000,
	}
}
