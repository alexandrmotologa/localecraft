export interface TranslationRecord {
  id: number;
  locale: string;
  value: string;
  status: 'ACCEPTED' | 'OVERFLOW' | 'MISSING' | 'ICU_ERROR';
  rendered_width_px: number;
  is_overflow: boolean;
  overflow_px: number;
  overflow_pct: number;
  iterations_used: number;
  updated_at: string;
}

export interface KeyRecord {
  id: number;
  key_name: string;
  source_locale: string;
  source_value: string;
  context: string;
  max_width_px: number;
  font_family: string;
  font_size_px: number;
  font_weight: number;
  is_icu: boolean;
  translations: Record<string, TranslationRecord>;
}

export interface StoreSummary {
  total_keys: number;
  total_locales: number;
  total_approved: number;
  total_overflows: number;
  total_missing: number;
  total_icu_errors: number;
}

export interface FontDefaults {
  family: string;
  size_px: number;
  weight: number;
  letter_spacing_px: number;
  padding_x_px: number;
  line_height: number;
}

export interface KeyConstraint {
  max_width_px: number;
  max_lines: number;
  context: string;
  format: string;
  allowed_overflow_pct: number;
  font_family: string;
  font_size_px: number;
  font_weight: number;
}

export interface GlossaryTerm {
  term: string;
  translations: Record<string, string>;
}

export interface Config {
  version: string;
  source_locale: string;
  target_locales: string[];
  files: Array<{
    path: string;
    format: string;
    output_pattern: string;
  }>;
  ui_constraints: {
    font_defaults: FontDefaults;
    keys: Record<string, KeyConstraint>;
  };
  glossary: GlossaryTerm[];
}

export interface MeasureResult {
  text: string;
  rendered_width_px: number;
  rendered_height_px: number;
  available_width_px: number;
  is_overflow: boolean;
  overflow_px: number;
  overflow_pct: number;
  target_char_limit: number;
  glyph_count: number;
  lines_used: number;
  compression_ratio: number;
}

export interface OptimizeResult {
  key: string;
  source_text: string;
  final_translation: string;
  status: string;
  iterations_used: number;
  final_measure: MeasureResult;
  attempts: Array<{
    iteration: number;
    translation: string;
    measure: MeasureResult;
    icu_valid: boolean;
    icu_errors?: string[];
    target_char_limit: number;
    compression_ratio: number;
  }>;
}
