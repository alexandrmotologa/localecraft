# localecraft

Context-aware visual software localization engine and layout overflow sentinel. Calculates exact font bounding boxes, validates ICU MessageFormat syntax, and rephrases expanded text within container bounds.

## Problem

Software localization frequently produces visual regressions:
- Text expansion: German and French translations expand by 30% to 45% over English source strings, clipping text or breaking fixed containers.
- Syntax corruption: Automated machine translation often corrupts ICU MessageFormat brackets or drops placeholder variables, causing client runtime crashes.
- Missing container context: Strings translated without bounding box metadata produce labels that do not fit their visual target.

## Solution

`localecraft` couples semantic localization with physical typography calculations:
- Calculates rendered pixel widths directly from font metrics (OpenType and TrueType) at build time.
- Compares text length against container limits defined in `localecraft.yaml`.
- Iteratively rephrases overflowing translations into concise synonyms under strict character bounds.
- Validates variable parity and CLDR plural category compliance for ICU MessageFormat.
- Includes a terminal auditing tool for CI/CD gates and an embedded React 19 visual studio.

## Architecture

```
Source Files (.json, .po, .arb)
              │
              ▼
   localecraft CLI / Engine
              │
      ┌───────┴───────┐
      ▼               ▼
Layout Constraints   ICU Syntax Gate
(Font & Container)   (Brackets & Parity)
      │               │
      ▼               │
Mathematical Typography Engine
(Glyph advance calculations)
      │
      ├───────────────────────────────┐
      │ (Fits)                        │ (Overflows)
      ▼                               ▼
Approved Locale Files          Autonomous Rewording Loop
                               (Character budget feedback)
                                      │
                                      ▼
                               React 19 Studio & PR
```

## Quickstart

### 1. Installation

Ensure Go 1.24+ is installed.

```bash
git clone https://github.com/alexandrmotologa/localecraft.git
cd localecraft
go build -o localecraft ./cmd/localecraft
```

### 2. Measure a String

```bash
./localecraft measure --text "Speichern und anwenden" --font "Inter" --size 14 --weight 500 --max-width 120
```

Output:
```text
Typography Measurement for: "Speichern und anwenden"
--------------------------------------------------
Font:              Inter 14.0px (weight 500)
Rendered Width:    164.00 px
Glyph Count:       22
Max Container:     120.00 px (Available: 96.00 px, Padding: 2x12.0px)
Status:            OVERFLOW (+68.00 px / +70.8%)
Compression Ratio: 0.585
Target Char Limit: 12 characters
--------------------------------------------------
```

### 3. Run an Audit

```bash
./localecraft audit --config ./examples/localecraft.yaml
```

To run in CI and return exit code 1 on errors:
```bash
./localecraft audit --config ./examples/localecraft.yaml --strict
```

### 4. Autonomous Translation and Shortening

```bash
./localecraft translate --config ./examples/localecraft.yaml
```

The optimizer detects overflow on long strings and compresses them to fit within container limits.

### 5. Launch Embedded Visual Studio

```bash
./localecraft serve --config ./examples/localecraft.yaml --port 8080
```

Open `http://localhost:8080` in your browser.

## Configuration

Place `localecraft.yaml` in your project root:

```yaml
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
    padding_x_px: 12

  keys:
    "nav.checkout_btn":
      max_width_px: 120
      max_lines: 1
      context: "Primary checkout CTA in top header bar."

    "settings.delete_account_confirm":
      max_width_px: 220
      max_lines: 1
      context: "Destructive confirmation button in red danger zone."

glossary:
  - term: "Workspace"
    translations:
      de: "Arbeitsbereich"
      fr: "Espace de travail"
      ja: "ワークスペース"

ai:
  provider: "heuristic" # "gemini", "openai", "heuristic"
```

## Supported Localization Formats

- **Nested and Flat JSON**: dot-separated key hierarchy (`json_nested`)
- **GNU Gettext**: PO and POT files with context and plurals (`gettext_po`)
- **Flutter ARB**: Application Resource Bundle files (`flutter_arb`)

## Testing

Run all unit and integration tests:

```bash
go test -v ./...
```

Run web frontend build:

```bash
cd web && npm test || npm run build
```

## Documentation

- [Architecture Overview](docs/ARCHITECTURE.md)
- [CLI Reference](docs/CLI_REFERENCE.md)
- [Configuration Specification](docs/CONFIGURATION.md)
- [ICU MessageFormat Guide](docs/ICU_GUIDE.md)

## License

[MIT](LICENSE)
