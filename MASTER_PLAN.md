# localecraft Architecture and Engineering Master Plan

> Visual Software Localization Engine and Layout Overflow Sentinel  
> Focus: Multimodal Internationalization, Mathematical Typography Simulation, Autonomous Length Optimization, and ICU MessageFormat Integrity

---

## 1. Overview and Problem Statement

### 1.1 The Localization Layout Problem
When software products expand into international markets (German, Japanese, Arabic, French, Spanish, Russian), localization introduces visual regressions that are difficult to detect during pull request review:

1. **Text Expansion and Visual Clipping**:
   Translations in languages such as German and French often expand by 30% to 45% compared to English source strings. A button designed for "Save" (4 characters, ~50px) expands to "Speichern und anwenden" (22 characters, ~185px), clipping text, breaking fixed flex layouts, or producing ambiguous truncation like `Speich...`.

2. **ICU MessageFormat Syntax Corruption**:
   Modern applications rely on ICU MessageFormat for pluralization and gender variants:
   `{count, plural, one{# item} other{# items}}`
   Standard machine translation models frequently corrupt braces, rename variable tokens, or omit required plural categories, causing client-side runtime errors in frameworks such as `react-intl`, `next-intl`, and `i18next`.

3. **Absence of Visual Container Context**:
   Translation pipelines treat strings as isolated text records without metadata about their rendering target. Words such as "Post" or "Back" are translated without knowing whether the container is a navigation button, table column, or destructive modal trigger.

### 1.2 The localecraft Approach
`localecraft` is a single-binary Go engine with an embedded React 19 visual localization studio:
- **Layout Constraint Ingestion**: Associates localization keys with explicit container constraints (maximum pixel width, maximum lines, font family, font size, weight, letter spacing).
- **Mathematical Typography Engine**: Computes exact glyph bounding boxes and advance widths using pure Go TrueType/OpenType font parsing and HarfBuzz shaper integration.
- **Autonomous Compression Loop**: When a target translation exceeds the designated pixel budget, the engine rephrases the text iteratively under strict character bounds to produce concise, natural terminology without altering semantic meaning.
- **Strict ICU Syntax Gate**: Validates AST structure, token stability, and CLDR plural category compliance across target locales.
- **Embedded Studio and CLI**: Provides terminal auditing commands for CI/CD pipelines alongside an embedded web studio for interactive visual comparison and one-click pull requests.

---

## 2. System Architecture

```
                      +---------------------------------------------+
                      |         Source Localization Files           |
                      |        (Nested JSON, Gettext PO, ARB)       |
                      +---------------------------------------------+
                                             |
                                             v
                      +---------------------------------------------+
                      |             localecraft Engine              |
                      |        (Single Go Binary, Zero CGO)         |
                      +---------------------------------------------+
                                             |
                                             v
                      +---------------------------------------------+
                      |          Layout Constraint Parser           |
                      |  - Container max-width & max-lines          |
                      |  - Font family, size, weight, letter-space  |
                      |  - Contextual descriptions & UI targets     |
                      +---------------------------------------------+
                                             |
                                             v
                      +---------------------------------------------+
                      |         Context-Aware Translation           |
                      |  - Translates with screen context & terms   |
                      |  - Preserves ICU placeholders & arguments   |
                      +---------------------------------------------+
                                             |
                                             v
                      +---------------------------------------------+
                      |        Typography Measurement Engine        |
                      |  - Loads TTF/OTF fonts                      |
                      |  - Computes exact glyph advance widths (px) |
                      |  - Compares with container pixel limit      |
                      +---------------------------------------------+
                                             |
                                  +----------+----------+
                                  |                     |
                    (Fits: Width <= MaxWidth)   (Overflow: Width > MaxWidth)
                                  v                     v
                      +--------------------+  +------------------------------+
                      | Accept Translation |  |   Autonomous Rewording Loop  |
                      +--------------------+  | - Calculates compression ratio |
                                  |           | - Synthesizes concise phrase |
                                  |           +------------------------------+
                                  |                         |
                                  +<------------------------+
                                  |
                                  v
                      +---------------------------------------------+
                      |          ICU Syntax & Grammar Gate          |
                      |  - Validates balanced braces & variables    |
                      |  - Checks CLDR plural keyword parity        |
                      +---------------------------------------------+
                                             |
                        +--------------------+--------------------+
                        |                                         |
                        v                                         v
        +-------------------------------+         +-------------------------------+
        |     Updated Locale Files      |         |     React 19 Visual Studio    |
        | - Clean, non-overflowing text |         | - Multi-locale comparison     |
        | - Verified ICU syntax parity  |         | - Real-time button simulation |
        | - CI/CD audit pass/fail exit  |         | - Manual review & PR trigger  |
        +-------------------------------+         +-------------------------------+
```

---

## 3. Mathematical Typography and Overflow Formulation

Given:
- Font family $F$
- Font size $S$ in pixels
- Font weight $W$ (e.g., 400 regular, 600 semibold, 700 bold)
- Letter spacing $L$ in pixels
- Horizontal padding $P_{left}, P_{right}$

The rendered text width is calculated as:

$$\text{RenderedWidth}(T) = \sum_{i=1}^{n} \text{AdvanceWidth}(g_i, F, S, W) + (n - 1) \cdot L$$

Where $g_i$ represents the glyph corresponding to character $T[i]$ and $n = |T|$.

### Constraint Verification:
$$\text{AvailableWidth} = \text{MaxContainerWidth} - (P_{left} + P_{right})$$

$$\text{Status}(T) = \begin{cases} 
\text{PASS}, & \text{RenderedWidth}(T) \le \text{AvailableWidth} \\
\text{OVERFLOW}, & \text{RenderedWidth}(T) > \text{AvailableWidth}
\end{cases}$$

### Compression Ratio and Target Character Constraint:
When $\text{Status}(T) = \text{OVERFLOW}$, the compression factor $\rho$ is defined as:

$$\rho = \frac{\text{AvailableWidth}}{\text{RenderedWidth}(T)}$$

The rephrasing agent receives an explicit character ceiling:

$$\text{MaxAllowedChars} = \max\left(3, \lfloor |T| \cdot \rho \cdot 0.96 \rfloor\right)$$

This mathematical constraint prevents repeated overflow cycles by establishing a strict upper bound on candidate length before re-measurement.

---

## 4. Repository Structure

```
localecraft/
├── .github/
│   └── workflows/
│       ├── test.yml                # CI test and lint workflow
│       └── audit.yml               # GitHub Action for PR localization audits
├── cmd/
│   └── localecraft/
│       └── main.go                 # Root CLI entrypoint (translate, audit, serve, pr, measure, init)
├── internal/
│   ├── config/                     # Configuration loader for localecraft.yaml
│   │   ├── config.go
│   │   └── config_test.go
│   ├── parser/                     # Localization file parsers
│   │   ├── catalog.go              # Core catalog interfaces
│   │   ├── json_nested.go          # Nested and flat JSON format
│   │   ├── json_nested_test.go
│   │   ├── gettext_po.go           # GNU gettext .po format
│   │   ├── gettext_po_test.go
│   │   ├── flutter_arb.go          # Flutter .arb format
│   │   └── flutter_arb_test.go
│   ├── typography/                 # Font engine and pixel measurement
│   │   ├── engine.go               # Typography engine interface
│   │   ├── fonts/                  # Embedded default fonts (Inter Regular & SemiBold)
│   │   ├── loader.go               # TTF/OTF loader
│   │   ├── shaper.go               # Font advance width calculator
│   │   └── typography_test.go
│   ├── icu/                        # ICU MessageFormat parser and validator
│   │   ├── ast.go                  # AST node definitions
│   │   ├── parser.go               # MessageFormat tokenizer and parser
│   │   ├── validator.go            # Parity and CLDR plural validation
│   │   └── icu_test.go
│   ├── agent/                      # Translation and rewording engine
│   │   ├── engine.go               # Provider interface (Gemini, OpenAI, Heuristic)
│   │   ├── gemini.go               # Gemini API client
│   │   ├── openai.go               # OpenAI-compatible client
│   │   ├── heuristic.go            # Deterministic heuristic engine for tests/offline
│   │   ├── optimizer.go            # Bounded iterative rewording loop
│   │   └── agent_test.go
│   ├── store/                      # SQLite storage and cache layer
│   │   ├── db.go
│   │   ├── schema.go
│   │   ├── queries.go
│   │   └── store_test.go
│   ├── git/                        # GitHub PR automation
│   │   ├── pr_builder.go
│   │   └── git_test.go
│   └── api/                        # HTTP REST and SSE server
│       ├── server.go
│       ├── handlers.go
│       └── routes.go
├── web/                            # Embedded React 19 Visual Studio
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── tailwind.config.js
│   └── src/
│       ├── main.tsx
│       ├── App.tsx
│       ├── api.ts
│       ├── types.ts
│       ├── components/
│       │   ├── Header.tsx
│       │   ├── ComponentSimulator.tsx
│       │   ├── VisualButtonPreview.tsx
│       │   ├── CardPreview.tsx
│       │   ├── PixelMeter.tsx
│       │   ├── ICUTester.tsx
│       │   ├── KeyList.tsx
│       │   └── LanguageMatrix.tsx
│       └── pages/
│           ├── StudioPage.tsx
│           ├── AuditPage.tsx
│           └── SettingsPage.tsx
├── examples/
│   ├── localecraft.yaml
│   ├── locales/
│   │   ├── en.json
│   │   ├── de.json
│   │   └── fr.json
│   ├── po/
│   │   ├── messages.en.po
│   │   └── messages.de.po
│   └── arb/
│       ├── app_en.arb
│       └── app_de.arb
├── docs/
│   ├── ARCHITECTURE.md
│   ├── CLI_REFERENCE.md
│   ├── CONFIGURATION.md
│   └── ICU_GUIDE.md
├── go.mod
├── go.sum
├── Makefile
├── LICENSE
└── README.md
```

---

## 5. Specification Schema (`localecraft.yaml`)

```yaml
version: "1"
source_locale: "en"
target_locales:
  - "de"
  - "fr"
  - "es"
  - "ja"
  - "ar"

files:
  - path: "examples/locales/en.json"
    format: "json_nested"
    output_pattern: "examples/locales/{{.Locale}}.json"

ui_constraints:
  font_defaults:
    family: "Inter"
    size_px: 14
    weight: 500
    letter_spacing_px: 0.0
    padding_x_px: 16

  keys:
    "nav.checkout_btn":
      max_width_px: 140
      max_lines: 1
      context: "Primary call to action in navigation header."

    "settings.delete_account_confirm":
      max_width_px: 240
      max_lines: 1
      context: "Destructive confirmation button in account settings."

    "inbox.task_counter":
      format: "icu_messageformat"
      context: "Task count label in header."

glossary:
  - term: "Workspace"
    translations:
      de: "Arbeitsbereich"
      fr: "Espace de travail"
      ja: "ワークスペース"
```

---

## 6. Implementation Plan by Phases

### Phase 1: Core Foundation & Typography Engine
- Initialize Go module, project dependencies, and standard directory tree.
- Implement pure Go typography measurement engine with bundled OpenType font data.
- Build exact glyph advance measurement and letter spacing calculation.
- Implement format parsers: Nested JSON, Gettext `.po`, and Flutter `.arb`.
- Write comprehensive unit tests for font measurement and all three file parsers.

### Phase 2: ICU MessageFormat Engine & Grammar Validator
- Implement lexer and AST parser for ICU MessageFormat syntax:
  - Plain string nodes, argument nodes, plural nodes, select nodes, selectordinal nodes.
- Implement structural validator:
  - Verifies exact placeholder preservation across source and target.
  - Verifies CLDR plural categories per locale.
  - Detects unbalanced brackets and malformed syntax.
- Write unit tests for standard and complex ICU strings.

### Phase 3: Translation & Autonomous Optimization Agent
- Implement LLM provider interfaces (Gemini API, OpenAI-compatible, and Heuristic engine).
- Implement layout verification gate comparing rendered width against container width.
- Implement bounded iterative rewording loop with character limit feedback.
- Add glossary enforcement and preserve variable integrity.
- Write automated tests verifying convergence on length-constrained German/French phrases.

### Phase 4: Embedded React 19 Visual Studio & Web Server
- Build React 19 single-page application with Tailwind CSS:
  - Multi-locale matrix grid.
  - Interactive Component Simulator (buttons, badges, cards) with real-time bounding box markers and overflow warnings.
  - ICU visual tester and variable simulator.
  - Real-time pixel width bar and auto-shorten trigger.
- Implement Go HTTP REST and SSE server in `internal/api`.
- Embed frontend production build into single Go binary using Go `embed.FS`.

### Phase 5: CLI Surface, GitHub Actions & Integration Verification
- Implement CLI commands in `cmd/localecraft`:
  - `localecraft audit`: CI gate with exit code 1 on strict failures.
  - `localecraft translate`: Autonomous localization and shortening.
  - `localecraft measure`: Instant string width calculation.
  - `localecraft serve`: Embedded studio server.
  - `localecraft pr`: Automated GitHub pull request generator.
  - `localecraft init`: Starter configuration generator.
- Create GitHub Actions workflow (`.github/workflows/audit.yml`).
- Create full end-to-end integration test suite.
- Write clear, factual technical documentation adhering to humanizer guidelines.
