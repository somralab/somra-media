import { type ReactNode, useEffect, useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { MediaGrid, type ViewMode } from '@/components/browse/MediaGrid';
import type { PosterCardItem } from '@/components/browse/PosterCard';
import { EmptyState } from '@/components/browse/EmptyState';
import { ErrorState } from '@/components/browse/ErrorState';
import { PosterSkeleton } from '@/components/browse/Skeleton';
import {
  useAutoMatch,
  useLibrary,
  useScanHistory,
  useTriggerScan,
} from '@/api/hooks/useLibraries';
import { useBrowseItems } from '@/api/hooks/useBrowse';
import type { BrowseFilters } from '@/api/endpoints/browse';
import { subscribeScanProgress, type ScanProgressEvent } from '@/api/scanEvents';

export default function LibraryDetailPage(): ReactNode {
  const { t } = useTranslation('browse');
  const params = useParams();
  const libraryId = Number(params.id ?? 0);
  const { data: library } = useLibrary(libraryId);
  const { data: scans } = useScanHistory(libraryId);
  const triggerScan = useTriggerScan(libraryId);
  const autoMatch = useAutoMatch(libraryId);
  const [progress, setProgress] = useState<ScanProgressEvent | null>(null);

  const [viewMode, setViewMode] = useState<ViewMode>('grid');
  const [sort, setSort] = useState<'title' | 'year' | 'created_at'>('title');
  const [genre, setGenre] = useState('');
  const [year, setYear] = useState('');
  const [watchStatus, setWatchStatus] = useState<'all' | 'unwatched' | 'in_progress' | 'completed'>('all');

  const filters = useMemo<BrowseFilters>(() => {
    const f: BrowseFilters = { limit: 500, sort, watchStatus };
    if (genre) f.genre = genre;
    if (year) f.year = Number(year);
    return f;
  }, [sort, genre, year, watchStatus]);

  const { data: page, isLoading, error, refetch } = useBrowseItems(libraryId, filters);

  useEffect(() => {
    if (libraryId <= 0) return;
    return subscribeScanProgress(libraryId, setProgress);
  }, [libraryId]);

  useEffect(() => {
    if (progress?.status === 'succeeded') {
      void refetch();
    }
  }, [progress, refetch]);

  const items = useMemo(
    () =>
      page?.items.map((item) => {
        const card: PosterCardItem = {
          id: item.id,
          libraryId: item.libraryId,
          title: item.title || String(item.id),
        };
        if (item.year !== undefined) card.year = item.year;
        if (item.posterUrl) card.posterUrl = item.posterUrl;
        return card;
      }) ?? [],
    [page?.items],
  );

  if (!library) {
    return (
      <section className="p-6">
        <PosterSkeleton />
      </section>
    );
  }

  return (
    <section className="mx-auto flex max-w-6xl flex-col gap-6 p-6">
      <header className="flex flex-col gap-2">
        <Link className="text-sm text-primary" to="/libraries">
          ← {t('nav.libraries', { ns: 'library' })}
        </Link>
        <h1 className="text-2xl font-semibold">{library.name}</h1>
        <p className="text-sm text-muted">{t(`kinds.${library.kind}`, { ns: 'library' })}</p>
      </header>

      <div className="flex flex-wrap gap-2">
        <Button onClick={() => triggerScan.mutate('full')} disabled={triggerScan.isPending}>
          {t('detail.scanFull', { ns: 'library' })}
        </Button>
        <Button
          variant="secondary"
          onClick={() => triggerScan.mutate('incremental')}
          disabled={triggerScan.isPending}
        >
          {t('detail.scanIncremental', { ns: 'library' })}
        </Button>
        <Button variant="secondary" onClick={() => autoMatch.mutate()} disabled={autoMatch.isPending}>
          {t('detail.autoMatch', { ns: 'library' })}
        </Button>
      </div>

      {progress && (
        <Card>
          <CardContent className="pt-4 text-sm">
            {t('detail.scanProgress', {
              ns: 'library',
              done: progress.filesDone,
              total: progress.filesTotal,
            })}{' '}
            · {t(`scanStatus.${progress.status}`, { ns: 'library', defaultValue: progress.status })}
          </CardContent>
        </Card>
      )}

      <div className="flex flex-wrap items-end gap-3 rounded-lg border border-border bg-surface p-4">
        <label className="flex flex-col gap-1 text-sm">
          <span>{t('filters.sort')}</span>
          <select
            className="rounded-md border border-border bg-bg px-3 py-2"
            value={sort}
            onChange={(e) => setSort(e.target.value as 'title' | 'year' | 'created_at')}
          >
            <option value="title">{t('filters.sortTitle')}</option>
            <option value="year">{t('filters.sortYear')}</option>
            <option value="created_at">{t('filters.sortAdded')}</option>
          </select>
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span>{t('filters.genre')}</span>
          <Input value={genre} onChange={(e) => setGenre(e.target.value)} placeholder={t('filters.genrePlaceholder')} />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span>{t('filters.year')}</span>
          <Input value={year} onChange={(e) => setYear(e.target.value)} type="number" placeholder="2024" />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span>{t('filters.watchStatus')}</span>
          <select
            className="rounded-md border border-border bg-bg px-3 py-2"
            value={watchStatus}
            onChange={(e) =>
              setWatchStatus(e.target.value as 'all' | 'unwatched' | 'in_progress' | 'completed')
            }
          >
            <option value="all">{t('filters.watchAll')}</option>
            <option value="unwatched">{t('filters.watchUnwatched')}</option>
            <option value="in_progress">{t('filters.watchInProgress')}</option>
            <option value="completed">{t('filters.watchCompleted')}</option>
          </select>
        </label>
        <div className="flex gap-2">
          <Button
            variant={viewMode === 'grid' ? 'primary' : 'secondary'}
            size="sm"
            onClick={() => setViewMode('grid')}
          >
            {t('view.grid')}
          </Button>
          <Button
            variant={viewMode === 'list' ? 'primary' : 'secondary'}
            size="sm"
            onClick={() => setViewMode('list')}
          >
            {t('view.list')}
          </Button>
        </div>
      </div>

      {isLoading && (
        <div className="grid grid-cols-4 gap-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <PosterSkeleton key={i} />
          ))}
        </div>
      )}

      {error && <ErrorState onRetry={() => void refetch()} />}

      {!isLoading && !error && items.length === 0 && <EmptyState />}

      {!isLoading && items.length > 0 && <MediaGrid items={items} viewMode={viewMode} />}

      <Card>
        <CardHeader>
          <CardTitle>{t('detail.scanHistory', { ns: 'library' })}</CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="flex flex-col gap-1 text-sm">
            {scans?.map((run) => (
              <li key={run.id}>
                #{run.id} · {run.scanType} ·{' '}
                {t(`scanStatus.${run.status}`, { ns: 'library', defaultValue: run.status })} ·{' '}
                {run.filesDone}/{run.filesTotal}
              </li>
            ))}
          </ul>
        </CardContent>
      </Card>

    </section>
  );
}
