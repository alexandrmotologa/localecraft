# CLI Reference

`localecraft` provides command-line tools for auditing catalogs, calculating font measurements, running autonomous translations, and hosting the embedded visual studio.

## Commands

### `localecraft audit`
Scans localization files and tests each key against physical layout limits and ICU syntax rules.

```bash
# Standard terminal report
localecraft audit --config ./localecraft.yaml

# Fail with exit code 1 if overflows or ICU errors are found (recommended for CI/CD)
localecraft audit --config ./localecraft.yaml --strict

# Output audit report as JSON
localecraft audit --config ./localecraft.yaml --json
```

### `localecraft translate`
Translates missing or overflown keys with autonomous length compression loops.

```bash
# Translate using configured provider
localecraft translate --config ./localecraft.yaml

# Filter by target locale
localecraft translate --config ./localecraft.yaml --locale de

# Translate a single key
localecraft translate --config ./localecraft.yaml --key nav.checkout_btn

# Override AI provider at runtime
localecraft translate --config ./localecraft.yaml --provider gemini --api-key "$GEMINI_API_KEY"
```

### `localecraft measure`
Calculates rendered pixel dimensions for any text string and font configuration.

```bash
# Basic measurement
localecraft measure --text "Speichern und anwenden"

# Constraint check with container max width
localecraft measure --text "Speichern und anwenden" --font "Inter" --size 14 --weight 600 --max-width 120

# Output as JSON
localecraft measure --text "Checkout" --size 14 --json
```

### `localecraft serve`
Starts the HTTP API server and serves the embedded React visual localization studio.

```bash
# Start server on default port 8080
localecraft serve --config ./localecraft.yaml

# Custom port and auto-open browser
localecraft serve --config ./localecraft.yaml --port 3000 --open
```

### `localecraft pr`
Commits approved localization updates and creates a GitHub pull request using the GitHub CLI.

```bash
localecraft pr --config ./localecraft.yaml --branch "update-locales" --title "i18n: layout updates"
```

### `localecraft init`
Generates a starter configuration file and sample catalogs in a designated directory.

```bash
localecraft init --path ./my-project
```
