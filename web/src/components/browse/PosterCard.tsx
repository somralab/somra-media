import { type ReactNode } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/cn';

export interface PosterCardItem {
  id: number;
  libraryId: number;
  title: string;
  year?: number;
  posterUrl?: string;
  watchState?: { positionMs: number; completed: boolean };
}

export interface PosterCardProps {
  item: PosterCardItem;
  className?: string;
  lazy?: boolean;
}

export function PosterCard({ item, className, lazy = true }: PosterCardProps): ReactNode {
  const { t } = useTranslation('browse');
  const progress =
    item.watchState && item.watchState.positionMs > 0 && !item.watchState.completed
      ? Math.min(100, Math.round((item.watchState.positionMs / 3_600_000) * 100))
      : 0;

  return (
    <Link
      to={`/libraries/${item.libraryId}/items/${item.id}`}
      className={cn(
        'group flex shrink-0 flex-col gap-2 rounded-lg transition-transform hover:scale-[1.02] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary',
        className,
      )}
      aria-label={t('poster.label', { title: item.title })}
    >
      <div className="relative aspect-[2/3] w-full overflow-hidden rounded-md bg-surface shadow-md">
        {item.posterUrl ? (
          <img
            src={item.posterUrl}
            alt=""
            loading={lazy ? 'lazy' : 'eager'}
            className="h-full w-full object-cover"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center p-2 text-center text-xs text-muted">
            {t('poster.noImage')}
          </div>
        )}
        {progress > 0 && (
          <div
            className="absolute inset-x-0 bottom-0 h-1 bg-border"
            role="progressbar"
            aria-valuenow={progress}
            aria-valuemin={0}
            aria-valuemax={100}
          >
            <div className="h-full bg-primary" style={{ width: `${progress}%` }} />
          </div>
        )}
      </div>
      <div className="min-w-0">
        <p className="truncate text-sm font-medium text-text">{item.title}</p>
        {item.year ? <p className="text-xs text-muted">{item.year}</p> : null}
      </div>
    </Link>
  );
}
