import { apiClient } from '../client';

export type LibraryKind = 'movie' | 'tv' | 'music';

export interface Library {
  id: number;
  name: string;
  kind: LibraryKind;
  paths: string[];
  watchEnabled: boolean;
}

export interface LibraryInput {
  name: string;
  kind: LibraryKind;
  paths: string[];
  watchEnabled?: boolean;
}

export interface ScanResponse {
  scanRunId: number;
  taskId: string;
}

export interface ScanRun {
  id: number;
  libraryId: number;
  taskId?: string;
  scanType: string;
  status: string;
  filesTotal: number;
  filesDone: number;
  error?: string;
}

export interface MediaItem {
  id: number;
  libraryId: number;
  kind: string;
  title: string;
  year?: number;
  overview?: string;
  posterUrl?: string;
  matchStatus: string;
  matchScore?: number;
}

export interface MatchCandidate {
  provider: string;
  externalId: string;
  title: string;
  year?: number;
  score: number;
  posterUrl?: string;
}

export function listLibraries(signal?: AbortSignal): Promise<Library[]> {
  return apiClient.fetch<Library[]>('/libraries', signal !== undefined ? { signal } : {});
}

export function createLibrary(body: LibraryInput): Promise<Library> {
  return apiClient.fetch<Library>('/libraries', { method: 'POST', body });
}

export function getLibrary(id: number, signal?: AbortSignal): Promise<Library> {
  return apiClient.fetch<Library>(`/libraries/${id}`, signal !== undefined ? { signal } : {});
}

export function updateLibrary(id: number, body: LibraryInput): Promise<Library> {
  return apiClient.fetch<Library>(`/libraries/${id}`, { method: 'PUT', body });
}

export function deleteLibrary(id: number): Promise<void> {
  return apiClient.fetch<void>(`/libraries/${id}`, { method: 'DELETE' });
}

export function triggerScan(
  id: number,
  type: 'full' | 'incremental' = 'full',
): Promise<ScanResponse> {
  return apiClient.fetch<ScanResponse>(`/libraries/${id}/scan`, {
    method: 'POST',
    body: { type },
  });
}

export function listScans(id: number, signal?: AbortSignal): Promise<ScanRun[]> {
  return apiClient.fetch<ScanRun[]>(
    `/libraries/${id}/scans`,
    signal !== undefined ? { signal } : {},
  );
}

export function listMediaItems(libraryId: number, signal?: AbortSignal): Promise<MediaItem[]> {
  return listMediaItemsPaginated(libraryId, { limit: 500 }, signal).then((page) => page.items);
}

async function listMediaItemsPaginated(
  libraryId: number,
  filters: { limit?: number },
  signal?: AbortSignal,
): Promise<{ items: MediaItem[] }> {
  const params = new URLSearchParams();
  if (filters.limit !== undefined) params.set('limit', String(filters.limit));
  const qs = params.toString();
  return apiClient.fetch<{ items: MediaItem[] }>(
    `/libraries/${libraryId}/items${qs ? `?${qs}` : ''}`,
    signal !== undefined ? { signal } : {},
  );
}

export function listMatchCandidates(
  itemId: number,
  signal?: AbortSignal,
): Promise<MatchCandidate[]> {
  return apiClient.fetch<MatchCandidate[]>(
    `/media-items/${itemId}/match-candidates`,
    signal !== undefined ? { signal } : {},
  );
}

export function rematchItem(itemId: number, provider: string, externalId: string): Promise<void> {
  return apiClient.fetch<void>(`/media-items/${itemId}/rematch`, {
    method: 'POST',
    body: { provider, externalId },
  });
}

export function autoMatchLibrary(libraryId: number): Promise<{ matched: number }> {
  return apiClient.fetch<{ matched: number }>(`/libraries/${libraryId}/match`, { method: 'POST' });
}
