import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { I18nextProvider } from 'react-i18next';
import AdminRequestsPage from '@/pages/AdminRequestsPage';
import i18n from '@/i18n';
import { TestProviders } from './testUtils';

const approveMutate = vi.fn();
const rejectMutate = vi.fn();
const patchPoliciesMutate = vi.fn();

vi.mock('@/api/hooks/useRequests', () => ({
  useRequests: () => ({
    data: {
      requests: [
        {
          id: 42,
          userId: 'u2',
          mediaKind: 'tv',
          provider: 'tmdb',
          externalId: '99',
          title: 'Pending Show',
          qualityResolution: '720p',
          status: 'pending',
          collisionFlag: false,
          createdAt: '2026-01-01T00:00:00Z',
          updatedAt: '2026-01-01T00:00:00Z',
        },
      ],
    },
    isLoading: false,
    isError: false,
  }),
  useRequestPolicies: () => ({
    data: { autoApproveRoles: ['admin'], userQuotaPerMonth: 10, adminSettings: {} },
    isLoading: false,
  }),
  useApproveRequest: () => ({ mutate: approveMutate, isPending: false }),
  useRejectRequest: () => ({ mutate: rejectMutate, isPending: false }),
  usePatchRequestPolicies: () => ({
    mutate: patchPoliciesMutate,
    isPending: false,
    isSuccess: false,
  }),
  useRequestRealtimeSync: () => ({ connected: false }),
}));

function renderPage(): ReturnType<typeof render> {
  return render(
    <TestProviders>
      <MemoryRouter>
        <I18nextProvider i18n={i18n}>
          <AdminRequestsPage />
        </I18nextProvider>
      </MemoryRouter>
    </TestProviders>,
  );
}

describe('<AdminRequestsPage />', () => {
  beforeEach(() => {
    approveMutate.mockClear();
    rejectMutate.mockClear();
    patchPoliciesMutate.mockClear();
  });

  it('shows pending queue and approves a request', async () => {
    await i18n.changeLanguage('en-US');
    renderPage();
    expect(screen.getByText('Pending Show')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /^approve$/i }));
    expect(approveMutate).toHaveBeenCalledWith({ id: 42 });
  });

  it('switches to policies tab and saves policies', async () => {
    await i18n.changeLanguage('en-US');
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /policies/i }));
    expect(screen.getByText(/monthly quota per user/i)).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByDisplayValue('10')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button', { name: /^save$/i }));
    expect(patchPoliciesMutate).toHaveBeenCalledWith({
      userQuotaPerMonth: 10,
      autoApproveRoles: ['admin'],
    });
  });
});
