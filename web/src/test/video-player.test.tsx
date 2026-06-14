import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '@/i18n';
import { VideoPlayer } from '@/components/VideoPlayer';

describe('VideoPlayer', () => {
  beforeEach(() => {
    vi.spyOn(HTMLMediaElement.prototype, 'play').mockImplementation(() => Promise.resolve());
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders video element for direct play', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <VideoPlayer manifestUrl="/api/v1/streaming/sessions/x/master.m3u8" mode="direct_play" />
      </I18nextProvider>,
    );
    expect(screen.getByTestId('video-player')).toBeInTheDocument();
  });

  it('reports time updates to the callback', () => {
    const onTimeUpdate = vi.fn();
    render(
      <I18nextProvider i18n={i18n}>
        <VideoPlayer
          manifestUrl="/api/v1/streaming/sessions/x/master.m3u8"
          mode="direct_play"
          onTimeUpdate={onTimeUpdate}
        />
      </I18nextProvider>,
    );
    const video = screen.getByTestId('video-player') as HTMLVideoElement;
    Object.defineProperty(video, 'currentTime', { value: 12.5, configurable: true });
    fireEvent.timeUpdate(video);
    expect(onTimeUpdate).toHaveBeenCalledWith(12500);
  });

  it('toggles playback on space key', () => {
    const pauseSpy = vi
      .spyOn(HTMLMediaElement.prototype, 'pause')
      .mockImplementation(() => undefined);
    render(
      <I18nextProvider i18n={i18n}>
        <VideoPlayer manifestUrl="/api/v1/streaming/sessions/x/master.m3u8" mode="direct_play" />
      </I18nextProvider>,
    );
    const video = screen.getByTestId('video-player') as HTMLVideoElement;
    Object.defineProperty(video, 'paused', { value: false, configurable: true });
    fireEvent.keyDown(window, { code: 'Space' });
    expect(pauseSpy).toHaveBeenCalled();
  });
});
