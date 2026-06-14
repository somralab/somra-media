import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  addFavorite,
  getDiscoverHome,
  getMediaDetail,
  listMediaItemsPaginated,
  removeWatchlist,
  searchMedia,
} from '@/api/endpoints/browse';
import i18n from '@/i18n';

const originalFetch = globalThis.fetch;

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('browse endpoints', () => {
  beforeEach(async () => {
    await i18n.changeLanguage('en-US');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it('getDiscoverHome fetches shelves', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      jsonResponse(200, { shelves: [{ id: 'recentlyAdded', titleKey: 'shelves.recentlyAdded', items: [] }] }),
    ) as unknown as typeof fetch;

    const home = await getDiscoverHome();
    expect(home.shelves).toHaveLength(1);
    expect(String((globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]?.[0])).toContain(
      '/discover/home',
    );
  });

  it('searchMedia encodes query params', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(jsonResponse(200, { results: [], query: 'test' })) as unknown as typeof fetch;

    await searchMedia('inception', 10);
    const url = String((globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]?.[0]);
    expect(url).toContain('q=inception');
    expect(url).toContain('limit=10');
  });

  it('listMediaItemsPaginated builds filter query string', async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValue(jsonResponse(200, { items: [], total: 0, offset: 0, limit: 50 })) as unknown as typeof fetch;

    await listMediaItemsPaginated(3, { sort: 'year', genre: 'Action', year: 2010, watchStatus: 'unwatched' });
    const url = String((globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]?.[0]);
    expect(url).toContain('/libraries/3/items');
    expect(url).toContain('sort=year');
    expect(url).toContain('genre=Action');
    expect(url).toContain('year=2010');
    expect(url).toContain('watchStatus=unwatched');
  });

  it('getMediaDetail and watchlist mutations hit expected paths', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(200, { id: 9, title: 'Film', genres: [], cast: [], images: [], isFavorite: false, inWatchlist: false }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }));
    globalThis.fetch = fetchMock as unknown as typeof fetch;

    const detail = await getMediaDetail(9);
    expect(detail.id).toBe(9);
    await addFavorite(9);
    await removeWatchlist(9);
    expect(String(fetchMock.mock.calls[1]?.[0])).toContain('/favorites/9');
    expect(String(fetchMock.mock.calls[2]?.[0])).toContain('/watchlist/9');
  });
});
