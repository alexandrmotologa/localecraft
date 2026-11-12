import React, { useState, useEffect } from 'react';
import { KeyRecord, TranslationRecord } from '../types';
import { PixelMeter } from './PixelMeter';
import { Sparkles, RefreshCw, AlertTriangle, CheckCircle, ArrowLeftRight } from 'lucide-react';

interface ComponentSimulatorProps {
  keyRecord: KeyRecord;
  targetLocale: string;
  setTargetLocale: (loc: string) => void;
  onReword: (key: string, locale: string) => Promise<void>;
  isRewording: boolean;
}

const RTL_LOCALES = new Set(['ar', 'he', 'fa', 'ur']);

export const ComponentSimulator: React.FC<ComponentSimulatorProps> = ({
  keyRecord,
  targetLocale,
  setTargetLocale,
  onReword,
  isRewording,
}) => {
  const [componentType, setComponentType] = useState<'button' | 'destructive' | 'pill' | 'card'>('button');
  const [isRTL, setIsRTL] = useState(false);

  // Auto-detect RTL when target locale changes
  useEffect(() => {
    const base = targetLocale.split('-')[0].toLowerCase();
    setIsRTL(RTL_LOCALES.has(base));
  }, [targetLocale]);

  const currentTranslation: TranslationRecord | undefined = keyRecord.translations[targetLocale];
  const activeText = currentTranslation?.value || keyRecord.source_value;
  const isOverflow = currentTranslation?.is_overflow ?? false;
  const renderedWidth = currentTranslation?.rendered_width_px ?? 60;
  const maxWidth = keyRecord.max_width_px > 0 ? keyRecord.max_width_px : 140;
  const paddingX = 12;
  const availableWidth = Math.max(0, maxWidth - paddingX * 2);

  return (
    <div className="bg-slate-900/60 rounded-xl border border-slate-800 p-5 space-y-5">
      {/* Header & Controls */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-sm font-semibold text-slate-100 flex items-center space-x-2">
            <span>Visual Component Simulator</span>
            <span className="text-xs font-mono font-normal text-slate-400">
              ({keyRecord.key_name})
            </span>
          </h3>
          <p className="text-xs text-slate-400">
            {keyRecord.context || 'Visual bounding box simulation for live UI controls'}
          </p>
        </div>

        {/* Component Selector & RTL Toggle */}
        <div className="flex items-center space-x-2">
          <button
            onClick={() => setIsRTL(!isRTL)}
            className={`flex items-center space-x-1 px-2.5 py-1 rounded-lg border text-xs font-mono transition ${
              isRTL
                ? 'bg-sky-500/20 text-sky-400 border-sky-500/40 font-medium'
                : 'bg-slate-950 text-slate-400 hover:text-slate-200 border-slate-800'
            }`}
            title="Toggle Right-to-Left (RTL) Layout"
          >
            <ArrowLeftRight className="w-3.5 h-3.5" />
            <span>RTL</span>
          </button>

          <div className="flex items-center space-x-1 bg-slate-950 p-1 rounded-lg border border-slate-800 text-xs">
            {(['button', 'destructive', 'pill', 'card'] as const).map((t) => (
              <button
                key={t}
                onClick={() => setComponentType(t)}
                className={`px-2.5 py-1 rounded capitalize transition ${
                  componentType === t
                    ? 'bg-slate-800 text-emerald-400 font-medium'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                {t}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Locale Tabs & Auto Reword */}
      <div className="flex items-center justify-between border-b border-slate-800/80 pb-3">
        <div className="flex items-center space-x-1.5">
          {Object.keys(keyRecord.translations).map((loc) => {
            const tr = keyRecord.translations[loc];
            const active = loc === targetLocale;
            return (
              <button
                key={loc}
                onClick={() => setTargetLocale(loc)}
                className={`px-3 py-1.5 rounded-lg text-xs font-medium uppercase font-mono transition flex items-center space-x-1.5 ${
                  active
                    ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/40'
                    : 'bg-slate-950 text-slate-400 hover:text-slate-200 border border-slate-800'
                }`}
              >
                <span>{loc}</span>
                {tr?.is_overflow && <span className="w-1.5 h-1.5 rounded-full bg-rose-400" />}
              </button>
            );
          })}
        </div>

        <button
          onClick={() => onReword(keyRecord.key_name, targetLocale)}
          disabled={isRewording}
          className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white shadow-sm transition disabled:opacity-50"
        >
          {isRewording ? (
            <RefreshCw className="w-3.5 h-3.5 animate-spin" />
          ) : (
            <Sparkles className="w-3.5 h-3.5" />
          )}
          <span>{isRewording ? 'Synthesizing...' : 'Autonomous Reword'}</span>
        </button>
      </div>

      {/* Simulator Stage */}
      <div className="relative py-12 px-6 bg-slate-950 rounded-xl border border-slate-800/90 flex flex-col items-center justify-center overflow-hidden">
        {/* Background Grid Pattern */}
        <div
          className="absolute inset-0 opacity-10 pointer-events-none"
          style={{
            backgroundImage: `radial-gradient(#10b981 1px, transparent 1px)`,
            backgroundSize: '16px 16px',
          }}
        />

        {/* Physical Container Constraint Box */}
        <div
          dir={isRTL ? 'rtl' : 'ltr'}
          className="relative transition-all flex flex-col items-center"
          style={{ width: `${maxWidth}px` }}
        >
          {/* Top Width Ruler */}
          <div className="w-full flex items-center justify-between text-[10px] font-mono text-slate-500 mb-1 border-b border-dashed border-slate-700 pb-0.5">
            <span>{isRTL ? `${maxWidth}px` : '0px'}</span>
            <span className="text-emerald-400 font-semibold">{maxWidth}px max {isRTL && '(RTL)'}</span>
            <span>{isRTL ? '0px' : `${maxWidth}px`}</span>
          </div>

          {/* Left / Right Vertical Boundary Markers */}
          <div className="absolute left-0 top-6 bottom-0 w-px border-l border-dashed border-slate-700" />
          <div className="absolute right-0 top-6 bottom-0 w-px border-r border-dashed border-slate-700" />

          {/* The Actual Rendered Component */}
          <div className="py-2 w-full flex justify-center">
            {componentType === 'button' && (
              <button
                className={`px-3 py-1.5 rounded-lg text-sm font-medium transition flex items-center justify-center shadow whitespace-nowrap ${
                  isOverflow
                    ? 'bg-rose-500/20 text-rose-300 border border-rose-500'
                    : 'bg-emerald-600 hover:bg-emerald-500 text-white'
                }`}
                style={{
                  fontFamily: keyRecord.font_family || 'Inter',
                  fontSize: `${keyRecord.font_size_px || 14}px`,
                  fontWeight: keyRecord.font_weight || 500,
                }}
              >
                {activeText}
              </button>
            )}

            {componentType === 'destructive' && (
              <button
                className={`px-3 py-1.5 rounded-lg text-sm font-medium transition flex items-center justify-center shadow whitespace-nowrap ${
                  isOverflow
                    ? 'bg-rose-950 text-rose-300 border border-rose-500 ring-2 ring-rose-500/30'
                    : 'bg-rose-600 hover:bg-rose-500 text-white'
                }`}
                style={{
                  fontFamily: keyRecord.font_family || 'Inter',
                  fontSize: `${keyRecord.font_size_px || 14}px`,
                  fontWeight: keyRecord.font_weight || 600,
                }}
              >
                {activeText}
              </button>
            )}

            {componentType === 'pill' && (
              <span
                className={`px-3 py-1 rounded-full text-xs font-mono transition flex items-center justify-center whitespace-nowrap ${
                  isOverflow
                    ? 'bg-rose-500/20 text-rose-300 border border-rose-500'
                    : 'bg-slate-800 text-slate-200 border border-slate-700'
                }`}
              >
                {activeText}
              </span>
            )}

            {componentType === 'card' && (
              <div
                className={`p-3 rounded-lg border w-full text-center transition ${
                  isOverflow
                    ? 'bg-rose-500/10 border-rose-500/50 text-rose-300'
                    : 'bg-slate-900 border-slate-800 text-slate-200'
                }`}
              >
                <div className="text-xs font-medium truncate">{activeText}</div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Measurement Metrics Bar */}
      <div className="bg-slate-950 p-4 rounded-lg border border-slate-800 space-y-2">
        <PixelMeter
          renderedWidth={renderedWidth}
          availableWidth={availableWidth}
          maxWidth={maxWidth}
          isOverflow={isOverflow}
          overflowPX={currentTranslation?.overflow_px}
          overflowPCT={currentTranslation?.overflow_pct}
        />

        <div className="flex items-center justify-between text-xs text-slate-400 pt-2 border-t border-slate-800/80">
          <div className="flex items-center space-x-2">
            {isOverflow ? (
              <span className="flex items-center space-x-1 text-rose-400 font-medium">
                <AlertTriangle className="w-3.5 h-3.5" />
                <span>Text exceeds container by {currentTranslation?.overflow_px?.toFixed(1)}px</span>
              </span>
            ) : (
              <span className="flex items-center space-x-1 text-emerald-400 font-medium">
                <CheckCircle className="w-3.5 h-3.5" />
                <span>Fits within physical container</span>
              </span>
            )}
          </div>
          <div className="font-mono text-[11px] text-slate-500">
            Font: {keyRecord.font_family} {keyRecord.font_size_px}px w{keyRecord.font_weight}
          </div>
        </div>
      </div>
    </div>
  );
};
