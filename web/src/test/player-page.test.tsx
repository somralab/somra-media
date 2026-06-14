import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import PlayerPage from '@/pages/PlayerPage';
import { TestProviders } from './testUtils';

vi.mock('@/components/VideoPlayer', () => ({
  VideoPlayer: () => <div data-testid="video-player-stub" />,
}));

vi.mock('@/api/hooks/usePlayback', () => ({
  usePlayback: () => ({
    mutate: vi.fn(),
    isPending: false,
    isError: false,
    data: {
      sessionId: 'sess-1',
      mode: 'direct_play',
      manifestUrl: '/api/v1/streaming/sessions/sess-1/master.m3u8',
      expiresAt: '2026-06-14T12:00:00Z',
    },
  }),
  useWatchStateQuery: () => ({ data: { positionMs: 1500, completed: false } }),
  useSaveWatchProgress: () => ({ mutate: vi.fn() }),
}));

describe('PlayerPage', () => {
  it('renders player with playback mode and back link', () => {
    render(
      <TestProviders>
        <MemoryRouter initialEntries={['/libraries/1/items/42/play']}>
          <Routes>
            <Route path="/libraries/:libraryId/items/:itemId/play" element={<PlayerPage />} />
          </Routes>
        </MemoryRouter>
      </TestProviders>,
    );

    expect(screen.getByTestId('video-player-stub')).toBeInTheDocument();
    expect(screen.getByText(/direct play/i)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /libraries/i })).toHaveAttribute(
      'href',
      '/libraries/1',
    );
  });
});
