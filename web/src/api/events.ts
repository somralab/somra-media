import { useEffect, useRef, useState } from 'react';

/**
 * Snapshot of a single server-sent event as exposed to consumers. Both the
 * raw `data` payload and the parsed-JSON form are surfaced so callers can
 * pick whichever is convenient without re-parsing.
 */
export interface ServerEvent {
  /** SSE event name. Defaults to `message` when the server does not send one. */
  type: string;
  /** Raw `data:` payload exactly as the server delivered it. */
  data: string;
  /** Parsed JSON view of `data`, or `undefined` if the payload is not JSON. */
  parsed?: unknown;
  /** Wall-clock timestamp of when the event was received by the client. */
  receivedAt: Date;
  /** Optional `id:` echoed by the server, when present. */
  lastEventId?: string;
}

export interface UseServerEventsState {
  connected: boolean;
  reconnectAttempts: number;
  lastEvent: ServerEvent | null;
  error: Event | null;
}

export interface UseServerEventsOptions {
  /** Endpoint path relative to the API base URL. Default: `/events/stream`. */
  path?: string;
  /** Subscribed event names. Default: `['hello', 'message']`. */
  eventNames?: readonly string[];
  /** Disable auto-connect (useful in tests). Default: `true`. */
  enabled?: boolean;
}

const DEFAULT_PATH = '/events/stream';
const DEFAULT_EVENT_NAMES = ['hello', 'message'] as const;
const INITIAL_BACKOFF_MS = 500;
const MAX_BACKOFF_MS = 30_000;

function resolveBaseURL(): string {
  const fromEnv = import.meta.env.VITE_API_BASE_URL;
  if (typeof fromEnv === 'string' && fromEnv.length > 0) {
    return fromEnv.replace(/\/$/, '');
  }
  return '/api/v1';
}

function buildURL(path: string): string {
  if (/^https?:\/\//i.test(path)) return path;
  const base = resolveBaseURL();
  const normalized = path.startsWith('/') ? path : `/${path}`;
  return `${base}${normalized}`;
}

function tryParseJSON(value: string): unknown {
  try {
    return JSON.parse(value);
  } catch {
    return undefined;
  }
}

/**
 * Lightweight `EventSource` wrapper with exponential-backoff reconnects
 * (capped at 30s). Tracks connection state, attempt count and the most
 * recent event so the UI can surface realtime activity without a global
 * store.
 *
 * The hook uses the host's `EventSource` implementation by default; tests
 * can swap it via `globalThis.EventSource = MockEventSource` before
 * rendering.
 */
export function useServerEvents(options: UseServerEventsOptions = {}): UseServerEventsState {
  const { path = DEFAULT_PATH, eventNames = DEFAULT_EVENT_NAMES, enabled = true } = options;

  const [state, setState] = useState<UseServerEventsState>({
    connected: false,
    reconnectAttempts: 0,
    lastEvent: null,
    error: null,
  });

  const sourceRef = useRef<EventSource | null>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const cancelledRef = useRef(false);

  useEffect(() => {
    if (!enabled) {
      return;
    }
    if (typeof window === 'undefined' || typeof globalThis.EventSource === 'undefined') {
      return;
    }

    cancelledRef.current = false;
    let attempts = 0;

    const cleanupSource = (): void => {
      const current = sourceRef.current;
      if (current) {
        current.close();
        sourceRef.current = null;
      }
    };

    const scheduleReconnect = (): void => {
      if (cancelledRef.current) return;
      attempts += 1;
      const delay = Math.min(INITIAL_BACKOFF_MS * 2 ** (attempts - 1), MAX_BACKOFF_MS);
      setState((prev) => ({ ...prev, connected: false, reconnectAttempts: attempts }));
      timerRef.current = setTimeout(connect, delay);
    };

    const handleEvent =
      (typeHint: string) =>
      (event: MessageEvent<string>): void => {
        const data = typeof event.data === 'string' ? event.data : '';
        const parsed = data.length > 0 ? tryParseJSON(data) : undefined;
        const next: ServerEvent = {
          type: typeHint,
          data,
          receivedAt: new Date(),
          ...(parsed !== undefined ? { parsed } : {}),
          ...(typeof event.lastEventId === 'string' && event.lastEventId.length > 0
            ? { lastEventId: event.lastEventId }
            : {}),
        };
        setState((prev) => ({ ...prev, lastEvent: next }));
      };

    const connect = (): void => {
      if (cancelledRef.current) return;
      cleanupSource();

      const source = new globalThis.EventSource(buildURL(path));
      sourceRef.current = source;

      source.onopen = (): void => {
        attempts = 0;
        setState((prev) => ({
          ...prev,
          connected: true,
          reconnectAttempts: 0,
          error: null,
        }));
      };

      source.onerror = (event: Event): void => {
        setState((prev) => ({ ...prev, connected: false, error: event }));
        source.close();
        sourceRef.current = null;
        scheduleReconnect();
      };

      source.onmessage = handleEvent('message');
      for (const name of eventNames) {
        source.addEventListener(name, handleEvent(name) as EventListener);
      }
    };

    connect();

    return () => {
      cancelledRef.current = true;
      if (timerRef.current !== null) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
      cleanupSource();
    };
  }, [enabled, path, eventNames]);

  return state;
}
