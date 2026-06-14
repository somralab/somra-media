import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient, apiFetch } from '@/api/client';
import { ApiError } from '@/api/ApiError';
import { getHealth, getVersion } from '@/api/endpoints/system';
import i18n from '@/i18n';

const originalFetch = globalThis.fetch;

function jsonResponse(
  status: number,
  body: unknown,
  headers: Record<string, string> = {},
): Response {
  return new Response(typeof body === 'string' ? body : JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json', ...headers },
  });
}

describe('apiFetch', () => {
  beforeEach(async () => {
    await i18n.changeLanguage('en-US');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it('parses a 2xx JSON response into the typed payload', async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValue(jsonResponse(200, { ok: true })) as unknown as typeof fetch;

    const result = await apiFetch<{ ok: boolean }>('/health');
    expect(result.ok).toBe(true);
  });

  it('treats a 204 No Content as an empty body', async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValue(new Response(null, { status: 204 })) as unknown as typeof fetch;
    const result = await apiFetch<undefined>('/empty');
    expect(result).toBeUndefined();
  });

  it('sends Accept-Language from i18n', async () => {
    await i18n.changeLanguage('tr-TR');
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(200, {}));
    globalThis.fetch = fetchMock as unknown as typeof fetch;
    await apiFetch('/x');
    const init = (fetchMock.mock.calls[0]?.[1] ?? {}) as RequestInit;
    expect((init.headers as Record<string, string>)['Accept-Language']).toBe('tr-TR');
    await i18n.changeLanguage('en-US');
  });

  it('serializes JSON bodies and sets Content-Type', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(200, {}));
    globalThis.fetch = fetchMock as unknown as typeof fetch;
    await apiFetch('/echo', { method: 'POST', body: { a: 1 } });
    const init = (fetchMock.mock.calls[0]?.[1] ?? {}) as RequestInit;
    expect(init.body).toBe(JSON.stringify({ a: 1 }));
    expect((init.headers as Record<string, string>)['Content-Type']).toBe('application/json');
  });

  it('translates a non-ok JSON error envelope into ApiError', async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValue(
        jsonResponse(
          404,
          { code: 'not_found', messageKey: 'errors.not_found', message: 'gone' },
          { 'X-Request-Id': 'abc' },
        ),
      ) as unknown as typeof fetch;

    await expect(apiFetch('/missing')).rejects.toMatchObject({
      status: 404,
      code: 'not_found',
      requestId: 'abc',
    });
  });

  it('falls back to a generic ApiError when the body is not an envelope', async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValue(jsonResponse(500, { weird: true })) as unknown as typeof fetch;
    await expect(apiFetch('/boom')).rejects.toBeInstanceOf(ApiError);
  });

  it('throws an invalid_json ApiError when the body cannot be decoded', async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValue(
        new Response('{not json', { status: 200, headers: { 'Content-Type': 'application/json' } }),
      ) as unknown as typeof fetch;
    await expect(apiFetch('/bad-json')).rejects.toMatchObject({ code: 'invalid_json' });
  });

  it('wraps fetch network failures with ApiError.network', async () => {
    globalThis.fetch = vi.fn().mockRejectedValue(new TypeError('boom')) as unknown as typeof fetch;
    await expect(apiFetch('/unreachable')).rejects.toMatchObject({ code: 'network_error' });
  });

  it('propagates AbortError without wrapping', async () => {
    const abort = new DOMException('aborted', 'AbortError');
    globalThis.fetch = vi.fn().mockRejectedValue(abort) as unknown as typeof fetch;
    await expect(apiFetch('/cancel')).rejects.toBe(abort);
  });

  it('accepts an absolute URL passthrough', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(200, {}));
    globalThis.fetch = fetchMock as unknown as typeof fetch;
    await apiFetch('https://example.test/x');
    const [url] = fetchMock.mock.calls[0] ?? [];
    expect(url).toBe('https://example.test/x');
  });

  it('exposes resolveBaseURL on the apiClient helper', () => {
    expect(typeof apiClient.baseURL()).toBe('string');
  });
});

describe('endpoints/system', () => {
  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it('getHealth fetches /health', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(200, { status: 'ok' }));
    globalThis.fetch = fetchMock as unknown as typeof fetch;
    await getHealth();
    const [url] = fetchMock.mock.calls[0] ?? [];
    expect(String(url)).toContain('/health');
  });

  it('getVersion fetches /version', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(200, { version: 'x' }));
    globalThis.fetch = fetchMock as unknown as typeof fetch;
    await getVersion();
    const [url] = fetchMock.mock.calls[0] ?? [];
    expect(String(url)).toContain('/version');
  });
});
