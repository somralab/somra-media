import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import LibraryDetailPage from '@/pages/LibraryDetailPage';
import { TestProviders } from './testUtils';

vi.mock('@/api/hooks/useLibraries', () => ({
  useLibrary: () => ({
    data: { id: 1, name: 'Movies', kind: 'movie', paths: ['/media'], watchEnabled: true },
  }),
  useScanHistory: () => ({ data: [{ id: 1, scanType: 'full', status: 'succeeded', filesTotal: 1, filesDone: 1 }] }),
  useMediaItems: () => ({ data: [{ id: 1, title: 'Inception', year: 2010, matchStatus: 'unmatched' }], refetch: vi.fn() }),
  useTriggerScan: () => ({ mutate: vi.fn(), isPending: false }),
  useAutoMatch: () => ({ mutate: vi.fn(), isPending: false }),
  useMatchCandidates: () => ({ data: [] }),
}));

vi.mock('@/api/scanEvents', () => ({
  subscribeScanProgress: () => () => undefined,
}));

describe('LibraryDetailPage', () => {
  it('renders library detail and scan controls', () => {
    const qc = new QueryClient();
    render(
      <QueryClientProvider client={qc}>
        <TestProviders>
          <MemoryRouter initialEntries={['/libraries/1']}>
            <Routes>
              <Route path="/libraries/:id" element={<LibraryDetailPage />} />
            </Routes>
          </MemoryRouter>
        </TestProviders>
      </QueryClientProvider>,
    );
    expect(screen.getByRole('heading', { name: 'Movies' })).toBeInTheDocument();
    expect(screen.getByText(/full scan/i)).toBeInTheDocument();
    expect(screen.getByText('Inception')).toBeInTheDocument();
  });
});
