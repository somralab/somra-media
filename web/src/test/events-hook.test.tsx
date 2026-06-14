import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { act, render, screen } from '@testing-library/react';
import { type ReactElement, useEffect } from 'react';
import { useServerEvents, type UseServerEventsState } from '@/api/events';

type Listener = (event: MessageEvent<string>) => void;

class MockEventSource {
  static instances: MockEventSource[] = [];

  readonly url: string;
  readyState = 0;
  onopen: ((event: Event) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  onmessage: Listener | null = null;
  closed = false;
  private readonly listeners = new Map<string, Set<Listener>>();

  constructor(url: string) {
    this.url = url;
    MockEventSource.instances.push(this);
  }

  addEventListener(type: string, listener: Listener): void {
    let set = this.listeners.get(type);
    if (!set) {
      set = new Set();
      this.listeners.set(type, set);
    }
    set.add(listener);
  }

  removeEventListener(type: string, listener: Listener): void {
    this.listeners.get(type)?.delete(listener);
  }

  dispatchOpen(): void {
    this.readyState = 1;
    this.onopen?.(new Event('open'));
  }

  dispatchNamed(type: string, data: string, lastEventId = ''): void {
    const event = new MessageEvent<string>(type, { data, lastEventId });
    this.listeners.get(type)?.forEach((cb) => cb(event));
    if (type === 'message') {
      this.onmessage?.(event);
    }
  }

  dispatchError(): void {
    this.onerror?.(new Event('error'));
  }

  close(): void {
    this.closed = true;
    this.readyState = 2;
  }
}

let lastState: UseServerEventsState | null = null;

function Probe(): ReactElement {
  const state = useServerEvents({ path: '/events/stream' });
  useEffect(() => {
    lastState = state;
  }, [state]);
  return (
    <div>
      <span data-testid="connected">{String(state.connected)}</span>
      <span data-testid="attempts">{String(state.reconnectAttempts)}</span>
      <span data-testid="last-type">{state.lastEvent?.type ?? ''}</span>
      <span data-testid="last-data">{state.lastEvent?.data ?? ''}</span>
    </div>
  );
}

describe('useServerEvents', () => {
  const originalEventSource = globalThis.EventSource;

  beforeEach(() => {
    MockEventSource.instances = [];
    lastState = null;
    (globalThis as unknown as { EventSource: typeof MockEventSource }).EventSource =
      MockEventSource;
  });

  afterEach(() => {
    (globalThis as unknown as { EventSource: typeof EventSource | undefined }).EventSource =
      originalEventSource;
  });

  it('reports connected after the open event and updates lastEvent on messages', async () => {
    render(<Probe />);

    expect(MockEventSource.instances).toHaveLength(1);
    const source = MockEventSource.instances[0]!;
    expect(source.url).toContain('/events/stream');
    expect(screen.getByTestId('connected').textContent).toBe('false');

    await act(async () => {
      source.dispatchOpen();
    });
    expect(screen.getByTestId('connected').textContent).toBe('true');

    await act(async () => {
      source.dispatchNamed('hello', '{"requestId":"abc"}');
    });
    expect(screen.getByTestId('last-type').textContent).toBe('hello');
    expect(screen.getByTestId('last-data').textContent).toBe('{"requestId":"abc"}');
    expect(lastState?.lastEvent?.parsed).toEqual({ requestId: 'abc' });
  });

  it('marks the connection as disconnected and schedules a reconnect on error', async () => {
    render(<Probe />);
    const source = MockEventSource.instances[0]!;
    await act(async () => {
      source.dispatchOpen();
    });
    expect(screen.getByTestId('connected').textContent).toBe('true');

    await act(async () => {
      source.dispatchError();
    });
    expect(screen.getByTestId('connected').textContent).toBe('false');
    expect(Number(screen.getByTestId('attempts').textContent)).toBeGreaterThanOrEqual(1);
    expect(source.closed).toBe(true);
  });
});
