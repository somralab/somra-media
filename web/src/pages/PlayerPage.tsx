import { type ReactNode, useEffect, useMemo, useRef } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { VideoPlayer } from '@/components/VideoPlayer';
import { usePlayback, useSaveWatchProgress, useWatchStateQuery } from '@/api/hooks/usePlayback';

const PROGRESS_DEBOUNCE_MS = 5000;

export default function PlayerPage(): ReactNode {
  const { t } = useTranslation(['player', 'streaming', 'library']);
  const params = useParams();
  const libraryId = Number(params.libraryId ?? 0);
  const itemId = Number(params.itemId ?? 0);
  const lastSave = useRef(0);

  const watchState = useWatchStateQuery(itemId);
  const playback = usePlayback(itemId);
  const saveProgress = useSaveWatchProgress(itemId);

  const startPositionMs = watchState.data?.positionMs ?? 0;

  useEffect(() => {
    if (itemId <= 0 || playback.isPending || playback.data) return;
    playback.mutate({ startPositionMs });
  }, [itemId, playback, startPositionMs]);

  const session = playback.data;

  const handleTimeUpdate = useMemo(
    () => (positionMs: number) => {
      const now = Date.now();
      if (now - lastSave.current < PROGRESS_DEBOUNCE_MS) return;
      lastSave.current = now;
      saveProgress.mutate(positionMs);
    },
    [saveProgress],
  );

  if (playback.isError) {
    return (
      <section className="mx-auto max-w-4xl p-6">
        <p className="mb-4 text-red-400">{t('errors.start_failed', { ns: 'streaming' })}</p>
        <Link className="text-sm text-primary" to={`/libraries/${libraryId}`}>
          ← {t('nav.libraries', { ns: 'library' })}
        </Link>
      </section>
    );
  }

  if (!session) {
    return <p className="p-6 text-muted">{t('states.loading', { ns: 'common' })}</p>;
  }

  return (
    <section className="mx-auto flex max-w-4xl flex-col gap-4 p-6">
      <header className="flex items-center justify-between gap-2">
        <Link className="text-sm text-primary" to={`/libraries/${libraryId}`}>
          ← {t('nav.libraries', { ns: 'library' })}
        </Link>
        <span className="text-xs text-muted">
          {t(`mode.${session.mode}`, { ns: 'streaming', defaultValue: session.mode })}
        </span>
      </header>
      <VideoPlayer
        manifestUrl={session.manifestUrl}
        mode={session.mode}
        startPositionMs={startPositionMs}
        onTimeUpdate={handleTimeUpdate}
      />
      <p className="text-xs text-muted">{t('shortcuts.hint')}</p>
    </section>
  );
}
