import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import { SubtitleSection } from '@/components/subtitles/SubtitleSection';
import i18n from '@/i18n';
import { TestProviders } from './testUtils';

vi.mock('@/api/endpoints/subtitles', () => ({
  listMediaSubtitles: vi
    .fn()
    .mockResolvedValue([
      { id: 1, mediaItemId: 5, language: 'en', source: 'external', provider: 'mock' },
    ]),
  searchSubtitles: vi
    .fn()
    .mockResolvedValue([
      { provider: 'mock', externalId: '9', language: 'en', releaseName: 'Release', score: 90 },
    ]),
  downloadSubtitle: vi.fn().mockResolvedValue({ id: 2, language: 'en', source: 'external' }),
  uploadSubtitle: vi.fn().mockResolvedValue({ id: 3, language: 'tr', source: 'uploaded' }),
}));

describe('SubtitleSection', () => {
  it('renders subtitle list and opens search modal', async () => {
    await i18n.changeLanguage('en-US');
    render(
      <TestProviders>
        <I18nextProvider i18n={i18n}>
          <SubtitleSection itemId={5} />
        </I18nextProvider>
      </TestProviders>,
    );

    expect(screen.getByTestId('subtitle-section')).toBeInTheDocument();
    expect(await screen.findByText(/en/i)).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /search/i }));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });
});
