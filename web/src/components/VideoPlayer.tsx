import { type ReactNode, useCallback, useEffect, useRef, useState } from 'react';
import Hls from 'hls.js';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/cn';
import { getAccessToken } from '@/stores/auth';
import { resolveStreamUrl, type PlaybackMode } from '@/api/endpoints/streaming';

export interface VideoPlayerProps {
  manifestUrl: string;
  mode: PlaybackMode;
  startPositionMs?: number;
  onTimeUpdate?: (positionMs: number) => void;
  className?: string;
}

function canPlayNativeHls(): boolean {
  const video = document.createElement('video');
  return video.canPlayType('application/vnd.apple.mpegurl') !== '';
}

export function VideoPlayer({
  manifestUrl,
  mode,
  startPositionMs = 0,
  onTimeUpdate,
  className,
}: VideoPlayerProps): ReactNode {
  const { t } = useTranslation('player');
  const videoRef = useRef<HTMLVideoElement>(null);
  const hlsRef = useRef<Hls | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [levels, setLevels] = useState<{ index: number; label: string }[]>([]);
  const [currentLevel, setCurrentLevel] = useState(-1);

  const src = resolveStreamUrl(manifestUrl);

  const attachHls = useCallback(() => {
    const video = videoRef.current;
    if (!video || mode === 'direct_play') return;

    if (canPlayNativeHls()) {
      video.src = src;
      return;
    }
    if (!Hls.isSupported()) {
      setError(t('errors.unsupported'));
      return;
    }

    const token = getAccessToken();
    const hls = new Hls({
      xhrSetup(xhr) {
        if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`);
      },
    });
    hlsRef.current = hls;
    hls.loadSource(src);
    hls.attachMedia(video);
    hls.on(Hls.Events.MANIFEST_PARSED, () => {
      const next = hls.levels.map((level, index) => ({
        index,
        label: level.height ? `${level.height}p` : t('quality.auto'),
      }));
      setLevels(next);
      if (startPositionMs > 0) {
        video.currentTime = startPositionMs / 1000;
      }
      void video.play().catch(() => undefined);
    });
    hls.on(Hls.Events.ERROR, (_, data) => {
      if (data.fatal) setError(t('errors.playback'));
    });
    return () => {
      hls.destroy();
      hlsRef.current = null;
    };
  }, [mode, src, startPositionMs, t]);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    setError(null);
    if (mode === 'direct_play') {
      video.src = src.replace('master.m3u8', 'source');
      if (startPositionMs > 0) video.currentTime = startPositionMs / 1000;
      void video.play().catch(() => undefined);
      return;
    }
    return attachHls();
  }, [attachHls, mode, src, startPositionMs]);

  useEffect(() => {
    const video = videoRef.current;
    if (!video || !onTimeUpdate) return;
    const handler = () => onTimeUpdate(Math.floor(video.currentTime * 1000));
    video.addEventListener('timeupdate', handler);
    return () => video.removeEventListener('timeupdate', handler);
  }, [onTimeUpdate]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const video = videoRef.current;
      if (!video) return;
      if (e.code === 'Space') {
        e.preventDefault();
        if (video.paused) void video.play();
        else video.pause();
      }
      if (e.code === 'ArrowRight') video.currentTime += 10;
      if (e.code === 'ArrowLeft') video.currentTime = Math.max(0, video.currentTime - 10);
      if (e.code === 'KeyF') {
        if (document.fullscreenElement) void document.exitFullscreen();
        else void video.requestFullscreen();
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);

  const onQualityChange = (index: number) => {
    setCurrentLevel(index);
    if (hlsRef.current) {
      hlsRef.current.currentLevel = index;
    }
  };

  if (error) {
    return <p className="text-sm text-red-400">{error}</p>;
  }

  return (
    <div className={cn('flex flex-col gap-2', className)}>
      {/* Captions are supplied via HLS subtitle renditions when available. */}
      {/* eslint-disable-next-line jsx-a11y/media-has-caption */}
      <video
        ref={videoRef}
        className="aspect-video w-full rounded-md bg-black"
        controls
        playsInline
        aria-label={t('quality.label')}
        data-testid="video-player"
      />
      {levels.length > 1 ? (
        <label className="flex items-center gap-2 text-sm">
          <span>{t('quality.label')}</span>
          <select
            className="rounded border border-border bg-surface px-2 py-1"
            value={currentLevel}
            onChange={(e) => onQualityChange(Number(e.target.value))}
          >
            <option value={-1}>{t('quality.auto')}</option>
            {levels.map((l) => (
              <option key={l.index} value={l.index}>
                {l.label}
              </option>
            ))}
          </select>
        </label>
      ) : null}
    </div>
  );
}
