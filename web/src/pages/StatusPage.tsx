import { useMemo, type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { useHealth } from '@/api/hooks/useHealth';
import { useVersion } from '@/api/hooks/useVersion';
import { useServerEvents } from '@/api/events';
import { isApiError } from '@/api/ApiError';
import type { HealthResponse } from '@/api/endpoints/system';

type HealthStatusToken = HealthResponse['status'] | 'down';

function normalizeStatus(value: string | undefined): HealthStatusToken {
  switch (value) {
    case 'ok':
    case 'degraded':
    case 'unavailable':
      return value;
    case 'down':
      return 'down';
    default:
      return 'unavailable';
  }
}

function healthLabelKey(status: HealthStatusToken): string {
  switch (status) {
    case 'ok':
      return 'status:health.ok';
    case 'degraded':
      return 'status:health.degraded';
    case 'down':
    case 'unavailable':
    default:
      return 'status:health.unavailable';
  }
}

function healthToneClass(status: HealthStatusToken): string {
  switch (status) {
    case 'ok':
      return 'text-green-600 dark:text-green-400';
    case 'degraded':
      return 'text-amber-600 dark:text-amber-400';
    case 'down':
    case 'unavailable':
    default:
      return 'text-red-600 dark:text-red-400';
  }
}

function formatDateTime(value: string, locale: string): string {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat(locale, {
    dateStyle: 'medium',
    timeStyle: 'medium',
  }).format(parsed);
}

function formatTimeOnly(date: Date, locale: string): string {
  return new Intl.DateTimeFormat(locale, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(date);
}

export default function StatusPage(): ReactNode {
  const { t, i18n } = useTranslation(['status', 'common']);
  const locale = i18n.resolvedLanguage ?? 'en-US';

  const health = useHealth();
  const version = useVersion();
  const events = useServerEvents();

  const lastUpdated = useMemo(
    () =>
      formatTimeOnly(
        health.dataUpdatedAt > 0 ? new Date(health.dataUpdatedAt) : new Date(),
        locale,
      ),
    [health.dataUpdatedAt, locale],
  );

  const renderHealthCard = (): ReactNode => {
    if (health.isPending && !health.data) {
      return (
        <p role="status" className="text-sm text-muted">
          {t('common:states.loading')}
        </p>
      );
    }
    if (health.isError) {
      const localized = isApiError(health.error) ? health.error.t(t) : t('status:errors.load');
      return (
        <p role="alert" className="text-sm text-red-600 dark:text-red-400">
          {localized}
        </p>
      );
    }
    const data = health.data;
    if (!data) {
      return <p className="text-sm text-muted">{t('common:states.empty')}</p>;
    }
    const checkEntries = data.checks ? Object.entries(data.checks) : [];
    return (
      <div className="flex flex-col gap-4 text-sm">
        <dl className="grid grid-cols-[max-content,1fr] gap-x-4 gap-y-2">
          <dt className="text-muted">{t('status:health.label')}</dt>
          <dd
            className={healthToneClass(data.status)}
            data-testid="status-health-value"
            data-status={data.status}
          >
            {t(healthLabelKey(data.status))}
          </dd>
          <dt className="text-muted">{t('status:fields.lastChecked')}</dt>
          <dd>{formatDateTime(data.time, locale)}</dd>
        </dl>
        <section>
          <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted">
            {t('status:checks.title')}
          </h3>
          {checkEntries.length === 0 ? (
            <p className="text-muted">{t('status:checks.empty')}</p>
          ) : (
            <ul className="grid grid-cols-2 gap-2">
              {checkEntries.map(([name, value]) => {
                const status = normalizeStatus(value?.status);
                return (
                  <li
                    key={name}
                    className="flex items-center justify-between gap-2 rounded border border-border px-2 py-1"
                  >
                    <span className="font-mono text-xs">{name}</span>
                    <span className={healthToneClass(status)} data-status={status}>
                      {t(healthLabelKey(status))}
                    </span>
                  </li>
                );
              })}
            </ul>
          )}
        </section>
        <p className="text-xs text-muted">{t('status:lastUpdated', { at: lastUpdated })}</p>
      </div>
    );
  };

  const renderVersionCard = (): ReactNode => {
    if (version.isPending && !version.data) {
      return (
        <p role="status" className="text-sm text-muted">
          {t('common:states.loading')}
        </p>
      );
    }
    if (version.isError) {
      const localized = isApiError(version.error) ? version.error.t(t) : t('status:errors.load');
      return (
        <p role="alert" className="text-sm text-red-600 dark:text-red-400">
          {localized}
        </p>
      );
    }
    const data = version.data;
    if (!data) {
      return <p className="text-sm text-muted">{t('common:states.empty')}</p>;
    }
    return (
      <dl className="grid grid-cols-[max-content,1fr] gap-x-4 gap-y-2 text-sm">
        <dt className="text-muted">{t('status:version.label')}</dt>
        <dd data-testid="status-version-value">{data.version}</dd>
        <dt className="text-muted">{t('status:version.commit')}</dt>
        <dd>
          <code className="font-mono text-xs">{data.commit}</code>
        </dd>
        <dt className="text-muted">{t('status:version.builtAt')}</dt>
        <dd>{formatDateTime(data.builtAt, locale)}</dd>
      </dl>
    );
  };

  const renderEventsCard = (): ReactNode => {
    const connectionLabel = events.connected
      ? t('status:events.connected')
      : events.reconnectAttempts > 0
        ? t('status:events.reconnecting', { attempt: events.reconnectAttempts })
        : t('status:events.disconnected');

    return (
      <div className="flex flex-col gap-2 text-sm">
        <p>
          <span className="text-muted">{t('status:health.label')}: </span>
          <span data-testid="status-events-connection">{connectionLabel}</span>
        </p>
        {events.lastEvent ? (
          <p data-testid="status-events-last">
            {t('status:events.last', {
              name: events.lastEvent.type,
              at: formatTimeOnly(events.lastEvent.receivedAt, locale),
            })}
          </p>
        ) : (
          <p className="text-muted">{t('status:events.none')}</p>
        )}
      </div>
    );
  };

  return (
    <section className="mx-auto flex max-w-2xl flex-col gap-6 p-6">
      <header className="flex flex-col gap-1">
        <h1 className="text-2xl font-semibold">{t('status:title')}</h1>
        <p className="text-sm text-muted">{t('status:subtitle')}</p>
      </header>

      <Card>
        <CardHeader>
          <CardTitle>{t('status:health.label')}</CardTitle>
          <CardDescription>{t('status:subtitle')}</CardDescription>
        </CardHeader>
        <CardContent>{renderHealthCard()}</CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t('status:version.label')}</CardTitle>
          <CardDescription>{t('status:fields.version')}</CardDescription>
        </CardHeader>
        <CardContent>{renderVersionCard()}</CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t('status:events.title')}</CardTitle>
        </CardHeader>
        <CardContent>{renderEventsCard()}</CardContent>
      </Card>
    </section>
  );
}
