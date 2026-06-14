import { apiClient } from '../client';
import type { components } from '../generated/openapi';

export type ManagedSubtitle = components['schemas']['ManagedSubtitle'];
export type SubtitleSearchResult = components['schemas']['SubtitleSearchResult'];

export async function listMediaSubtitles(itemId: number): Promise<ManagedSubtitle[]> {
  const res = await apiClient.fetch<{ subtitles: ManagedSubtitle[] }>(
    `/media-items/${itemId}/subtitles`,
  );
  return res.subtitles ?? [];
}

export async function searchSubtitles(params: {
  mediaItemId: number;
  language?: string;
  query?: string;
}): Promise<SubtitleSearchResult[]> {
  const res = await apiClient.fetch<{ results: SubtitleSearchResult[] }>('/subtitles/search', {
    method: 'POST',
    body: params,
  });
  return res.results ?? [];
}

export async function downloadSubtitle(body: {
  mediaItemId: number;
  provider: string;
  externalId: string;
  language: string;
}): Promise<ManagedSubtitle> {
  return apiClient.fetch<ManagedSubtitle>('/subtitles/download', { method: 'POST', body });
}

export async function uploadSubtitle(
  mediaItemId: number,
  language: string,
  file: File,
): Promise<ManagedSubtitle> {
  const form = new FormData();
  form.append('mediaItemId', String(mediaItemId));
  form.append('language', language);
  form.append('file', file);
  const token = (await import('@/stores/auth')).getAccessToken();
  const base = apiClient.baseURL();
  const res = await fetch(`${base}/subtitles/upload`, {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: form,
    credentials: 'include',
  });
  if (!res.ok) {
    throw new Error(`upload failed: ${res.status}`);
  }
  return (await res.json()) as ManagedSubtitle;
}
