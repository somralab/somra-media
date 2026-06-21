import { type FormEvent, type ReactNode, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import type { PluginInstance, PluginType } from '@/api/endpoints/plugins';
import {
  useCreatePluginInstance,
  usePatchPluginInstance,
  usePluginCatalog,
} from '@/api/hooks/usePlugins';

export interface PluginInstanceFormProps {
  pluginType: PluginType;
  instance?: PluginInstance | undefined;
  onSaved?: (() => void) | undefined;
  onCancel?: (() => void) | undefined;
}

export function PluginInstanceForm({
  pluginType,
  instance,
  onSaved,
  onCancel,
}: PluginInstanceFormProps): ReactNode {
  const { t } = useTranslation('automation');
  const catalogQuery = usePluginCatalog();
  const createMutation = useCreatePluginInstance();
  const patchMutation = usePatchPluginInstance();

  const implementations = useMemo(
    () => (catalogQuery.data?.catalog ?? []).filter((entry) => entry.pluginType === pluginType),
    [catalogQuery.data, pluginType],
  );

  const [name, setName] = useState(instance?.name ?? '');
  const [implementation, setImplementation] = useState(
    instance?.implementation ?? implementations[0]?.implementation ?? '',
  );
  const [configText, setConfigText] = useState(JSON.stringify(instance?.config ?? {}, null, 2));
  const [enabled, setEnabled] = useState(instance?.enabled ?? true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!instance && implementations.length > 0 && !implementation) {
      setImplementation(implementations[0]?.implementation ?? '');
    }
  }, [implementations, instance, implementation]);

  const handleSubmit = (e: FormEvent): void => {
    e.preventDefault();
    setError(null);
    let config: Record<string, unknown>;
    try {
      config = JSON.parse(configText) as Record<string, unknown>;
    } catch {
      setError(t('plugins.config'));
      return;
    }

    if (instance) {
      patchMutation.mutate(
        { id: instance.id, body: { name, config, enabled } },
        { onSuccess: () => onSaved?.() },
      );
      return;
    }

    createMutation.mutate(
      { pluginType, implementation, name, config, enabled },
      { onSuccess: () => onSaved?.() },
    );
  };

  const pending = createMutation.isPending || patchMutation.isPending;

  return (
    <form className="space-y-4" onSubmit={handleSubmit}>
      {!instance ? (
        <label className="block space-y-1 text-sm">
          <span>{t('plugins.implementation')}</span>
          <select
            className="w-full rounded-md border border-border bg-surface px-3 py-2"
            value={implementation}
            onChange={(e) => setImplementation(e.target.value)}
            required
          >
            {implementations.map((entry) => (
              <option key={entry.implementation} value={entry.implementation}>
                {entry.implementation}
              </option>
            ))}
          </select>
        </label>
      ) : null}

      <label className="block space-y-1 text-sm">
        <span>{t('plugins.name')}</span>
        <Input value={name} onChange={(e) => setName(e.target.value)} required />
      </label>

      <label className="block space-y-1 text-sm">
        <span>{t('plugins.config')}</span>
        <textarea
          className="min-h-32 w-full rounded-md border border-border bg-surface px-3 py-2 font-mono text-xs"
          value={configText}
          onChange={(e) => setConfigText(e.target.value)}
          spellCheck={false}
        />
      </label>

      <label className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          checked={enabled}
          onChange={(e) => setEnabled(e.target.checked)}
          className="accent-primary"
        />
        {t('plugins.enabled')}
      </label>

      {error ? (
        <p className="text-destructive text-sm" role="alert">
          {error}
        </p>
      ) : null}

      <div className="flex flex-wrap gap-2">
        <Button type="submit" disabled={pending}>
          {instance ? t('plugins.save') : t('plugins.create')}
        </Button>
        {onCancel ? (
          <Button type="button" variant="secondary" onClick={onCancel}>
            {t('plugins.cancel')}
          </Button>
        ) : null}
      </div>
    </form>
  );
}
