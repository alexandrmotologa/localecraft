# Configuration Specification (`localecraft.yaml`)

`localecraft.yaml` defines project localization rules, typography parameters, container boundaries, and AI provider credentials.

## Schema Example

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
    letter_spacing_px: 0.0
    padding_x_px: 12
    line_height: 1.4

  keys:
    "nav.checkout_btn":
      max_width_px: 120
      max_lines: 1
      context: "Primary checkout CTA in top header bar."

    "settings.delete_account_confirm":
      max_width_px: 220
      max_lines: 1
      context: "Destructive confirmation button in red danger zone."

    "dashboard.inbox_counter":
      format: "icu_messageformat"
      context: "Inbox task counter on header."

glossary:
  - term: "Workspace"
    translations:
      de: "Arbeitsbereich"
      fr: "Espace de travail"
      ja: "ワークスペース"

ai:
  provider: "heuristic" # "gemini", "openai", "heuristic"
  model: "gemini-2.5-flash"
  api_key: ""
  base_url: ""
```

## Section Details

### `files`
- `path`: Path to source locale file relative to `localecraft.yaml`.
- `format`: One of `json_nested`, `gettext_po`, or `flutter_arb`.
- `output_pattern`: Destination pattern supporting `{{.Locale}}`.

### `ui_constraints.font_defaults`
- `family`: Font family name. Built-in values include `Inter`, `Roboto`, `mono`, or `default`. Custom TTF files can be loaded by placing them in the working tree.
- `size_px`: Font size in pixels (default 14).
- `weight`: CSS numeric weight (400 regular, 500 medium, 600 semibold, 700 bold).
- `padding_x_px`: Total horizontal padding per container (left + right).

### `ui_constraints.keys`
- `max_width_px`: Container pixel budget.
- `max_lines`: Maximum line wrapping count.
- `context`: Semantic screen context injected into AI prompts.
- `format`: Set to `icu_messageformat` to enforce syntax validation.

### `glossary`
List of terms with predefined translations across target locales. Takes precedence over generic language models.
