import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import LibraryPage from '@/pages/LibraryPage';
import { TestProviders } from './testUtils';

vi.mock('@/api/hooks/useLibraries', () => ({
  useLibraries: () => ({ data: [], isLoading: false, error: null }),
  useCreateLibrary: () => ({ mutate: vi.fn(), isPending: false }),
}));

describe('LibraryPage', () => {
  it('renders library management heading', () => {
    const qc = new QueryClient();
    render(
      <QueryClientProvider client={qc}>
        <TestProviders>
          <MemoryRouter>
            <LibraryPage />
          </MemoryRouter>
        </TestProviders>
      </QueryClientProvider>,
    );
    expect(screen.getByRole('heading', { level: 1, name: /libraries/i })).toBeInTheDocument();
    expect(screen.getByText(/new library/i)).toBeInTheDocument();
  });
});
