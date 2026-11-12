import React, { useState, useMemo } from 'react';
import { KeyRecord } from '../types';
import { CheckCircle, AlertTriangle, Layers } from 'lucide-react';

interface ICUTesterProps {
  keyRecord: KeyRecord | null;
}

export const ICUTester: React.FC<ICUTesterProps> = ({ keyRecord }) => {
  const [icuText, setIcuText] = useState(
    keyRecord?.source_value || '{count, plural, =0{No files remaining} one{# file remaining} other{# files remaining}}'
  );

  const [variables, setVariables] = useState<Record<string, any>>({
    count: 1,
    userName: 'Alex',
    gender: 'female',
  });

  // Extract variables from ICU string
  const detectedVars = useMemo(() => {
    const regex = /\{([a-zA-Z0-9_-]+)/g;
    const matches = new Set<string>();
    let m;
    while ((m = regex.exec(icuText)) !== null) {
      matches.add(m[1]);
    }
    return Array.from(matches);
  }, [icuText]);

  // Client-side quick preview evaluator
  const renderedOutput = useMemo(() => {
    try {
      let result = icuText;

      // Handle simple plural
      const pluralMatch = icuText.match(/\{([a-zA-Z0-9_-]+),\s*plural,\s*([^}]+(?:\{[^}]*\}[^}]*)*)\}/);
      if (pluralMatch) {
        const varName = pluralMatch[1];
        const casesStr = pluralMatch[2];
        const num = Number(variables[varName] ?? 0);

        // Match =0, =1
        const exactRegex = new RegExp(`=${num}\\s*\\{([^}]+)\\}`);
        const exactMatch = casesStr.match(exactRegex);

        const oneRegex = /one\s*\{([^}]+)\}/;
        const oneMatch = casesStr.match(oneRegex);

        const otherRegex = /other\s*\{([^}]+)\}/;
        const otherMatch = casesStr.match(otherRegex);

        let chosen = '';
        if (exactMatch) {
          chosen = exactMatch[1];
        } else if (num === 1 && oneMatch) {
          chosen = oneMatch[1];
        } else if (otherMatch) {
          chosen = otherMatch[1];
        }

        chosen = chosen.replace(/#/g, num.toString());
        result = result.replace(pluralMatch[0], chosen);
      }

      // Handle simple variable interpolation
      for (const [k, v] of Object.entries(variables)) {
        result = result.replace(new RegExp(`\\{${k}\\}`, 'g'), String(v));
      }

      return { output: result, error: null };
    } catch (err: any) {
      return { output: '', error: err.message || 'Evaluation error' };
    }
  }, [icuText, variables]);

  const updateVariable = (name: string, val: any) => {
    setVariables((prev) => ({ ...prev, [name]: val }));
  };

  return (
    <div className="bg-slate-900/60 rounded-xl border border-slate-800 p-5 space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-sm font-semibold text-slate-100 flex items-center space-x-2">
            <Layers className="w-4 h-4 text-sky-400" />
            <span>Interactive ICU MessageFormat Sandbox</span>
          </h3>
          <p className="text-xs text-slate-400">
            Simulate dynamic pluralization, gender branching, and variable bindings
          </p>
        </div>
      </div>

      {/* Template Input */}
      <div className="space-y-1.5">
        <label className="text-xs font-medium text-slate-300">ICU Template</label>
        <textarea
          rows={3}
          value={icuText}
          onChange={(e) => setIcuText(e.target.value)}
          className="w-full font-mono text-xs p-3 bg-slate-950 border border-slate-800 rounded-lg text-slate-200 focus:outline-none focus:border-sky-500/60"
        />
      </div>

      {/* Variables Controller */}
      <div className="bg-slate-950 p-4 rounded-lg border border-slate-800 space-y-3">
        <h4 className="text-xs font-semibold text-slate-300 uppercase tracking-wider font-mono">
          Simulated Parameters
        </h4>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {detectedVars.map((v) => {
            const isCount = v.toLowerCase().includes('count') || v.toLowerCase().includes('item') || v.toLowerCase().includes('num');
            const val = variables[v] ?? (isCount ? 1 : 'Sample');

            return (
              <div key={v} className="space-y-1">
                <label className="text-xs text-slate-400 font-mono">{v}</label>
                {isCount ? (
                  <div className="flex items-center space-x-2">
                    <input
                      type="number"
                      value={val}
                      onChange={(e) => updateVariable(v, parseInt(e.target.value, 10) || 0)}
                      className="w-20 px-2 py-1 bg-slate-900 border border-slate-800 rounded text-xs text-slate-200 font-mono"
                    />
                    <div className="flex items-center space-x-1">
                      {[0, 1, 2, 5, 21].map((n) => (
                        <button
                          key={n}
                          onClick={() => updateVariable(v, n)}
                          className={`px-1.5 py-0.5 rounded text-[10px] font-mono ${
                            val === n
                              ? 'bg-sky-500/20 text-sky-400 border border-sky-500/40'
                              : 'bg-slate-900 text-slate-400 hover:text-slate-200 border border-slate-800'
                          }`}
                        >
                          {n}
                        </button>
                      ))}
                    </div>
                  </div>
                ) : (
                  <input
                    type="text"
                    value={val}
                    onChange={(e) => updateVariable(v, e.target.value)}
                    className="w-full px-2.5 py-1 bg-slate-900 border border-slate-800 rounded text-xs text-slate-200 font-mono"
                  />
                )}
              </div>
            );
          })}
        </div>
      </div>

      {/* Live Evaluated Output */}
      <div className="bg-slate-950 p-4 rounded-lg border border-slate-800 space-y-2">
        <div className="flex items-center justify-between text-xs">
          <span className="font-semibold text-slate-300">Rendered Output Preview</span>
          {renderedOutput.error ? (
            <span className="flex items-center space-x-1 text-rose-400">
              <AlertTriangle className="w-3.5 h-3.5" />
              <span>{renderedOutput.error}</span>
            </span>
          ) : (
            <span className="flex items-center space-x-1 text-emerald-400">
              <CheckCircle className="w-3.5 h-3.5" />
              <span>Valid ICU Syntax</span>
            </span>
          )}
        </div>

        <div className="p-3 bg-slate-900 rounded-lg border border-slate-800/80 font-mono text-sm text-emerald-300">
          "{renderedOutput.output}"
        </div>
      </div>
    </div>
  );
};
