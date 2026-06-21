import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  createAutomationMonitor,
  createQualityProfile,
  deleteAutomationMonitor,
  getAutomationDownload,
  getAutomationMonitor,
  getQualityProfile,
  listAutomationDownloads,
  listAutomationMonitors,
  listQualityProfiles,
  patchAutomationMonitor,
  patchQualityProfile,
} from '@/api/endpoints/automation';
import i18n from '@/i18n';

const originalFetch = globalThis.fetch;

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('automation endpoints', () => {
  beforeEach(async () => {
    await i18n.changeLanguage('en-US');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it('lists downloads and quality profiles', async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(200, { downloads: [{ id: 1, title: 'x' }] }))
      .mockResolvedValueOnce(jsonResponse(200, { profiles: [{ id: 1, name: 'default' }] }))
      .mockResolvedValueOnce(jsonResponse(200, { monitors: [{ id: 1, title: 'Show' }] }))
      .mockResolvedValueOnce(jsonResponse(200, { id: 1, title: 'x', status: 'queued' }))
      .mockResolvedValueOnce(jsonResponse(200, { id: 1, name: 'default' }))
      .mockResolvedValueOnce(
        jsonResponse(200, { id: 1, title: 'Show' }),
      ) as unknown as typeof fetch;

    const downloads = await listAutomationDownloads();
    expect(downloads.downloads).toHaveLength(1);

    const profiles = await listQualityProfiles();
    expect(profiles.profiles[0]?.name).toBe('default');

    const monitors = await listAutomationMonitors();
    expect(monitors.monitors[0]?.title).toBe('Show');

    await getAutomationDownload(1);
    await getQualityProfile(1);
    await getAutomationMonitor(1);

    const calls = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls.map((c) =>
      String(c[0]),
    );
    expect(calls.some((u) => u.includes('/automation/downloads'))).toBe(true);
    expect(calls.some((u) => u.includes('/automation/quality-profiles'))).toBe(true);
    expect(calls.some((u) => u.includes('/automation/monitors'))).toBe(true);
  });

  it('creates and patches automation resources', async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(201, { id: 2 }))
      .mockResolvedValueOnce(jsonResponse(200, { id: 1, name: '720p' }))
      .mockResolvedValueOnce(jsonResponse(201, { id: 3, title: 'New Show' }))
      .mockResolvedValueOnce(jsonResponse(200, { id: 3, title: 'Renamed' }))
      .mockResolvedValueOnce(new Response(null, { status: 204 })) as unknown as typeof fetch;

    const createdProfile = await createQualityProfile({ name: '720p', spec: '{}' });
    expect(createdProfile.id).toBe(2);

    const patchedProfile = await patchQualityProfile(1, { name: '720p' });
    expect(patchedProfile.name).toBe('720p');

    const createdMonitor = await createAutomationMonitor({
      title: 'New Show',
      externalId: 'ext-1',
      provider: 'tmdb',
    });
    expect(createdMonitor.id).toBe(3);

    const patchedMonitor = await patchAutomationMonitor(3, { title: 'Renamed' });
    expect(patchedMonitor.title).toBe('Renamed');

    await deleteAutomationMonitor(3);

    const calls = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls;
    expect(String(calls[0]?.[1]?.method ?? '')).toBe('POST');
    expect(String(calls[calls.length - 1]?.[1]?.method ?? '')).toBe('DELETE');
  });
});
