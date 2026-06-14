import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  defaultCapabilities,
  getWatchState,
  resolveStreamUrl,
  saveWatchState,
  startPlayback,
  stopPlayback,
} from '@/api/endpoints/streaming';
import i18n from '@/i18n';

const originalFetch = globalThis.fetch;

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('streaming endpoints', () => {
  beforeEach(async () => {
    await i18n.changeLanguage('en-US');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it('defaultCapabilities returns browser-oriented codec lists', () => {
    const caps = defaultCapabilities();
    expect(caps.videoCodecs).toContain('h264');
    expect(caps.audioCodecs).toContain('aac');
    expect(caps.containers).toContain('mp4');
  });

  it('startPlayback posts capabilities and returns session payload', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse(200, {
        sessionId: 'sess-1',
        mode: 'direct_play',
        manifestUrl: '/api/v1/streaming/sessions/sess-1/master.m3u8',
        expiresAt: '2026-06-14T12:00:00Z',
      }),
    );
    globalThis.fetch = fetchMock as unknown as typeof fetch;

    const result = await startPlayback(42, { startPositionMs: 1000 });
    expect(result.sessionId).toBe('sess-1');
    expect(result.mode).toBe('direct_play');

    const init = (fetchMock.mock.calls[0]?.[1] ?? {}) as RequestInit;
    const body = JSON.parse(String(init.body));
    expect(body.startPositionMs).toBe(1000);
    expect(body.capabilities.videoCodecs).toContain('h264');
  });

  it('stopPlayback sends DELETE and accepts empty body', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(new Response(null, { status: 204 })) as unknown as typeof fetch;
    await expect(stopPlayback('sess-1')).resolves.toBeUndefined();
  });

  it('getWatchState and saveWatchState round-trip progress', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(200, { mediaItemId: 7, positionMs: 5000, completed: false }))
      .mockResolvedValueOnce(jsonResponse(200, { mediaItemId: 7, positionMs: 9000, completed: false }));
    globalThis.fetch = fetchMock as unknown as typeof fetch;

    const state = await getWatchState(7);
    expect(state.positionMs).toBe(5000);

    const saved = await saveWatchState(7, { positionMs: 9000, completed: false });
    expect(saved.positionMs).toBe(9000);
  });

  it('resolveStreamUrl keeps absolute URLs and prefixes relative API paths', () => {
    expect(resolveStreamUrl('https://cdn.example/stream.m3u8')).toBe('https://cdn.example/stream.m3u8');
    expect(resolveStreamUrl('/api/v1/streaming/sessions/x/master.m3u8')).toBe(
      '/api/v1/streaming/sessions/x/master.m3u8',
    );
    expect(resolveStreamUrl('streaming/sessions/x/master.m3u8')).toMatch(/\/streaming\/sessions\/x\/master\.m3u8$/);
  });
});
