import React, { useState } from 'react';
import { KeyRecord } from '../types';
import { Sparkles, Save, Check, RefreshCw, AlertTriangle, CheckCircle } from 'lucide-react';

interface LanguageMatrixProps {
  keyRecord: KeyRecord;
  onSaveTranslation: (key: string, locale: string, value: string) => Promise<void>;
  onReword: (key: string, locale: string) => Promise<void>;
  rewordingLocales: Record<string, boolean>;
}

export const LanguageMatrix: React.FC<LanguageMatrixProps> = ({
  keyRecord,
  onSaveTranslation,
  onReword,
  rewordingLocales,
}) => {
  const [editedValues, setEditedValues] = useState<Record<string, string>>({});
  const [savedLocales, setSavedLocales] = useState<Record<string, boolean>>({});

  const handleInputChange = (locale: string, val: string) => {
    setEditedValues((prev) => ({ ...prev, [locale]: val }));
    setSavedLocales((prev) => ({ ...prev, [locale]: false }));
  };

  const handleSave = async (locale: string) => {
    const val = editedValues[locale] ?? keyRecord.translations[locale]?.value ?? '';
    await onSaveTranslation(keyRecord.key_name, locale, val);
    setSavedLocales((prev) => ({ ...prev, [locale]: true }));
    setTimeout(() => {
      setSavedLocales((prev) => ({ ...prev, [locale]: false }));
    }, 2000);
  };

  return (
    <div className="bg-slate-900/60 rounded-xl border border-slate-800 p-5 space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-sm font-semibold text-slate-100 flex items-center space-x-2">
            <span>Multi-Locale Translation Matrix</span>
            <span className="text-xs font-mono font-normal text-slate-400">
              ({keyRecord.key_name})
            </span>
          </h3>
          <p className="text-xs text-slate-400">
            Source ({keyRecord.source_locale}): <strong className="text-slate-200">"{keyRecord.source_value}"</strong>
          </p>
        </div>
      </div>

      <div className="divide-y divide-slate-800/60 border border-slate-800 rounded-lg overflow-hidden bg-slate-950">
        {Object.entries(keyRecord.translations).map(([locale, trans]) => {
          const isRewording = rewordingLocales[locale] || false;
          const isSaved = savedLocales[locale] || false;
          const currentVal = editedValues[locale] ?? trans.value;
          const isOverflow = trans.status === 'OVERFLOW';

          return (
            <div key={locale} className="p-3.5 flex flex-col md:flex-row items-start md:items-center justify-between gap-3 hover:bg-slate-900/40 transition">
              {/* Locale & Status Badge */}
              <div className="flex items-center space-x-2.5 min-w-[100px]">
                <span className="font-mono text-xs uppercase px-2 py-1 bg-slate-800 text-slate-200 rounded font-semibold">
                  {locale}
                </span>
                {isOverflow ? (
                  <span className="flex items-center space-x-1 text-[11px] text-rose-400 bg-rose-500/10 px-1.5 py-0.5 rounded border border-rose-500/20 font-medium">
                    <AlertTriangle className="w-3 h-3" />
                    <span>Overflow</span>
                  </span>
                ) : trans.status === 'ACCEPTED' ? (
                  <span className="flex items-center space-x-1 text-[11px] text-emerald-400 bg-emerald-500/10 px-1.5 py-0.5 rounded border border-emerald-500/20 font-medium">
                    <CheckCircle className="w-3 h-3" />
                    <span>Fits</span>
                  </span>
                ) : (
                  <span className="text-[11px] text-slate-400 bg-slate-800 px-1.5 py-0.5 rounded font-medium">
                    {trans.status}
                  </span>
                )}
              </div>

              {/* Translation Input */}
              <div className="flex-1 w-full">
                <input
                  type="text"
                  value={currentVal}
                  onChange={(e) => handleInputChange(locale, e.target.value)}
                  className={`w-full px-3 py-1.5 rounded-lg text-xs bg-slate-900 border transition focus:outline-none ${
                    isOverflow
                      ? 'border-rose-500/50 text-rose-200 focus:border-rose-400'
                      : 'border-slate-800 text-slate-200 focus:border-emerald-500/60'
                  }`}
                />
              </div>

              {/* Pixel Width Stats */}
              <div className="min-w-[120px] text-right text-xs font-mono text-slate-400">
                <span className={isOverflow ? 'text-rose-400 font-semibold' : 'text-slate-300'}>
                  {trans.rendered_width_px.toFixed(1)}px
                </span>
                <span className="text-slate-600"> / {keyRecord.max_width_px}px</span>
              </div>

              {/* Actions */}
              <div className="flex items-center space-x-1.5">
                <button
                  onClick={() => handleSave(locale)}
                  className={`p-1.5 rounded-lg text-xs border transition ${
                    isSaved
                      ? 'bg-emerald-500/20 text-emerald-400 border-emerald-500/40'
                      : 'bg-slate-800 hover:bg-slate-700 text-slate-300 border-slate-700'
                  }`}
                  title="Save Translation"
                >
                  {isSaved ? <Check className="w-3.5 h-3.5" /> : <Save className="w-3.5 h-3.5" />}
                </button>

                <button
                  onClick={() => onReword(keyRecord.key_name, locale)}
                  disabled={isRewording}
                  className="flex items-center space-x-1 px-2.5 py-1.5 rounded-lg text-xs font-medium bg-emerald-600/20 hover:bg-emerald-600/30 text-emerald-300 border border-emerald-500/30 transition disabled:opacity-50"
                  title="Auto-Reword with Character Constraint"
                >
                  {isRewording ? (
                    <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                  ) : (
                    <Sparkles className="w-3.5 h-3.5" />
                  )}
                  <span className="hidden sm:inline">Shorten</span>
                </button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
