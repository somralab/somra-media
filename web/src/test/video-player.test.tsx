import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '@/i18n';
import { VideoPlayer } from '@/components/VideoPlayer';

describe('VideoPlayer', () => {
  it('renders video element', () => {
    vi.spyOn(HTMLMediaElement.prototype, 'play').mockImplementation(() => Promise.resolve());
    render(
      <I18nextProvider i18n={i18n}>
        <VideoPlayer manifestUrl="/api/v1/streaming/sessions/x/master.m3u8" mode="direct_play" />
      </I18nextProvider>,
    );
    expect(screen.getByTestId('video-player')).toBeInTheDocument();
  });
});
