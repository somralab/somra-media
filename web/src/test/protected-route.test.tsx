import { describe, expect, it, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { TestProviders } from './testUtils';
import { useAuthStore } from '@/stores/auth';
import * as authApi from '@/api/endpoints/auth';

vi.mock('@/api/endpoints/auth', () => ({
  getSetupStatus: vi.fn(),
}));

const mockedSetup = vi.mocked(authApi.getSetupStatus);

describe('ProtectedRoute', () => {
  it('redirects to wizard when onboarding is incomplete', async () => {
    useAuthStore.setState({
      accessToken: 'token',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
      user: { id: '1', username: 'admin', roles: ['admin'], disabled: false },
    });
    mockedSetup.mockResolvedValue({
      setupRequired: false,
      completed: false,
      phase: 'library',
    });

    render(
      <TestProviders>
        <MemoryRouter initialEntries={['/libraries']}>
          <Routes>
            <Route
              path="/libraries"
              element={
                <ProtectedRoute>
                  <div>Libraries</div>
                </ProtectedRoute>
              }
            />
            <Route path="/setup/wizard" element={<div>Wizard</div>} />
          </Routes>
        </MemoryRouter>
      </TestProviders>,
    );

    await waitFor(() => {
      expect(screen.getByText('Wizard')).toBeInTheDocument();
    });
  });

  it('redirects unauthenticated users to wizard when setup is incomplete', async () => {
    useAuthStore.getState().clearSession();
    mockedSetup.mockResolvedValue({
      setupRequired: true,
      completed: false,
      phase: 'language',
    });

    render(
      <TestProviders>
        <MemoryRouter initialEntries={['/libraries']}>
          <Routes>
            <Route
              path="/libraries"
              element={
                <ProtectedRoute>
                  <div>Libraries</div>
                </ProtectedRoute>
              }
            />
            <Route path="/setup/wizard" element={<div>Wizard</div>} />
            <Route path="/login" element={<div>Login</div>} />
          </Routes>
        </MemoryRouter>
      </TestProviders>,
    );

    await waitFor(() => {
      expect(screen.getByText('Wizard')).toBeInTheDocument();
    });
  });

  it('redirects unauthenticated users to login when onboarding is complete', async () => {
    useAuthStore.getState().clearSession();
    mockedSetup.mockResolvedValue({
      setupRequired: false,
      completed: true,
      phase: 'complete',
    });

    render(
      <TestProviders>
        <MemoryRouter initialEntries={['/libraries']}>
          <Routes>
            <Route
              path="/libraries"
              element={
                <ProtectedRoute>
                  <div>Libraries</div>
                </ProtectedRoute>
              }
            />
            <Route path="/setup/wizard" element={<div>Wizard</div>} />
            <Route path="/login" element={<div>Login</div>} />
          </Routes>
        </MemoryRouter>
      </TestProviders>,
    );

    await waitFor(() => {
      expect(screen.getByText('Login')).toBeInTheDocument();
    });
  });
});
