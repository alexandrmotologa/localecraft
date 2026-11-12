<p align="center">
  <img src="docs/images/logo.png?raw=true" alt="localecraft Logo" width="128" />
</p>

<h1 align="center">localecraft</h1>

<p align="center">
  <strong>Context-aware visual software localization engine and layout overflow sentinel.</strong><br />
  Calculates exact font bounding boxes, validates ICU MessageFormat syntax, and rephrases expanded text within container bounds.
</p>

<p align="center">
  <a href="https://github.com/alexandrmotologa/localecraft/actions"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go" alt="Go Version" /></a>
  <a href="https://react.dev"><img src="https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react" alt="React 19" /></a>
  <a href="https://tailwindcss.com"><img src="https://img.shields.io/badge/Tailwind-v4-38B2AC?style=flat&logo=tailwind-css" alt="Tailwind CSS" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat" alt="MIT License" /></a>
</p>

---

## Interactive Demo

<p align="center">
  <img src="docs/images/localecraft_demo.gif?raw=true" alt="localecraft Visual Studio Interactive Demo" width="100%" />
</p>

## Problem

Software localization frequently produces visual regressions:
* **Text expansion**: German, French, and Russian translations expand by 30% to 45% over English source strings, clipping text or breaking fixed containers.
* **Syntax corruption**: Automated translation pipelines frequently corrupt ICU MessageFormat brackets or drop placeholder variables, causing client runtime crashes.
* **Missing container context**: Strings translated without bounding box metadata produce labels that do not fit their visual target.

## Features

* **Mathematical Typography Engine**: Computes exact glyph advance widths and bounding box dimensions from OpenType/TrueType font metrics at build time.
* **Autonomous Rewording Loop**: Iteratively prompts language models under explicit character bounds to produce concise phrasing when target translations overflow UI containers.
* **Strict ICU MessageFormat Gate**: Verifies syntax balance, argument parity, and Unicode CLDR plural rules across locales.
* **Codebase Scanner (`localecraft scan`)**: Analyzes React, Vue, Svelte, and HTML codebases to discover translation keys and infer container constraints from Tailwind classes (`w-32`, `max-w-[140px]`).
* **Pseudo-localization Generator (`localecraft pseudoloc`)**: Transforms source strings into accented text with +30-50% expansion padding to test layout resilience.
* **Multi-Format Support**: Reads and writes Nested JSON, GNU Gettext (`.po`), Flutter ARB (`.arb`), CSV, and XLIFF 1.2.
* **Embedded React 19 Studio**: Visual component simulator with real-time boundary markers, RTL mirroring, typography playground, and batch auto-fix workflows.

## Visual Studio & Responsive Layouts

The embedded web studio provides real-time layout simulation across desktop, tablet, and mobile viewports.

### Desktop Studio

<p align="center">
  <img src="docs/images/screenshot_desktop.png?raw=true" alt="localecraft Desktop Studio" width="100%" />
</p>

### Responsive Tablet & Mobile

<table align="center" width="100%">
  <tr>
    <td align="center" width="55%">
      <strong>Tablet Layout (768px)</strong><br />
      <img src="docs/images/screenshot_tablet.png?raw=true" alt="Tablet Layout" width="100%" />
    </td>
    <td align="center" width="45%">
      <strong>Mobile Layout (390px)</strong><br />
      <img src="docs/images/screenshot_mobile.png?raw=true" alt="Mobile Layout" width="100%" />
    </td>
  </tr>
</table>

### Language Matrix & ICU Syntax Sandbox

<table align="center" width="100%">
  <tr>
    <td align="center" width="50%">
      <strong>Language Translation Matrix</strong><br />
      <img src="docs/images/screenshot_matrix.png?raw=true" alt="Language Matrix" width="100%" />
    </td>
    <td align="center" width="50%">
      <strong>ICU MessageFormat Sandbox</strong><br />
      <img src="docs/images/screenshot_icu.png?raw=true" alt="ICU Sandbox" width="100%" />
    </td>
  </tr>
</table>

## Architecture

```
Source Files (.json, .po, .arb, .xliff, .csv)
                    │
                    ▼
          localecraft CLI / Engine
                    │
          ┌─────────┴─────────┐
          ▼                   ▼
    Layout Constraints     ICU Syntax Gate
   (Font & Container)     (Brackets & Parity)
          │                   │
          ▼                   │
   Mathematical Typography Engine
   (Glyph advance calculations)
          │
          ├─────────────────────────────┐
          │ (Fits)                      │ (Overflows)
          ▼                             ▼
   Approved Locale Files       Autonomous Rewording Loop
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

### 4. Scan Frontend Codebase for Constraints

```bash
./localecraft scan --dir ./src --update-config
```

### 5. Generate Pseudo-localization for Layout Testing

```bash
./localecraft pseudoloc --config ./examples/localecraft.yaml --expansion 40
```

### 6. Autonomous Translation and Shortening

```bash
./localecraft translate --config ./examples/localecraft.yaml
```

The optimizer detects overflow on long strings and compresses them to fit within container limits.

### 7. Export and Import for External Translators

```bash
./localecraft export --format csv --locale de --output ./translations_de.csv
./localecraft import --input ./translations_de.csv --locale de
```

### 8. Launch Embedded Visual Studio

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

* **Nested and Flat JSON**: dot-separated key hierarchy (`json_nested`)
* **GNU Gettext**: PO and POT files with context and plurals (`gettext_po`)
* **Flutter ARB**: Application Resource Bundle files (`flutter_arb`)
* **Tabular CSV**: Comma-separated values for spreadsheets (`csv`)
* **XLIFF 1.2**: Standard XML Localization Interchange File Format (`xliff`)

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

* [Architecture Overview](docs/ARCHITECTURE.md)
* [CLI Reference](docs/CLI_REFERENCE.md)
* [Configuration Specification](docs/CONFIGURATION.md)
* [ICU MessageFormat Guide](docs/ICU_GUIDE.md)

## License

[MIT](LICENSE)
