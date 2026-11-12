import React from 'react';
import { StoreSummary } from '../types';
import { Globe, GitPullRequest, Download, AlertTriangle, CheckCircle, HelpCircle, Sparkles, RefreshCw } from 'lucide-react';

interface HeaderProps {
  summary: StoreSummary | null;
  activeTab: 'studio' | 'matrix' | 'icu' | 'config';
  setActiveTab: (tab: 'studio' | 'matrix' | 'icu' | 'config') => void;
  onExport: () => void;
  onPR: () => void;
  onBatchAutoFix: () => void;
  isExporting: boolean;
  isPRing: boolean;
  isBatchFixing: boolean;
  batchProgress?: { processed: number; total: number; percent: number } | null;
}

export const Header: React.FC<HeaderProps> = ({
  summary,
  activeTab,
  setActiveTab,
  onExport,
  onPR,
  onBatchAutoFix,
  isExporting,
  isPRing,
  isBatchFixing,
  batchProgress,
}) => {
  return (
    <header className="border-b border-slate-800 bg-slate-900/90 backdrop-blur-md sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Brand */}
        <div className="flex items-center space-x-3">
          <div className="w-9 h-9 rounded-lg bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
            <Globe className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center space-x-2">
              <span className="font-semibold text-white text-base tracking-tight">localecraft</span>
              <span className="text-[10px] font-mono uppercase bg-emerald-500/10 text-emerald-400 px-1.5 py-0.5 rounded border border-emerald-500/20 font-medium">v1.0</span>
            </div>
            <p className="text-xs text-slate-400">Visual Localization Sentinel</p>
          </div>
        </div>

        {/* Navigation Tabs */}
        <nav className="flex items-center space-x-1 bg-slate-950 p-1 rounded-lg border border-slate-800 text-xs font-medium">
          <button
            onClick={() => setActiveTab('studio')}
            className={`px-3 py-1.5 rounded-md transition-colors ${
              activeTab === 'studio'
                ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Studio Canvas
          </button>
          <button
            onClick={() => setActiveTab('matrix')}
            className={`px-3 py-1.5 rounded-md transition-colors ${
              activeTab === 'matrix'
                ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Locale Matrix
          </button>
          <button
            onClick={() => setActiveTab('icu')}
            className={`px-3 py-1.5 rounded-md transition-colors ${
              activeTab === 'icu'
                ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            ICU Message Sandbox
          </button>
        </nav>

        {/* Audit Stats & Actions */}
        <div className="flex items-center space-x-3">
          {summary && (
            <div className="hidden lg:flex items-center space-x-2 text-xs">
              <span className="flex items-center space-x-1 text-emerald-400 bg-emerald-500/10 px-2 py-1 rounded border border-emerald-500/20">
                <CheckCircle className="w-3.5 h-3.5" />
                <span>{summary.total_approved} Passed</span>
              </span>
              {summary.total_overflows > 0 && (
                <span className="flex items-center space-x-1 text-rose-400 bg-rose-500/10 px-2 py-1 rounded border border-rose-500/20">
                  <AlertTriangle className="w-3.5 h-3.5" />
                  <span>{summary.total_overflows} Overflow</span>
                </span>
              )}
              {summary.total_missing > 0 && (
                <span className="flex items-center space-x-1 text-amber-400 bg-amber-500/10 px-2 py-1 rounded border border-amber-500/20">
                  <HelpCircle className="w-3.5 h-3.5" />
                  <span>{summary.total_missing} Missing</span>
                </span>
              )}
            </div>
          )}

          {summary && summary.total_overflows > 0 && (
            <button
              onClick={onBatchAutoFix}
              disabled={isBatchFixing}
              className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-gradient-to-r from-amber-500 to-rose-500 hover:from-amber-400 hover:to-rose-400 text-white shadow-sm transition disabled:opacity-50"
              title="Automatically reword all overflowing keys across all languages"
            >
              {isBatchFixing ? (
                <>
                  <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                  <span>Fixing {batchProgress ? `${batchProgress.processed}/${batchProgress.total}` : '...'}</span>
                </>
              ) : (
                <>
                  <Sparkles className="w-3.5 h-3.5" />
                  <span>Auto-Fix All Overflows ({summary.total_overflows})</span>
                </>
              )}
            </button>
          )}

          <button
            onClick={onExport}
            disabled={isExporting}
            className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 transition"
          >
            <Download className="w-3.5 h-3.5" />
            <span>{isExporting ? 'Exporting...' : 'Export Files'}</span>
          </button>

          <button
            onClick={onPR}
            disabled={isPRing}
            className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-emerald-600 hover:bg-emerald-500 text-white shadow-sm shadow-emerald-500/20 transition"
          >
            <GitPullRequest className="w-3.5 h-3.5" />
            <span>{isPRing ? 'Opening PR...' : 'Create PR'}</span>
          </button>
        </div>
      </div>
    </header>
  );
};
