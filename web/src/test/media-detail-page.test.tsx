import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import MediaDetailPage from '@/pages/MediaDetailPage';
import { TestProviders } from './testUtils';

const mutateFavorite = vi.fn();
const mutateWatchlist = vi.fn();

const { mockDetail } = vi.hoisted(() => ({
  mockDetail: {
    value: {
      id: 5,
      libraryId: 1,
      title: 'Inception',
      year: 2010,
      overview: 'A dream within a dream.',
      genres: ['Sci-Fi'],
      cast: [{ name: 'Leonardo DiCaprio', role: 'actor', order: 0 }],
      images: [],
      isFavorite: false,
      inWatchlist: true,
      watchState: { positionMs: 60000, completed: false },
    } as Record<string, unknown>,
  },
}));

vi.mock('@/api/hooks/useBrowse', () => ({
  useMediaDetail: () => ({
    data: mockDetail.value,
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  }),
  useFavoriteToggle: () => ({ mutate: mutateFavorite, isPending: false }),
  useWatchlistToggle: () => ({ mutate: mutateWatchlist, isPending: false }),
}));

function renderDetail(): ReturnType<typeof render> {
  return render(
    <TestProviders>
      <MemoryRouter initialEntries={['/libraries/1/items/5']}>
        <Routes>
          <Route path="/libraries/:libraryId/items/:itemId" element={<MediaDetailPage />} />
        </Routes>
      </MemoryRouter>
    </TestProviders>,
  );
}

describe('MediaDetailPage', () => {
  it('renders detail actions and cast', () => {
    mockDetail.value = {
      id: 5,
      libraryId: 1,
      title: 'Inception',
      year: 2010,
      overview: 'A dream within a dream.',
      genres: ['Sci-Fi'],
      cast: [{ name: 'Leonardo DiCaprio', role: 'actor', order: 0 }],
      images: [],
      isFavorite: false,
      inWatchlist: true,
      watchState: { positionMs: 60000, completed: false },
    };

    renderDetail();

    expect(screen.getByRole('heading', { name: 'Inception' })).toBeInTheDocument();
    expect(screen.getByText('Sci-Fi')).toBeInTheDocument();
    expect(screen.getByText(/Leonardo DiCaprio/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /resume/i })).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /favorite/i }));
    expect(mutateFavorite).toHaveBeenCalled();
  });

  it('renders unmatched items when cast and genres are null', () => {
    mockDetail.value = {
      id: 1,
      libraryId: 1,
      title: 'The Revenant',
      year: 2015,
      overview: '',
      genres: null,
      cast: null,
      images: null,
      isFavorite: false,
      inWatchlist: false,
    };

    renderDetail();

    expect(screen.getByRole('heading', { name: 'The Revenant' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /play/i })).toBeInTheDocument();
    expect(screen.queryByText(/cast/i)).not.toBeInTheDocument();
  });
});
