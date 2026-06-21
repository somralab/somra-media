import { type ReactNode, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';
import { useTestPluginInstance } from '@/api/hooks/usePlugins';

export interface PluginTestButtonProps {
  instanceId: number;
}

export function PluginTestButton({ instanceId }: PluginTestButtonProps): ReactNode {
  const { t } = useTranslation('automation');
  const testMutation = useTestPluginInstance();
  const [message, setMessage] = useState<string | null>(null);

  const handleTest = (): void => {
    setMessage(null);
    testMutation.mutate(instanceId, {
      onSuccess: (result) => {
        setMessage(result.success ? t('plugins.testSuccess') : t('plugins.testFailed'));
      },
      onError: () => setMessage(t('plugins.testFailed')),
    });
  };

  return (
    <div className="flex flex-wrap items-center gap-2">
      <Button
        type="button"
        variant="secondary"
        size="sm"
        onClick={handleTest}
        disabled={testMutation.isPending}
      >
        {t('plugins.test')}
      </Button>
      {message ? (
        <span className="text-xs text-muted" role="status">
          {message}
        </span>
      ) : null}
    </div>
  );
}
