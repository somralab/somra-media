import { apiFetch } from '@/api/client';

export type PlaybackMode = 'direct_play' | 'direct_stream' | 'transcode';

export interface ClientCapabilities {
  videoCodecs?: string[];
  audioCodecs?: string[];
  containers?: string[];
  maxBitrate?: number;
  supportsHdr?: boolean;
  supportsHevc?: boolean;
  maxAudioChannels?: number;
}

export interface PlayRequest {
  capabilities?: ClientCapabilities;
  audioStreamIndex?: number | null;
  subtitleStreamIndex?: number | null;
  startPositionMs?: number;
}

export interface PlayResponse {
  sessionId: string;
  mode: PlaybackMode;
  manifestUrl: string;
  expiresAt: string;
  reason?: string;
}

export interface WatchStatePayload {
  positionMs: number;
  completed: boolean;
}

export interface WatchState {
  mediaItemId: number;
  positionMs: number;
  completed: boolean;
}

export function defaultCapabilities(): ClientCapabilities {
  return {
    videoCodecs: ['h264', 'avc1'],
    audioCodecs: ['aac', 'mp4a'],
    containers: ['mp4', 'm4v', 'mov', 'webm'],
    maxBitrate: 20_000_000,
    supportsHdr: false,
    supportsHevc: false,
    maxAudioChannels: 2,
  };
}

export async function startPlayback(itemId: number, body: PlayRequest): Promise<PlayResponse> {
  return apiFetch<PlayResponse>(`/media-items/${itemId}/play`, {
    method: 'POST',
    body: { capabilities: defaultCapabilities(), ...body },
  });
}

export async function stopPlayback(sessionId: string): Promise<void> {
  await apiFetch<void>(`/streaming/sessions/${sessionId}`, { method: 'DELETE' });
}

export async function getWatchState(itemId: number): Promise<WatchState> {
  return apiFetch<WatchState>(`/watch-state/${itemId}`);
}

export async function saveWatchState(itemId: number, payload: WatchStatePayload): Promise<WatchState> {
  return apiFetch<WatchState>(`/watch-state/${itemId}`, { method: 'PUT', body: payload });
}

export function resolveStreamUrl(manifestPath: string): string {
  if (/^https?:\/\//i.test(manifestPath)) return manifestPath;
  const base = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';
  const normalized = manifestPath.startsWith('/') ? manifestPath : `/${manifestPath}`;
  if (normalized.startsWith('/api/')) return normalized;
  return `${base.replace(/\/$/, '')}${normalized}`;
}
