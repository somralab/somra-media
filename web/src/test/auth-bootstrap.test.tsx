import { describe, expect, it, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { AuthBootstrap } from '@/components/AuthBootstrap';
import { TestProviders } from './testUtils';
import * as authApi from '@/api/endpoints/auth';
import { useAuthStore } from '@/stores/auth';

vi.mock('@/api/endpoints/auth', () => ({
  refreshSession: vi.fn(),
}));

const mockedRefresh = vi.mocked(authApi.refreshSession);

describe('AuthBootstrap', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.getState().clearSession();
  });

  it('restores session from refresh cookie on mount', async () => {
    mockedRefresh.mockResolvedValue({
      accessToken: 'tok',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
      user: { id: '1', username: 'admin', roles: ['admin'], disabled: false },
    });

    render(
      <TestProviders>
        <AuthBootstrap>
          <div>App content</div>
        </AuthBootstrap>
      </TestProviders>,
    );

    await waitFor(() => {
      expect(screen.getByText('App content')).toBeInTheDocument();
    });
    expect(useAuthStore.getState().accessToken).toBe('tok');
    expect(useAuthStore.getState().user?.username).toBe('admin');
  });

  it('renders children with cleared session when refresh fails', async () => {
    mockedRefresh.mockRejectedValue(new Error('no session'));

    render(
      <TestProviders>
        <AuthBootstrap>
          <div>App content</div>
        </AuthBootstrap>
      </TestProviders>,
    );

    await waitFor(() => {
      expect(screen.getByText('App content')).toBeInTheDocument();
    });
    expect(useAuthStore.getState().accessToken).toBeNull();
  });
});
