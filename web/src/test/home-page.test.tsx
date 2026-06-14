import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import HomePage from '@/pages/HomePage';
import { TestProviders } from './testUtils';

vi.mock('@/api/hooks/useBrowse', () => ({
  useDiscoverHome: () => ({
    data: {
      shelves: [
        {
          id: 'continueWatching',
          titleKey: 'shelves.continueWatching',
          items: [{ id: 1, libraryId: 2, title: 'In Progress', watchState: { positionMs: 1000, completed: false } }],
        },
        {
          id: 'recentlyAdded',
          titleKey: 'shelves.recentlyAdded',
          items: [{ id: 2, libraryId: 2, title: 'New Film', year: 2024 }],
        },
      ],
    },
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  }),
}));

describe('HomePage', () => {
  it('renders discover shelves', () => {
    render(
      <TestProviders>
        <MemoryRouter>
          <HomePage />
        </MemoryRouter>
      </TestProviders>,
    );
    expect(screen.getByRole('heading', { name: /discover/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /continue watching/i })).toBeInTheDocument();
    expect(screen.getByText('In Progress')).toBeInTheDocument();
    expect(screen.getByText('New Film')).toBeInTheDocument();
  });
});
