import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { I18nextProvider } from 'react-i18next';
import RequestDiscoverPage from '@/pages/RequestDiscoverPage';
import i18n from '@/i18n';
import { TestProviders } from './testUtils';

const mutate = vi.fn();

vi.mock('@/api/hooks/useDiscover', () => ({
  useDiscover: (query: string) => ({
    data:
      query === 'inception'
        ? {
            results: [
              {
                mediaKind: 'movie',
                provider: 'tmdb',
                externalId: '27205',
                title: 'Inception',
                inLibrary: false,
              },
            ],
          }
        : { results: [] },
    isLoading: false,
    isError: false,
  }),
}));

vi.mock('@/api/hooks/useRequests', () => ({
  useCreateRequest: () => ({ mutate, isPending: false, isError: false }),
}));

function renderPage(): ReturnType<typeof render> {
  return render(
    <TestProviders>
      <MemoryRouter>
        <I18nextProvider i18n={i18n}>
          <RequestDiscoverPage />
        </I18nextProvider>
      </MemoryRouter>
    </TestProviders>,
  );
}

describe('<RequestDiscoverPage />', () => {
  beforeEach(() => {
    mutate.mockClear();
  });

  it('renders search form and navigates discover results', async () => {
    await i18n.changeLanguage('en-US');
    renderPage();
    expect(screen.getByRole('heading', { name: /request content/i })).toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText(/search by title/i), {
      target: { value: 'inception' },
    });
    fireEvent.click(screen.getByRole('button', { name: /^search$/i }));

    await waitFor(() => {
      expect(screen.getByText('Inception')).toBeInTheDocument();
    });
  });

  it('opens request modal with quality selector', async () => {
    await i18n.changeLanguage('en-US');
    renderPage();
    fireEvent.change(screen.getByPlaceholderText(/search by title/i), {
      target: { value: 'inception' },
    });
    fireEvent.click(screen.getByRole('button', { name: /^search$/i }));

    await waitFor(() => {
      expect(screen.getByText('Inception')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /^request$/i }));
    expect(screen.getByRole('group', { name: /preferred quality/i })).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /submit request/i }));
    expect(mutate).toHaveBeenCalledWith(
      expect.objectContaining({
        title: 'Inception',
        qualityResolution: '1080p',
      }),
      expect.any(Object),
    );
  });

  it('shows empty results message', async () => {
    await i18n.changeLanguage('en-US');
    renderPage();
    fireEvent.change(screen.getByPlaceholderText(/search by title/i), {
      target: { value: 'nothing' },
    });
    fireEvent.click(screen.getByRole('button', { name: /^search$/i }));
    await waitFor(() => {
      expect(screen.getByText(/no results found/i)).toBeInTheDocument();
    });
  });
});
