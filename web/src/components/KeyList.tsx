import React from 'react';
import { KeyRecord } from '../types';
import { Search, AlertTriangle, CheckCircle, HelpCircle, FileCode } from 'lucide-react';

interface KeyListProps {
  keys: KeyRecord[];
  selectedKey: KeyRecord | null;
  onSelectKey: (k: KeyRecord) => void;
  search: string;
  setSearch: (s: string) => void;
  statusFilter: string;
  setStatusFilter: (s: string) => void;
}

export const KeyList: React.FC<KeyListProps> = ({
  keys,
  selectedKey,
  onSelectKey,
  search,
  setSearch,
  statusFilter,
  setStatusFilter,
}) => {
  return (
    <div className="flex flex-col h-full border-r border-slate-800 bg-slate-900/40">
      {/* Search & Filter Header */}
      <div className="p-3 border-b border-slate-800 space-y-2">
        <div className="relative">
          <Search className="w-4 h-4 text-slate-500 absolute left-3 top-2.5" />
          <input
            type="text"
            placeholder="Search key or English text..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-3 py-1.5 bg-slate-950 border border-slate-800 rounded-lg text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-emerald-500/50"
          />
        </div>

        {/* Filter Pills */}
        <div className="flex items-center space-x-1 text-[11px] overflow-x-auto pb-1">
          {['', 'OVERFLOW', 'MISSING', 'ACCEPTED', 'ICU_ERROR'].map((st) => {
            const label = st === '' ? 'All' : st.toLowerCase();
            const active = statusFilter === st;
            return (
              <button
                key={st}
                onClick={() => setStatusFilter(st)}
                className={`px-2 py-1 rounded-md capitalize transition ${
                  active
                    ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30'
                    : 'text-slate-400 hover:text-slate-200 bg-slate-950/60 border border-slate-800/80'
                }`}
              >
                {label}
              </button>
            );
          })}
        </div>
      </div>

      {/* Keys List */}
      <div className="flex-1 overflow-y-auto divide-y divide-slate-800/50">
        {keys.length === 0 ? (
          <div className="p-6 text-center text-xs text-slate-500">
            No localization keys match your current filter.
          </div>
        ) : (
          keys.map((k) => {
            const isSelected = selectedKey?.id === k.id;
            const hasOverflow = Object.values(k.translations).some((t) => t.status === 'OVERFLOW');
            const hasMissing = Object.values(k.translations).some((t) => t.status === 'MISSING');
            const hasICU = k.is_icu;

            return (
              <button
                key={k.id}
                onClick={() => onSelectKey(k)}
                className={`w-full text-left p-3 transition flex flex-col space-y-1.5 ${
                  isSelected
                    ? 'bg-emerald-500/10 border-l-2 border-emerald-500 text-slate-100'
                    : 'hover:bg-slate-800/40 text-slate-300'
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="font-mono text-xs font-medium text-slate-200 truncate max-w-[200px]">
                    {k.key_name}
                  </span>
                  <div className="flex items-center space-x-1">
                    {hasICU && (
                      <span className="text-[10px] bg-sky-500/10 text-sky-400 border border-sky-500/20 px-1 rounded flex items-center space-x-0.5">
                        <FileCode className="w-2.5 h-2.5" />
                        <span>ICU</span>
                      </span>
                    )}
                    {hasOverflow ? (
                      <AlertTriangle className="w-3.5 h-3.5 text-rose-400" />
                    ) : hasMissing ? (
                      <HelpCircle className="w-3.5 h-3.5 text-amber-400" />
                    ) : (
                      <CheckCircle className="w-3.5 h-3.5 text-emerald-400" />
                    )}
                  </div>
                </div>

                <p className="text-xs text-slate-400 truncate">
                  "{k.source_value}"
                </p>

                <div className="flex items-center justify-between text-[11px] text-slate-500">
                  <span>Limit: {k.max_width_px > 0 ? `${k.max_width_px}px` : 'Unbounded'}</span>
                  <div className="flex items-center space-x-1">
                    {Object.entries(k.translations).map(([loc, trans]) => {
                      let dotColor = 'bg-emerald-400';
                      if (trans.status === 'OVERFLOW') dotColor = 'bg-rose-400 animate-ping';
                      if (trans.status === 'MISSING') dotColor = 'bg-slate-600';
                      if (trans.status === 'ICU_ERROR') dotColor = 'bg-amber-400';
                      return (
                        <span key={loc} className="inline-flex items-center text-[10px] font-mono text-slate-400">
                          <span className={`w-1.5 h-1.5 rounded-full mr-0.5 ${dotColor}`} />
                          {loc}
                        </span>
                      );
                    })}
                  </div>
                </div>
              </button>
            );
          })
        )}
      </div>
    </div>
  );
};
