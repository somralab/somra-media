import { type ReactNode } from 'react';
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { PluginInstanceForm } from '@/components/automation/PluginInstanceForm';
import { PluginInstanceList } from '@/components/automation/PluginInstanceList';
import { usePluginInstance, usePluginInstancesByType } from '@/api/hooks/usePlugins';
import type { PluginType } from '@/api/endpoints/plugins';

export interface PluginInstancesPageProps {
  pluginType: PluginType;
  basePath: string;
}

export function PluginInstancesPage({ pluginType, basePath }: PluginInstancesPageProps): ReactNode {
  const { t } = useTranslation('automation');
  const { id } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const isNew = location.pathname.endsWith('/new');
  const instanceId = id ? Number(id) : 0;

  const { instances, isLoading } = usePluginInstancesByType(pluginType);
  const instanceQuery = usePluginInstance(instanceId, Boolean(instanceId));

  const typeLabel = t(`plugins.types.${pluginType}`);

  if (isNew || instanceId > 0) {
    return (
      <div className="mx-auto max-w-2xl space-y-6 p-6">
        <header className="space-y-2">
          <Link to={basePath} className="text-sm text-primary hover:underline">
            ← {typeLabel}
          </Link>
          <h1 className="text-2xl font-semibold">{isNew ? t('plugins.add') : t('plugins.edit')}</h1>
        </header>
        <Card className="p-4">
          {instanceId > 0 && instanceQuery.isLoading ? (
            <p className="text-sm text-muted">{t('plugins.edit')}</p>
          ) : (
            <PluginInstanceForm
              pluginType={pluginType}
              instance={isNew ? undefined : instanceQuery.data}
              onSaved={() => void navigate(basePath)}
              onCancel={() => void navigate(basePath)}
            />
          )}
        </Card>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6 p-6">
      <header className="flex flex-wrap items-center justify-between gap-2">
        <div className="space-y-1">
          <Link to="/settings/automation" className="text-sm text-primary hover:underline">
            ← {t('hub.title')}
          </Link>
          <h1 className="text-2xl font-semibold">{t('plugins.listTitle', { type: typeLabel })}</h1>
          <p className="text-sm text-muted">{t('plugins.listSubtitle')}</p>
        </div>
        <Link to={`${basePath}/new`}>
          <Button>{t('plugins.add')}</Button>
        </Link>
      </header>
      {isLoading ? (
        <p className="text-sm text-muted">{t('plugins.listSubtitle')}</p>
      ) : (
        <PluginInstanceList instances={instances} pluginType={pluginType} basePath={basePath} />
      )}
    </div>
  );
}

export default function IndexersPage(): ReactNode {
  return <PluginInstancesPage pluginType="indexer" basePath="/settings/automation/indexers" />;
}

export function IndexerFormPage(): ReactNode {
  return <PluginInstancesPage pluginType="indexer" basePath="/settings/automation/indexers" />;
}

export function DownloadClientsPage(): ReactNode {
  return (
    <PluginInstancesPage
      pluginType="download_client"
      basePath="/settings/automation/download-clients"
    />
  );
}

export function DownloadClientFormPage(): ReactNode {
  return (
    <PluginInstancesPage
      pluginType="download_client"
      basePath="/settings/automation/download-clients"
    />
  );
}
