import i18n from '@/i18n';
import { clearAuthSession, getAccessToken, setAuthSession, useAuthStore } from '@/stores/auth';
import { ApiError, type ErrorEnvelope } from './ApiError';

export interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  signal?: AbortSignal;
  headers?: Record<string, string>;
  body?: unknown;
  credentials?: RequestCredentials;
  _retry?: boolean;
}

const DEFAULT_BASE_URL = '/api/v1';

function resolveBaseURL(): string {
  const fromEnv = import.meta.env.VITE_API_BASE_URL;
  if (typeof fromEnv === 'string' && fromEnv.length > 0) {
    return fromEnv.replace(/\/$/, '');
  }
  return DEFAULT_BASE_URL;
}

/**
 * Build the absolute URL for a path, accepting both `/health` and `health`.
 * The base URL is taken from `import.meta.env.VITE_API_BASE_URL` (default
 * `/api/v1`).
 */
function buildURL(path: string, base = resolveBaseURL()): string {
  if (/^https?:\/\//i.test(path)) return path;
  const normalized = path.startsWith('/') ? path : `/${path}`;
  return `${base}${normalized}`;
}

function looksLikeErrorEnvelope(value: unknown): value is ErrorEnvelope {
  if (typeof value !== 'object' || value === null) return false;
  const v = value as Record<string, unknown>;
  return (
    typeof v.code === 'string' && typeof v.messageKey === 'string' && typeof v.message === 'string'
  );
}

/**
 * Lightweight typed fetch wrapper. The caller supplies the expected
 * response shape via the `T` generic; the runtime always parses JSON for
 * 2xx and never inspects the payload itself (it is up to the typed
 * endpoint wrappers to keep the contract honest).
 */
export async function apiFetch<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', signal, headers = {}, body, credentials, _retry } = options;

  const finalHeaders: Record<string, string> = {
    Accept: 'application/json',
    'Accept-Language': i18n.language || 'en-US',
    ...headers,
  };

  const token = getAccessToken();
  if (token && !finalHeaders.Authorization) {
    finalHeaders.Authorization = `Bearer ${token}`;
  }

  let serializedBody: BodyInit | undefined;
  if (body !== undefined) {
    serializedBody = JSON.stringify(body);
    finalHeaders['Content-Type'] = finalHeaders['Content-Type'] ?? 'application/json';
  }

  const url = buildURL(path);
  let response: Response;
  try {
    response = await fetch(url, {
      method,
      headers: finalHeaders,
      credentials: credentials ?? 'same-origin',
      ...(serializedBody !== undefined ? { body: serializedBody } : {}),
      ...(signal !== undefined ? { signal } : {}),
    });
  } catch (cause) {
    if (cause instanceof DOMException && cause.name === 'AbortError') {
      throw cause;
    }
    throw ApiError.network(
      'api.errors.network',
      'Network request failed. Check your connection and try again.',
      cause,
    );
  }

  const requestId = response.headers.get('X-Request-Id') ?? undefined;

  if (response.status === 204) {
    return undefined as T;
  }

  const contentType = response.headers.get('Content-Type') ?? '';
  const isJson = contentType.toLowerCase().includes('application/json');

  let parsed: unknown = undefined;
  if (isJson) {
    try {
      parsed = await response.json();
    } catch (cause) {
      throw new ApiError({
        status: response.status,
        code: 'invalid_json',
        messageKey: 'api.errors.invalidJson',
        message: 'The server response could not be decoded as JSON.',
        ...(requestId !== undefined ? { requestId } : {}),
        cause,
      });
    }
  }

  if (!response.ok) {
    if (response.status === 401 && !_retry && !path.includes('/auth/')) {
      try {
        const refreshResp = await fetch(buildURL('/auth/refresh'), {
          method: 'POST',
          credentials: 'include',
          headers: { Accept: 'application/json' },
        });
        if (refreshResp.ok) {
          const refreshed = (await refreshResp.json()) as { accessToken: string; expiresAt: string };
          const user = useAuthStore.getState().user;
          if (user) {
            setAuthSession(refreshed.accessToken, refreshed.expiresAt, user);
            return apiFetch<T>(path, { ...options, _retry: true });
          }
        }
        clearAuthSession();
      } catch {
        clearAuthSession();
      }
    }
    if (looksLikeErrorEnvelope(parsed)) {
      const envelope = parsed as ErrorEnvelope;
      if (envelope.requestId === undefined && requestId !== undefined) {
        envelope.requestId = requestId;
      }
      throw ApiError.fromEnvelope(response.status, envelope);
    }
    throw new ApiError({
      status: response.status,
      code: 'http_error',
      messageKey: 'api.errors.http',
      message: `HTTP ${response.status} ${response.statusText || 'error'}`.trim(),
      ...(requestId !== undefined ? { requestId } : {}),
    });
  }

  return parsed as T;
}

export const apiClient = {
  baseURL: resolveBaseURL,
  fetch: apiFetch,
};

export type ApiClient = typeof apiClient;
