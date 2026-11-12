import React from 'react';
import { Sliders, RefreshCw, CheckCircle2, AlertCircle } from 'lucide-react';
import { KeyRecord } from '../types';

interface TypographyPlaygroundProps {
  keys: KeyRecord[];
  fontSizeDelta: number;
  setFontSizeDelta: (delta: number) => void;
  paddingDelta: number;
  setPaddingDelta: (delta: number) => void;
  onReset: () => void;
}

export const TypographyPlayground: React.FC<TypographyPlaygroundProps> = ({
  keys,
  fontSizeDelta,
  setFontSizeDelta,
  paddingDelta,
  setPaddingDelta,
  onReset,
}) => {
  // Count how many overflows would be resolved with current deltas
  let totalEvaluated = 0;
  let simulatedOverflows = 0;
  let baselineOverflows = 0;

  for (const k of keys) {
    for (const tr of Object.values(k.translations)) {
      if (tr.status === 'OVERFLOW') {
        baselineOverflows++;
      }
      if (k.max_width_px > 0 && tr.value) {
        totalEvaluated++;
        const effFontSize = Math.max(10, (k.font_size_px || 14) + fontSizeDelta);
        const effPadding = Math.max(0, 12 + paddingDelta);
        const avail = Math.max(0, k.max_width_px - effPadding * 2);

        // Simulated width scaling
        const fontScale = effFontSize / (k.font_size_px || 14);
        const simWidth = tr.rendered_width_px * fontScale;

        if (simWidth > avail) {
          simulatedOverflows++;
        }
      }
    }
  }

  const resolvedCount = Math.max(0, baselineOverflows - simulatedOverflows);
  const isTweaked = fontSizeDelta !== 0 || paddingDelta !== 0;

  return (
    <div className="bg-slate-900/80 rounded-xl border border-slate-800 p-5 space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <div className="w-8 h-8 rounded-lg bg-teal-500/10 border border-teal-500/30 flex items-center justify-center text-teal-400">
            <Sliders className="w-4 h-4" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-slate-100">Typography & Spacing Playground</h3>
            <p className="text-xs text-slate-400">Simulate global CSS styling changes without modifying codebase</p>
          </div>
        </div>

        {isTweaked && (
          <button
            onClick={onReset}
            className="flex items-center space-x-1 text-xs text-slate-400 hover:text-slate-200 px-2 py-1 rounded bg-slate-800 border border-slate-700 transition"
          >
            <RefreshCw className="w-3 h-3" />
            <span>Reset Sliders</span>
          </button>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-5 pt-2">
        {/* Font Size Slider */}
        <div className="space-y-2 bg-slate-950 p-3.5 rounded-lg border border-slate-800">
          <div className="flex items-center justify-between text-xs">
            <span className="text-slate-300 font-medium">Font Size Delta</span>
            <span className="font-mono text-teal-400 font-semibold">
              {fontSizeDelta > 0 ? `+${fontSizeDelta}` : fontSizeDelta}px
            </span>
          </div>
          <input
            type="range"
            min={-3}
            max={3}
            step={1}
            value={fontSizeDelta}
            onChange={(e) => setFontSizeDelta(parseInt(e.target.value, 10))}
            className="w-full accent-teal-500 cursor-pointer"
          />
          <div className="flex justify-between text-[10px] text-slate-500 font-mono">
            <span>-3px (11px)</span>
            <span>0px (14px)</span>
            <span>+3px (17px)</span>
          </div>
        </div>

        {/* Padding Slider */}
        <div className="space-y-2 bg-slate-950 p-3.5 rounded-lg border border-slate-800">
          <div className="flex items-center justify-between text-xs">
            <span className="text-slate-300 font-medium">Container Padding (X) Delta</span>
            <span className="font-mono text-teal-400 font-semibold">
              {paddingDelta > 0 ? `+${paddingDelta}` : paddingDelta}px
            </span>
          </div>
          <input
            type="range"
            min={-6}
            max={6}
            step={1}
            value={paddingDelta}
            onChange={(e) => setPaddingDelta(parseInt(e.target.value, 10))}
            className="w-full accent-teal-500 cursor-pointer"
          />
          <div className="flex justify-between text-[10px] text-slate-500 font-mono">
            <span>-6px (6px)</span>
            <span>0px (12px)</span>
            <span>+6px (18px)</span>
          </div>
        </div>
      </div>

      {/* Real-time simulation feedback */}
      <div className="p-3 bg-slate-950 rounded-lg border border-slate-800 flex items-center justify-between text-xs">
        <div className="flex items-center space-x-2">
          {resolvedCount > 0 ? (
            <span className="flex items-center space-x-1.5 text-emerald-400 font-medium">
              <CheckCircle2 className="w-4 h-4" />
              <span>
                Styling resolves <strong>{resolvedCount}</strong> layout overflow{resolvedCount > 1 ? 's' : ''} across all locales!
              </span>
            </span>
          ) : (
            <span className="flex items-center space-x-1.5 text-slate-400">
              <AlertCircle className="w-4 h-4 text-slate-500" />
              <span>Simulated overflows: <strong>{simulatedOverflows}</strong> (Baseline: {baselineOverflows})</span>
            </span>
          )}
        </div>
        <span className="text-[11px] font-mono text-slate-500">
          Evaluated {totalEvaluated} locale keys
        </span>
      </div>
    </div>
  );
};
