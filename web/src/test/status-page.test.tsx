import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { act, render, screen, waitFor } from '@testing-library/react';
import { TestProviders, createTestQueryClient } from './testUtils';
import StatusPage from '@/pages/StatusPage';
import { ApiError } from '@/api/ApiError';
import i18n from '@/i18n';

vi.mock('@/api/endpoints/system', () => ({
  getHealth: vi.fn(),
  getVersion: vi.fn(),
}));

vi.mock('@/api/events', () => ({
  useServerEvents: () => ({
    connected: false,
    reconnectAttempts: 0,
    lastEvent: null,
    error: null,
  }),
}));

import { getHealth, getVersion } from '@/api/endpoints/system';

const getHealthMock = vi.mocked(getHealth);
const getVersionMock = vi.mocked(getVersion);

const healthyPayload = {
  status: 'ok' as const,
  time: '2025-01-01T12:00:00Z',
  checks: { database: { status: 'ok' as const } },
};

const versionPayload = {
  version: '0.1.0-dev',
  commit: '466b6eccd4007f1535820ee1e4f0514ac9581827',
  builtAt: '2025-01-01T11:00:00Z',
};

describe('<StatusPage />', () => {
  beforeEach(async () => {
    await i18n.changeLanguage('en-US');
    getHealthMock.mockReset();
    getVersionMock.mockReset();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('shows a loading state initially', async () => {
    let resolveHealth!: (v: typeof healthyPayload) => void;
    let resolveVersion!: (v: typeof versionPayload) => void;
    getHealthMock.mockImplementation(
      () =>
        new Promise<typeof healthyPayload>((resolve) => {
          resolveHealth = resolve;
        }),
    );
    getVersionMock.mockImplementation(
      () =>
        new Promise<typeof versionPayload>((resolve) => {
          resolveVersion = resolve;
        }),
    );

    render(
      <TestProviders client={createTestQueryClient()}>
        <StatusPage />
      </TestProviders>,
    );

    expect(screen.getAllByRole('status').length).toBeGreaterThan(0);

    await act(async () => {
      resolveHealth(healthyPayload);
      resolveVersion(versionPayload);
    });
    await waitFor(() => {
      expect(screen.getByTestId('status-version-value')).toHaveTextContent('0.1.0-dev');
    });
  });

  it('renders health + version using i18n labels on success (en-US)', async () => {
    getHealthMock.mockResolvedValue(healthyPayload);
    getVersionMock.mockResolvedValue(versionPayload);

    render(
      <TestProviders client={createTestQueryClient()}>
        <StatusPage />
      </TestProviders>,
    );

    await waitFor(() => {
      expect(screen.getByTestId('status-health-value')).toHaveTextContent('Healthy');
    });
    expect(screen.getByTestId('status-version-value')).toHaveTextContent('0.1.0-dev');
    expect(screen.getByText('Dependency checks')).toBeInTheDocument();
    expect(screen.getByText('System status')).toBeInTheDocument();
  });

  it('renders the localized error message (tr-TR) when the API fails', async () => {
    await i18n.changeLanguage('tr-TR');
    getHealthMock.mockRejectedValue(
      new ApiError({
        status: 503,
        code: 'service_unavailable',
        messageKey: 'api.errors.network',
        message: 'Server-localized fallback.',
      }),
    );
    getVersionMock.mockResolvedValue(versionPayload);

    const client = createTestQueryClient();
    client.setQueryDefaults(['system', 'health'], { retry: false });

    render(
      <TestProviders client={client}>
        <StatusPage />
      </TestProviders>,
    );

    await waitFor(
      () => {
        const alerts = screen.queryAllByRole('alert');
        expect(alerts.length).toBeGreaterThan(0);
      },
      { timeout: 4000 },
    );
    const text = screen
      .getAllByRole('alert')
      .map((el) => el.textContent)
      .join('\n');
    expect(text).toContain(
      'Ağ isteği başarısız oldu. Bağlantınızı kontrol edin ve tekrar deneyin.',
    );
    await i18n.changeLanguage('en-US');
  });
});
