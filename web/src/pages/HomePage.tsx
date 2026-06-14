import { type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { MediaRow } from '@/components/browse/MediaRow';
import { MediaRowSkeleton } from '@/components/browse/Skeleton';
import { EmptyState } from '@/components/browse/EmptyState';
import { ErrorState } from '@/components/browse/ErrorState';
import { useDiscoverHome } from '@/api/hooks/useBrowse';
import type { PosterCardItem } from '@/components/browse/PosterCard';

export default function HomePage(): ReactNode {
  const { t } = useTranslation('discover');
  const { data, isLoading, error, refetch } = useDiscoverHome();

  return (
    <section className="mx-auto flex max-w-6xl flex-col gap-8 p-6">
      <header className="flex flex-col gap-1">
        <h1 className="text-2xl font-semibold">{t('title')}</h1>
        <p className="text-sm text-muted">{t('subtitle')}</p>
      </header>

      {isLoading && (
        <>
          <MediaRowSkeleton />
          <MediaRowSkeleton />
        </>
      )}

      {error && <ErrorState ns="discover" onRetry={() => void refetch()} />}

      {!isLoading && !error && data?.shelves.length === 0 && (
        <EmptyState titleKey="empty.title" descriptionKey="empty.description" ns="discover" />
      )}

      {data?.shelves.map((shelf) => (
        <MediaRow
          key={shelf.id}
          titleKey={shelf.titleKey}
          items={shelf.items.map((item): PosterCardItem => {
            const card: PosterCardItem = {
              id: item.id,
              libraryId: item.libraryId,
              title: item.title,
            };
            if (item.year !== undefined) card.year = item.year;
            if (item.posterUrl) card.posterUrl = item.posterUrl;
            if (item.watchState) card.watchState = item.watchState;
            return card;
          })}
        />
      ))}
    </section>
  );
}
