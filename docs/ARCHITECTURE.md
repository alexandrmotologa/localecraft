# Architecture Overview

`localecraft` is a self-contained localization engine written in Go with an embedded React 19 visual studio. It connects linguistic translation pipelines directly to physical layout constraints.

## System Components

### 1. Configuration Layer (`internal/config`)
Loads `localecraft.yaml` specifications. Parses source and target locale definitions, file mapping patterns, typography defaults, key-level container limits, glossary terminology, and AI provider configurations.

### 2. Localization File Parsers (`internal/parser`)
Supports three common localization file formats through a unified catalog interface:
- **Nested JSON (`json_nested`)**: Dot-separated key path hierarchy with structure preservation.
- **GNU Gettext (`gettext_po`)**: Message IDs, context identifiers (`msgctxt`), comments, and plural forms.
- **Flutter ARB (`flutter_arb`)**: Application Resource Bundle files with `@key` metadata and placeholder definitions.

### 3. Mathematical Typography Engine (`internal/typography`)
Calculates rendered bounding-box pixel dimensions without a headless browser:
- Parses TrueType and OpenType fonts using `golang.org/x/image/font/opentype`.
- Evaluates individual glyph advances and letter spacing at 72 DPI.
- Compares rendered text width against container limits minus horizontal padding.
- Derives compression factor ratios when text exceeds boundary widths.

### 4. ICU MessageFormat Engine (`internal/icu`)
Validates and executes ICU syntax:
- Lexer and recursive descent parser producing an AST of text, argument, plural, and select nodes.
- Syntax and variable parity verification between source and target messages.
- CLDR plural category checks for target languages (e.g., German, French, Spanish, Russian, Arabic, Romanian).
- Dynamic message evaluation supporting variable interpolation and plural category selection.

### 5. Autonomous Rewording Loop (`internal/agent`)
Coordinates iterative translation under length constraints:
- Pluggable provider architecture supporting Google Gemini, OpenAI-compatible endpoints, and deterministic heuristic evaluation.
- Iterative shortening loop that requests punchy synonyms within exact character bounds when container limits are breached.
- Halts when text fits within container boundaries or max iterations are exhausted.

### 6. Persistence and Web API (`internal/store`, `internal/api`)
- SQLite storage using `modernc.org/sqlite` without CGO dependencies.
- REST endpoints for measuring text, running translations, updating keys, and exporting files.
- Server-Sent Events (SSE) stream for real-time visual studio synchronization.

### 7. Embedded React 19 Visual Studio (`web/`)
- Single-page application compiled into the Go binary via `embed.FS`.
- Real-time container simulator displaying visual button and card previews with boundary markers.
- Side-by-side multi-locale matrix grid.
- Interactive ICU sandbox with parameter sliders.
