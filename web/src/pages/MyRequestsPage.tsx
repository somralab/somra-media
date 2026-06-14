import { type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { RequestsNav } from '@/components/requests/RequestsNav';
import { RequestStatusBadge } from '@/components/requests/RequestStatusBadge';
import { useCancelRequest, useRequestRealtimeSync, useRequests } from '@/api/hooks/useRequests';

export default function MyRequestsPage(): ReactNode {
  const { t } = useTranslation('requests');
  const { t: tc } = useTranslation('common');
  const { connected } = useRequestRealtimeSync();
  const requestsQuery = useRequests();
  const cancelMutation = useCancelRequest();

  const requests = requestsQuery.data?.requests ?? [];

  return (
    <div className="mx-auto max-w-4xl space-y-6 p-6">
      <header className="space-y-2">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h1 className="text-2xl font-semibold">{t('myRequests.title')}</h1>
          <span
            className="text-xs text-muted"
            title={connected ? t('myRequests.liveConnected') : t('myRequests.livePolling')}
          >
            {connected ? t('myRequests.live') : t('myRequests.polling')}
          </span>
        </div>
        <p className="text-sm text-muted">{t('myRequests.subtitle')}</p>
        <RequestsNav />
      </header>

      {requestsQuery.isLoading ? <p className="text-muted">{tc('states.loading')}</p> : null}

      {requestsQuery.isError ? (
        <p className="text-danger" role="alert">
          {tc('states.error')}
        </p>
      ) : null}

      {!requestsQuery.isLoading && requests.length === 0 ? (
        <p className="text-muted">{t('myRequests.empty')}</p>
      ) : null}

      <ul className="space-y-3">
        {requests.map((req) => (
          <li key={req.id}>
            <Card className="flex flex-wrap items-center justify-between gap-4 p-4">
              <div className="flex min-w-0 flex-1 gap-3">
                {req.posterUrl ? (
                  <img
                    src={req.posterUrl}
                    alt=""
                    className="h-16 w-11 shrink-0 rounded object-cover"
                  />
                ) : null}
                <div className="min-w-0">
                  <p className="font-medium">{req.title}</p>
                  <p className="text-xs text-muted">
                    {t(`mediaKind.${req.mediaKind}`)} ·{' '}
                    {t(`quality.options.${req.qualityResolution}`)}
                  </p>
                  <div className="mt-1">
                    <RequestStatusBadge status={req.status} />
                  </div>
                  {req.adminNote ? (
                    <p className="mt-1 text-xs text-muted">{req.adminNote}</p>
                  ) : null}
                </div>
              </div>
              {(req.status === 'pending' || req.status === 'approved') && (
                <Button
                  variant="secondary"
                  size="sm"
                  disabled={cancelMutation.isPending}
                  onClick={() => cancelMutation.mutate(req.id)}
                >
                  {t('myRequests.cancel')}
                </Button>
              )}
            </Card>
          </li>
        ))}
      </ul>
    </div>
  );
}
