import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import { PosterCard } from '@/components/browse/PosterCard';
import { MediaRow } from '@/components/browse/MediaRow';
import { MediaGrid } from '@/components/browse/MediaGrid';
import { EmptyState } from '@/components/browse/EmptyState';
import { ErrorState } from '@/components/browse/ErrorState';
import { SearchBar } from '@/components/search/SearchBar';
import { SearchResultsDropdown } from '@/components/search/SearchResultsDropdown';
import { createQueryClient } from '@/lib/queryClient';
import i18n from '@/i18n';

function wrap(ui: React.ReactNode) {
  const qc = createQueryClient();
  return render(
    <QueryClientProvider client={qc}>
      <I18nextProvider i18n={i18n}>
        <MemoryRouter>{ui}</MemoryRouter>
      </I18nextProvider>
    </QueryClientProvider>,
  );
}

describe('browse components', () => {
  it('renders PosterCard with title', () => {
    wrap(
      <PosterCard
        item={{ id: 1, libraryId: 2, title: 'Inception', year: 2010, posterUrl: '' }}
        lazy={false}
      />,
    );
    expect(screen.getByText('Inception')).toBeInTheDocument();
    expect(screen.getByRole('link')).toHaveAttribute('href', '/libraries/2/items/1');
  });

  it('renders MediaRow shelf title', () => {
    wrap(
      <MediaRow
        titleKey="shelves.recentlyAdded"
        items={[{ id: 1, libraryId: 1, title: 'Alpha' }]}
      />,
    );
    expect(screen.getByRole('heading', { name: /recently added/i })).toBeInTheDocument();
  });

  it('renders EmptyState and ErrorState', () => {
    wrap(<EmptyState />);
    expect(screen.getByText(/no titles found/i)).toBeInTheDocument();

    wrap(<ErrorState onRetry={vi.fn()} />);
    expect(screen.getByRole('alert')).toBeInTheDocument();
  });

  it('renders SearchBar with placeholder', () => {
    wrap(<SearchBar value="" onChange={vi.fn()} debounceMs={0} />);
    expect(screen.getByRole('searchbox')).toBeInTheDocument();
  });

  it('renders MediaGrid container', () => {
    const items = Array.from({ length: 8 }, (_, i) => ({
      id: i + 1,
      libraryId: 1,
      title: `Title ${i + 1}`,
    }));
    wrap(<MediaGrid items={items} viewMode="grid" />);
    expect(screen.getByRole('list')).toBeInTheDocument();
  });

  it('returns null when MediaGrid has no items', () => {
    const { container } = wrap(<MediaGrid items={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders SearchResultsDropdown states', () => {
    wrap(<SearchResultsDropdown results={[]} query="" />);
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();

    wrap(<SearchResultsDropdown results={[]} query="missing" isLoading />);
    expect(screen.getByRole('listbox')).toBeInTheDocument();

    wrap(
      <SearchResultsDropdown
        results={[
          {
            id: 1,
            libraryId: 2,
            kind: 'movie',
            title: 'Found',
            year: 2020,
            matchStatus: 'matched',
          },
        ]}
        query="found"
      />,
    );
    expect(screen.getByRole('option', { name: /found/i })).toBeInTheDocument();
  });
});
