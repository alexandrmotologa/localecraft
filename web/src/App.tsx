import React, { useState, useEffect } from 'react';
import { KeyRecord, StoreSummary } from './types';
import {
  fetchKeys,
  fetchSummary,
  translateKey,
  saveTranslation,
  exportCatalogs,
  triggerPullRequest,
} from './api';
import { Header } from './components/Header';
import { KeyList } from './components/KeyList';
import { ComponentSimulator } from './components/ComponentSimulator';
import { LanguageMatrix } from './components/LanguageMatrix';
import { ICUTester } from './components/ICUTester';

export const App: React.FC = () => {
  const [keys, setKeys] = useState<KeyRecord[]>([]);
  const [selectedKey, setSelectedKey] = useState<KeyRecord | null>(null);
  const [summary, setSummary] = useState<StoreSummary | null>(null);
  const [activeTab, setActiveTab] = useState<'studio' | 'matrix' | 'icu' | 'config'>('studio');
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [targetLocale, setTargetLocale] = useState('de');
  const [isRewording, setIsRewording] = useState(false);
  const [rewordingLocales, setRewordingLocales] = useState<Record<string, boolean>>({});
  const [isExporting, setIsExporting] = useState(false);
  const [isPRing, setIsPRing] = useState(false);
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 4000);
  };

  const loadData = async () => {
    try {
      const [fetchedKeys, fetchedSummary] = await Promise.all([
        fetchKeys(search, statusFilter),
        fetchSummary(),
      ]);
      setKeys(fetchedKeys);
      setSummary(fetchedSummary);

      if (fetchedKeys.length > 0 && !selectedKey) {
        setSelectedKey(fetchedKeys[0]);
      } else if (selectedKey) {
        // Refresh currently selected key from latest list
        const updated = fetchedKeys.find((k) => k.id === selectedKey.id);
        if (updated) setSelectedKey(updated);
      }
    } catch (err: any) {
      console.error('Failed to load data:', err);
    }
  };

  useEffect(() => {
    loadData();
  }, [search, statusFilter]);

  // Connect Server-Sent Events (SSE) for real-time reactivity
  useEffect(() => {
    const es = new EventSource('/api/events');
    es.addEventListener('translation_updated', (e: MessageEvent) => {
      try {
        const payload = JSON.parse(e.data);
        setKeys((prev) =>
          prev.map((k) => {
            if (k.key_name === payload.key) {
              return {
                ...k,
                translations: {
                  ...k.translations,
                  [payload.locale]: {
                    ...k.translations[payload.locale],
                    value: payload.translation,
                    status: payload.status,
                    rendered_width_px: payload.measure.rendered_width_px,
                    is_overflow: payload.measure.is_overflow,
                    overflow_px: payload.measure.overflow_px,
                    overflow_pct: payload.measure.overflow_pct,
                  },
                },
              };
            }
            return k;
          })
        );

        setSelectedKey((prev) => {
          if (prev && prev.key_name === payload.key) {
            return {
              ...prev,
              translations: {
                ...prev.translations,
                [payload.locale]: {
                  ...prev.translations[payload.locale],
                  value: payload.translation,
                  status: payload.status,
                  rendered_width_px: payload.measure.rendered_width_px,
                  is_overflow: payload.measure.is_overflow,
                  overflow_px: payload.measure.overflow_px,
                  overflow_pct: payload.measure.overflow_pct,
                },
              },
            };
          }
          return prev;
        });

        fetchSummary().then(setSummary).catch(console.error);
      } catch (err) {
        console.error('SSE parse error:', err);
      }
    });

    return () => {
      es.close();
    };
  }, []);

  const handleReword = async (keyName: string, locale: string) => {
    setIsRewording(true);
    setRewordingLocales((prev) => ({ ...prev, [locale]: true }));
    try {
      const res = await translateKey(keyName, locale);
      showToast(`Autonomous reword accepted: "${res.final_translation}" (${res.iterations_used} iterations)`);
      await loadData();
    } catch (err: any) {
      showToast(`Reword failed: ${err.message}`);
    } finally {
      setIsRewording(false);
      setRewordingLocales((prev) => ({ ...prev, [locale]: false }));
    }
  };

  const handleSaveTranslation = async (keyName: string, locale: string, val: string) => {
    try {
      await saveTranslation(keyName, locale, val);
      showToast(`Saved translation for [${locale}]`);
      await loadData();
    } catch (err: any) {
      showToast(`Failed to save: ${err.message}`);
    }
  };

  const handleExport = async () => {
    setIsExporting(true);
    try {
      const res = await exportCatalogs();
      showToast(`Successfully exported ${res.exported_files.length} localization files.`);
    } catch (err: any) {
      showToast(`Export error: ${err.message}`);
    } finally {
      setIsExporting(false);
    }
  };

  const handlePR = async () => {
    setIsPRing(true);
    try {
      const res = await triggerPullRequest();
      if (res.success) {
        showToast(`Pull request created: ${res.pr_url}`);
      } else {
        showToast('Pull request branch pushed.');
      }
    } catch (err: any) {
      showToast(`PR creation failed: ${err.message}`);
    } finally {
      setIsPRing(false);
    }
  };

  return (
    <div className="min-h-screen flex flex-col bg-slate-950 text-slate-100 font-sans">
      <Header
        summary={summary}
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        onExport={handleExport}
        onPR={handlePR}
        isExporting={isExporting}
        isPRing={isPRing}
      />

      {/* Toast Notification */}
      {toastMessage && (
        <div className="fixed bottom-5 right-5 z-50 bg-emerald-600 text-white text-xs font-medium px-4 py-2.5 rounded-lg shadow-lg shadow-emerald-950/40 border border-emerald-500 animate-fade-in flex items-center space-x-2">
          <span>{toastMessage}</span>
        </div>
      )}

      {/* Main Workspace Body */}
      <main className="flex-1 flex overflow-hidden">
        {/* Left Sidebar: Keys List */}
        <aside className="w-80 md:w-96 flex-shrink-0 h-[calc(100vh-64px)] overflow-hidden">
          <KeyList
            keys={keys}
            selectedKey={selectedKey}
            onSelectKey={setSelectedKey}
            search={search}
            setSearch={setSearch}
            statusFilter={statusFilter}
            setStatusFilter={setStatusFilter}
          />
        </aside>

        {/* Center / Right Content Canvas */}
        <section className="flex-1 h-[calc(100vh-64px)] overflow-y-auto p-6 space-y-6">
          {selectedKey ? (
            <>
              {activeTab === 'studio' && (
                <div className="space-y-6 max-w-5xl">
                  <ComponentSimulator
                    keyRecord={selectedKey}
                    targetLocale={targetLocale}
                    setTargetLocale={setTargetLocale}
                    onReword={handleReword}
                    isRewording={isRewording}
                  />

                  <LanguageMatrix
                    keyRecord={selectedKey}
                    onSaveTranslation={handleSaveTranslation}
                    onReword={handleReword}
                    rewordingLocales={rewordingLocales}
                  />
                </div>
              )}

              {activeTab === 'matrix' && (
                <div className="max-w-5xl">
                  <LanguageMatrix
                    keyRecord={selectedKey}
                    onSaveTranslation={handleSaveTranslation}
                    onReword={handleReword}
                    rewordingLocales={rewordingLocales}
                  />
                </div>
              )}

              {activeTab === 'icu' && (
                <div className="max-w-5xl">
                  <ICUTester keyRecord={selectedKey} />
                </div>
              )}
            </>
          ) : (
            <div className="h-full flex items-center justify-center text-slate-500 text-sm">
              Select a localization key from the sidebar to inspect layout simulation.
            </div>
          )}
        </section>
      </main>
    </div>
  );
};
