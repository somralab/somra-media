import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { I18nextProvider } from 'react-i18next';
import MyRequestsPage from '@/pages/MyRequestsPage';
import i18n from '@/i18n';
import { TestProviders } from './testUtils';

const cancelMutate = vi.fn();

vi.mock('@/api/hooks/useRequests', () => ({
  useRequests: () => ({
    data: {
      requests: [
        {
          id: 1,
          userId: 'u1',
          mediaKind: 'movie',
          provider: 'tmdb',
          externalId: '1',
          title: 'Test Movie',
          qualityResolution: '1080p',
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
  useCancelRequest: () => ({ mutate: cancelMutate, isPending: false }),
  useRequestRealtimeSync: () => ({ connected: true }),
}));

function renderPage(): ReturnType<typeof render> {
  return render(
    <TestProviders>
      <MemoryRouter>
        <I18nextProvider i18n={i18n}>
          <MyRequestsPage />
        </I18nextProvider>
      </MemoryRouter>
    </TestProviders>,
  );
}

describe('<MyRequestsPage />', () => {
  beforeEach(() => {
    cancelMutate.mockClear();
  });

  it('lists user requests with status badge', async () => {
    await i18n.changeLanguage('en-US');
    renderPage();
    expect(screen.getByRole('heading', { name: /my requests/i })).toBeInTheDocument();
    expect(screen.getByText('Test Movie')).toBeInTheDocument();
    expect(screen.getByText('Pending')).toBeInTheDocument();
    expect(screen.getByText(/live updates/i)).toBeInTheDocument();
  });

  it('cancels a pending request', async () => {
    await i18n.changeLanguage('en-US');
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    expect(cancelMutate).toHaveBeenCalledWith(1);
  });
});
