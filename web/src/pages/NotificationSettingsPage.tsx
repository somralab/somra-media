import { type ReactNode } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { useNotificationPrefs, usePatchNotificationPrefs } from '@/api/hooks/useNotificationPrefs';
import type { NotificationPreference } from '@/api/endpoints/notifications';

const EVENT_LABEL_KEYS: Record<string, string> = {
  'request.created': 'events.requestCreated',
  'request.approved': 'events.requestApproved',
  'request.rejected': 'events.requestRejected',
  'request.completed': 'events.requestCompleted',
};

export default function NotificationSettingsPage(): ReactNode {
  const { t } = useTranslation('notifications');
  const { t: tc } = useTranslation('common');
  const prefsQuery = useNotificationPrefs();
  const patchPrefs = usePatchNotificationPrefs();

  const preferences = prefsQuery.data?.preferences ?? [];

  const handleToggle = (pref: NotificationPreference, enabled: boolean): void => {
    patchPrefs.mutate({
      preferences: [
        {
          id: pref.id,
          eventType: pref.eventType,
          channelId: pref.channelId,
          enabled,
          debounceSeconds: pref.debounceSeconds,
        },
      ],
    });
  };

  const handleDebounceChange = (pref: NotificationPreference, debounceSeconds: number): void => {
    patchPrefs.mutate({
      preferences: [
        {
          id: pref.id,
          eventType: pref.eventType,
          channelId: pref.channelId,
          enabled: pref.enabled,
          debounceSeconds,
        },
      ],
    });
  };

  return (
    <section className="mx-auto flex max-w-2xl flex-col gap-6 p-6">
      <header className="space-y-1">
        <Link to="/settings" className="text-sm text-primary hover:underline">
          {t('backToSettings')}
        </Link>
        <h1 className="text-2xl font-semibold">{t('page.title')}</h1>
        <p className="text-sm text-muted">{t('page.subtitle')}</p>
      </header>

      {prefsQuery.isLoading ? <p className="text-muted">{tc('states.loading')}</p> : null}

      {prefsQuery.isError ? (
        <p className="text-danger" role="alert">
          {t('page.loadFailed')}
        </p>
      ) : null}

      {!prefsQuery.isLoading && preferences.length === 0 ? (
        <p className="text-muted">{t('page.empty')}</p>
      ) : null}

      {preferences.map((pref) => {
        const eventKey = EVENT_LABEL_KEYS[pref.eventType] ?? 'events.unknown';
        return (
          <Card key={pref.id}>
            <CardHeader>
              <CardTitle className="text-base">{t(eventKey)}</CardTitle>
              <CardDescription>
                {t('pref.channelLabel', { channelId: pref.channelId })}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={pref.enabled}
                  onChange={(e) => handleToggle(pref, e.target.checked)}
                  disabled={patchPrefs.isPending}
                />
                {t('pref.enabled')}
              </label>
              <label className="block space-y-1 text-sm">
                <span>{t('pref.debounceLabel')}</span>
                <input
                  type="number"
                  min={0}
                  className="w-full rounded-md border border-border bg-surface px-3 py-2"
                  defaultValue={pref.debounceSeconds}
                  onBlur={(e) => handleDebounceChange(pref, Number(e.target.value))}
                />
                <span className="text-xs text-muted">{t('pref.debounceDescription')}</span>
              </label>
            </CardContent>
          </Card>
        );
      })}

      {patchPrefs.isSuccess ? (
        <p className="text-sm text-emerald-600" role="status">
          {t('page.saved')}
        </p>
      ) : null}

      <Button variant="ghost" asChild>
        <Link to="/settings">{tc('actions.cancel')}</Link>
      </Button>
    </section>
  );
}
