import { type ReactNode } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { PluginTestButton } from '@/components/automation/PluginTestButton';
import type { PluginInstance, PluginType } from '@/api/endpoints/plugins';
import { useDeletePluginInstance, usePatchPluginInstance } from '@/api/hooks/usePlugins';

export interface PluginInstanceListProps {
  instances: PluginInstance[];
  pluginType: PluginType;
  basePath: string;
}

export function PluginInstanceList({
  instances,
  pluginType,
  basePath,
}: PluginInstanceListProps): ReactNode {
  const { t } = useTranslation('automation');
  const patchMutation = usePatchPluginInstance();
  const deleteMutation = useDeletePluginInstance();
  const typeLabel = t(`plugins.types.${pluginType}`);

  if (instances.length === 0) {
    return <p className="text-sm text-muted">{t('plugins.empty')}</p>;
  }

  return (
    <div className="space-y-3">
      {instances.map((instance) => (
        <Card key={instance.id} className="space-y-3 p-4">
          <div className="flex flex-wrap items-start justify-between gap-2">
            <div>
              <h3 className="font-medium">{instance.name}</h3>
              <p className="text-xs text-muted">{instance.implementation}</p>
            </div>
            <span
              className={`rounded-full px-2 py-0.5 text-xs ${
                instance.enabled ? 'bg-primary/15 text-primary' : 'bg-muted/20 text-muted'
              }`}
            >
              {instance.enabled ? t('plugins.enabled') : t('plugins.disabled')}
            </span>
          </div>
          <div className="flex flex-wrap gap-2">
            <Link to={`${basePath}/${instance.id}`}>
              <Button variant="secondary" size="sm">
                {t('plugins.edit')}
              </Button>
            </Link>
            <PluginTestButton instanceId={instance.id} />
            <Button
              variant="secondary"
              size="sm"
              onClick={() =>
                patchMutation.mutate({ id: instance.id, body: { enabled: !instance.enabled } })
              }
              disabled={patchMutation.isPending}
            >
              {instance.enabled ? t('plugins.disabled') : t('plugins.enabled')}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                if (window.confirm(t('plugins.deleteConfirm'))) {
                  deleteMutation.mutate(instance.id);
                }
              }}
              disabled={deleteMutation.isPending}
            >
              {t('plugins.delete')}
            </Button>
          </div>
        </Card>
      ))}
      <p className="text-xs text-muted">{t('plugins.listTitle', { type: typeLabel })}</p>
    </div>
  );
}
