import { describe, expect, it, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { TestProviders } from './testUtils';
import { useAuthStore } from '@/stores/auth';

vi.mock('@/api/endpoints/auth', () => ({
  getSetupStatus: vi
    .fn()
    .mockResolvedValue({ setupRequired: false, phase: 'library', completed: false }),
}));

describe('ProtectedRoute', () => {
  it('redirects to wizard when onboarding is incomplete', async () => {
    useAuthStore.setState({
      accessToken: 'token',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
      user: { id: '1', username: 'admin', roles: ['admin'], disabled: false },
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
});
