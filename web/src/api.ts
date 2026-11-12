import { Config, KeyRecord, StoreSummary, MeasureResult, OptimizeResult } from './types';

const BASE_URL = '';

export async function fetchConfig(): Promise<Config> {
  const res = await fetch(`${BASE_URL}/api/config`);
  if (!res.ok) throw new Error(`Failed to fetch config: ${res.statusText}`);
  return res.json();
}

export async function fetchSummary(): Promise<StoreSummary> {
  const res = await fetch(`${BASE_URL}/api/summary`);
  if (!res.ok) throw new Error(`Failed to fetch summary: ${res.statusText}`);
  return res.json();
}

export async function fetchKeys(search = '', status = ''): Promise<KeyRecord[]> {
  const params = new URLSearchParams();
  if (search) params.set('search', search);
  if (status) params.set('status', status);

  const res = await fetch(`${BASE_URL}/api/keys?${params.toString()}`);
  if (!res.ok) throw new Error(`Failed to fetch keys: ${res.statusText}`);
  return res.json();
}

export async function measureText(params: {
  text: string;
  family?: string;
  size_px?: number;
  weight?: number;
  max_width_px?: number;
  padding_x_px?: number;
}): Promise<MeasureResult> {
  const res = await fetch(`${BASE_URL}/api/measure`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  });
  if (!res.ok) throw new Error(`Failed to measure text: ${res.statusText}`);
  return res.json();
}

export async function translateKey(key: string, targetLocale: string): Promise<OptimizeResult> {
  const res = await fetch(`${BASE_URL}/api/translate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ key, target_locale: targetLocale }),
  });
  if (!res.ok) throw new Error(`Translation failed: ${res.statusText}`);
  return res.json();
}

export async function saveTranslation(key: string, locale: string, value: string): Promise<void> {
  const res = await fetch(`${BASE_URL}/api/save-translation`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ key, locale, value }),
  });
  if (!res.ok) throw new Error(`Failed to save translation: ${res.statusText}`);
}

export async function exportCatalogs(): Promise<{ exported_files: string[] }> {
  const res = await fetch(`${BASE_URL}/api/export`, { method: 'POST' });
  if (!res.ok) throw new Error(`Export failed: ${res.statusText}`);
  return res.json();
}

export async function triggerPullRequest(branch?: string, title?: string): Promise<{ success: boolean; pr_url: string }> {
  const res = await fetch(`${BASE_URL}/api/pr`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ branch, title }),
  });
  if (!res.ok) throw new Error(`PR creation failed: ${res.statusText}`);
  return res.json();
}

export async function triggerBatchTranslate(): Promise<{ queued: boolean; tasks: number }> {
  const res = await fetch(`${BASE_URL}/api/batch-translate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
  });
  if (!res.ok) throw new Error(`Batch translate failed: ${res.statusText}`);
  return res.json();
}
