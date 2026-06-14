import { type ReactNode, useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import {
  useAutoMatch,
  useLibrary,
  useMatchCandidates,
  useMediaItems,
  useScanHistory,
  useTriggerScan,
} from '@/api/hooks/useLibraries';
import { subscribeScanProgress, type ScanProgressEvent } from '@/api/scanEvents';

export default function LibraryDetailPage(): ReactNode {
  const { t } = useTranslation('library');
  const params = useParams();
  const navigate = useNavigate();
  const libraryId = Number(params.id ?? 0);
  const { data: library } = useLibrary(libraryId);
  const { data: scans } = useScanHistory(libraryId);
  const { data: items, refetch: refetchItems } = useMediaItems(libraryId);
  const triggerScan = useTriggerScan(libraryId);
  const autoMatch = useAutoMatch(libraryId);
  const [progress, setProgress] = useState<ScanProgressEvent | null>(null);
  const [selectedItem, setSelectedItem] = useState<number | null>(null);
  const { data: candidates } = useMatchCandidates(selectedItem ?? 0);

  useEffect(() => {
    if (libraryId <= 0) return;
    return subscribeScanProgress(libraryId, setProgress);
  }, [libraryId]);

  useEffect(() => {
    if (progress?.status === 'succeeded') {
      void refetchItems();
    }
  }, [progress, refetchItems]);

  if (!library) {
    return <p className="p-6 text-muted">{t('states.loading', { ns: 'common' })}</p>;
  }

  return (
    <section className="mx-auto flex max-w-4xl flex-col gap-6 p-6">
      <header className="flex flex-col gap-2">
        <Link className="text-sm text-primary" to="/libraries">
          ← {t('nav.libraries')}
        </Link>
        <h1 className="text-2xl font-semibold">{library.name}</h1>
        <p className="text-sm text-muted">{t(`kinds.${library.kind}`)}</p>
      </header>

      <div className="flex flex-wrap gap-2">
        <Button onClick={() => triggerScan.mutate('full')} disabled={triggerScan.isPending}>
          {t('detail.scanFull')}
        </Button>
        <Button
          variant="secondary"
          onClick={() => triggerScan.mutate('incremental')}
          disabled={triggerScan.isPending}
        >
          {t('detail.scanIncremental')}
        </Button>
        <Button
          variant="secondary"
          onClick={() => autoMatch.mutate()}
          disabled={autoMatch.isPending}
        >
          {t('detail.autoMatch')}
        </Button>
      </div>

      {progress && (
        <Card>
          <CardContent className="pt-4 text-sm">
            {t('detail.scanProgress', { done: progress.filesDone, total: progress.filesTotal })} ·{' '}
            {t(`scanStatus.${progress.status}`, { defaultValue: progress.status })}
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>{t('detail.scanHistory')}</CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="flex flex-col gap-1 text-sm">
            {scans?.map((run) => (
              <li key={run.id}>
                #{run.id} · {run.scanType} ·{' '}
                {t(`scanStatus.${run.status}`, { defaultValue: run.status })} · {run.filesDone}/
                {run.filesTotal}
              </li>
            ))}
          </ul>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t('detail.mediaItems')}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 sm:grid-cols-2">
            {items?.map((item) => (
              <button
                key={item.id}
                type="button"
                className="flex gap-3 rounded-md border border-border p-3 text-left hover:bg-surface"
                onClick={() => setSelectedItem(item.id)}
              >
                {item.posterUrl ? (
                  <img src={item.posterUrl} alt="" className="h-20 w-14 rounded object-cover" />
                ) : (
                  <div className="flex h-20 w-14 items-center justify-center rounded bg-surface text-xs text-muted">
                    {t('detail.noPoster')}
                  </div>
                )}
                <div>
                  <div className="font-medium">{item.title || item.id}</div>
                  {item.year && <div className="text-xs text-muted">{item.year}</div>}
                  <div className="text-xs text-muted">
                    {t(`matchStatus.${item.matchStatus}`, { defaultValue: item.matchStatus })}
                  </div>
                  <Button
                    className="mt-2"
                    size="sm"
                    onClick={(e) => {
                      e.stopPropagation();
                      void navigate(`/libraries/${libraryId}/items/${item.id}/play`);
                    }}
                  >
                    {t('play.button', { ns: 'streaming' })}
                  </Button>
                </div>
              </button>
            ))}
          </div>
        </CardContent>
      </Card>

      {selectedItem && candidates && candidates.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>{t('detail.rematch')}</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="flex flex-col gap-2 text-sm">
              {candidates.slice(0, 5).map((c) => (
                <li key={`${c.provider}-${c.externalId}`}>
                  {c.title} ({c.provider}) — score {c.score.toFixed(2)}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}
    </section>
  );
}
