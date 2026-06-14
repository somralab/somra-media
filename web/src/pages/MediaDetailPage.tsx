import { type ReactNode } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { ErrorState } from '@/components/browse/ErrorState';
import { PosterSkeleton } from '@/components/browse/Skeleton';
import { SubtitleSection } from '@/components/subtitles/SubtitleSection';
import { useFavoriteToggle, useMediaDetail, useWatchlistToggle } from '@/api/hooks/useBrowse';

export default function MediaDetailPage(): ReactNode {
  const { t } = useTranslation('detail');
  const params = useParams();
  const navigate = useNavigate();
  const libraryId = Number(params.libraryId ?? 0);
  const itemId = Number(params.itemId ?? 0);
  const { data: detail, isLoading, error, refetch } = useMediaDetail(itemId);
  const favoriteToggle = useFavoriteToggle(itemId, detail?.isFavorite ?? false);
  const watchlistToggle = useWatchlistToggle(itemId, detail?.inWatchlist ?? false);

  if (isLoading) {
    return (
      <section className="p-6">
        <PosterSkeleton />
      </section>
    );
  }

  if (error || !detail) {
    return <ErrorState ns="detail" onRetry={() => void refetch()} />;
  }

  const resumeMs = detail.watchState?.positionMs ?? 0;
  const hasProgress = resumeMs > 0 && !detail.watchState?.completed;
  const cast = detail.cast ?? [];
  const genres = detail.genres ?? [];

  return (
    <section className="mx-auto flex max-w-5xl flex-col gap-6 p-6">
      <Link className="text-sm text-primary" to={`/libraries/${libraryId}`}>
        ← {t('backToLibrary')}
      </Link>

      <div className="grid gap-6 md:grid-cols-[200px_1fr]">
        {detail.posterUrl ? (
          <img
            src={detail.posterUrl}
            alt=""
            className="aspect-[2/3] w-full rounded-lg object-cover shadow-lg"
          />
        ) : (
          <div className="flex aspect-[2/3] items-center justify-center rounded-lg bg-surface text-muted">
            {t('noPoster')}
          </div>
        )}

        <div className="flex flex-col gap-4">
          <header>
            <h1 className="text-3xl font-bold">{detail.title}</h1>
            <div className="mt-1 flex flex-wrap gap-2 text-sm text-muted">
              {detail.year ? <span>{detail.year}</span> : null}
              {detail.contentRating ? <span>{detail.contentRating}</span> : null}
              {genres.map((g) => (
                <span key={g} className="rounded bg-surface px-2 py-0.5">
                  {g}
                </span>
              ))}
            </div>
          </header>

          {detail.overview ? (
            <p className="text-sm leading-relaxed text-text">{detail.overview}</p>
          ) : null}

          <div className="flex flex-wrap gap-2">
            <Button
              onClick={() =>
                void navigate(`/libraries/${libraryId}/items/${itemId}/play`, {
                  state: hasProgress ? { startPositionMs: resumeMs } : undefined,
                })
              }
            >
              {hasProgress ? t('actions.resume') : t('actions.play')}
            </Button>
            <Button
              variant="secondary"
              disabled={favoriteToggle.isPending}
              onClick={() => favoriteToggle.mutate()}
            >
              {detail.isFavorite ? t('actions.unfavorite') : t('actions.favorite')}
            </Button>
            <Button
              variant="secondary"
              disabled={watchlistToggle.isPending}
              onClick={() => watchlistToggle.mutate()}
            >
              {detail.inWatchlist ? t('actions.removeWatchlist') : t('actions.addWatchlist')}
            </Button>
          </div>
        </div>
      </div>

      <SubtitleSection itemId={itemId} />

      {cast.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>{t('cast.title')}</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
              {cast.map((member) => (
                <li key={`${member.name}-${member.role}`} className="text-sm">
                  <span className="font-medium">{member.name}</span>
                  <span className="text-muted"> · {member.role}</span>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}

      {detail.seasons && detail.seasons.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>{t('seasons.title')}</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            {detail.seasons.map((season) => (
              <div key={season.seasonNumber}>
                <h3 className="mb-2 font-medium">
                  {t('seasons.number', { number: season.seasonNumber })}
                </h3>
                <ul className="flex flex-col gap-1 text-sm">
                  {season.episodes.map((ep) => (
                    <li
                      key={ep.id}
                      className="flex items-center justify-between rounded border border-border px-3 py-2"
                    >
                      <span>
                        {t('seasons.episode', { number: ep.episodeNumber })}
                        {ep.title ? ` — ${ep.title}` : ''}
                      </span>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() =>
                          void navigate(`/libraries/${libraryId}/items/${itemId}/play`)
                        }
                      >
                        {t('actions.play')}
                      </Button>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
    </section>
  );
}
