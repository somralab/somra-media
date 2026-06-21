import { type FormEvent, type ReactNode, useEffect, useState } from 'react';
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { QualityProfilePicker } from '@/components/automation/QualityProfilePicker';
import {
  useAutomationMonitor,
  useAutomationMonitors,
  useCreateAutomationMonitor,
  useDeleteAutomationMonitor,
  usePatchAutomationMonitor,
  useQualityProfiles,
} from '@/api/hooks/useAutomation';

const BASE = '/settings/automation/monitors';

export default function AutomationMonitorsPage(): ReactNode {
  const { t } = useTranslation('automation');
  const { id } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const isNew = location.pathname.endsWith('/new');
  const monitorId = id ? Number(id) : 0;

  const monitorsQuery = useAutomationMonitors(!monitorId && !isNew);
  const monitorQuery = useAutomationMonitor(monitorId, monitorId > 0);
  const profilesQuery = useQualityProfiles(isNew || monitorId > 0);
  const createMutation = useCreateAutomationMonitor();
  const patchMutation = usePatchAutomationMonitor();
  const deleteMutation = useDeleteAutomationMonitor();

  const [title, setTitle] = useState('');
  const [provider, setProvider] = useState('tmdb');
  const [externalId, setExternalId] = useState('');
  const [qualityProfile, setQualityProfile] = useState('');
  const [enabled, setEnabled] = useState(true);

  useEffect(() => {
    if (monitorQuery.data) {
      setTitle(monitorQuery.data.title ?? '');
      setProvider(monitorQuery.data.provider ?? 'tmdb');
      setExternalId(monitorQuery.data.externalId ?? '');
      setQualityProfile(monitorQuery.data.qualityProfile ?? '');
      setEnabled(Boolean(monitorQuery.data.enabled));
    }
  }, [monitorQuery.data]);

  if (isNew || monitorId > 0) {
    const handleSubmit = (e: FormEvent): void => {
      e.preventDefault();
      if (monitorId > 0) {
        patchMutation.mutate(
          { id: monitorId, body: { title, qualityProfile, enabled } },
          { onSuccess: () => void navigate(BASE) },
        );
        return;
      }
      createMutation.mutate(
        { title, provider, externalId, qualityProfile, enabled },
        { onSuccess: () => void navigate(BASE) },
      );
    };

    return (
      <div className="mx-auto max-w-2xl space-y-6 p-6">
        <header className="space-y-2">
          <Link to={BASE} className="text-sm text-primary hover:underline">
            ← {t('monitors.title')}
          </Link>
          <h1 className="text-2xl font-semibold">
            {isNew ? t('monitors.add') : t('monitors.edit')}
          </h1>
          <p className="text-xs text-muted">{t('monitors.note')}</p>
        </header>
        <Card className="p-4">
          <form className="space-y-4" onSubmit={handleSubmit}>
            <label className="block space-y-1 text-sm">
              <span>{t('monitors.seriesTitle')}</span>
              <Input value={title} onChange={(e) => setTitle(e.target.value)} required />
            </label>
            {isNew ? (
              <>
                <label className="block space-y-1 text-sm">
                  <span>{t('monitors.provider')}</span>
                  <Input value={provider} onChange={(e) => setProvider(e.target.value)} required />
                </label>
                <label className="block space-y-1 text-sm">
                  <span>{t('monitors.externalId')}</span>
                  <Input
                    value={externalId}
                    onChange={(e) => setExternalId(e.target.value)}
                    required
                  />
                </label>
              </>
            ) : null}
            <QualityProfilePicker
              profiles={profilesQuery.data?.profiles ?? []}
              value={qualityProfile}
              onChange={setQualityProfile}
            />
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={enabled}
                onChange={(e) => setEnabled(e.target.checked)}
                className="accent-primary"
              />
              {t('monitors.enabled')}
            </label>
            <div className="flex flex-wrap gap-2">
              <Button type="submit" disabled={createMutation.isPending || patchMutation.isPending}>
                {isNew ? t('monitors.create') : t('monitors.save')}
              </Button>
              {!isNew ? (
                <Button
                  type="button"
                  variant="ghost"
                  onClick={() => {
                    if (window.confirm(t('monitors.deleteConfirm'))) {
                      deleteMutation.mutate(monitorId, {
                        onSuccess: () => void navigate(BASE),
                      });
                    }
                  }}
                  disabled={deleteMutation.isPending}
                >
                  {t('monitors.delete')}
                </Button>
              ) : null}
            </div>
          </form>
        </Card>
      </div>
    );
  }

  const monitors = monitorsQuery.data?.monitors ?? [];

  return (
    <div className="mx-auto max-w-3xl space-y-6 p-6">
      <header className="flex flex-wrap items-center justify-between gap-2">
        <div className="space-y-1">
          <Link to="/settings/automation" className="text-sm text-primary hover:underline">
            ← {t('hub.title')}
          </Link>
          <h1 className="text-2xl font-semibold">{t('monitors.title')}</h1>
          <p className="text-sm text-muted">{t('monitors.subtitle')}</p>
          <p className="text-xs text-muted">{t('monitors.note')}</p>
        </div>
        <Link to={`${BASE}/new`}>
          <Button>{t('monitors.add')}</Button>
        </Link>
      </header>
      {monitors.length === 0 ? (
        <p className="text-sm text-muted">{t('monitors.empty')}</p>
      ) : (
        <div className="space-y-3">
          {monitors.map((monitor) => (
            <Card key={monitor.id} className="space-y-2 p-4">
              <div className="flex flex-wrap items-start justify-between gap-2">
                <div>
                  <p className="font-medium">{monitor.title}</p>
                  <p className="text-xs text-muted">
                    {monitor.provider}:{monitor.externalId}
                  </p>
                </div>
                <span
                  className={`rounded-full px-2 py-0.5 text-xs ${
                    monitor.enabled ? 'bg-primary/15 text-primary' : 'bg-muted/20 text-muted'
                  }`}
                >
                  {monitor.enabled ? t('plugins.enabled') : t('plugins.disabled')}
                </span>
              </div>
              <p className="text-xs text-muted">
                {t('monitors.lastEpisode')}: S{monitor.lastSeason ?? 0}E{monitor.lastEpisode ?? 0}
              </p>
              <Link to={`${BASE}/${monitor.id}`}>
                <Button variant="secondary" size="sm">
                  {t('monitors.edit')}
                </Button>
              </Link>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
