import { type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';

export interface ErrorStateProps {
  messageKey?: string;
  ns?: string;
  onRetry?: () => void;
}

export function ErrorState({
  messageKey = 'error.loadFailed',
  ns = 'browse',
  onRetry,
}: ErrorStateProps): ReactNode {
  const { t } = useTranslation(ns);

  return (
    <div className="flex flex-col items-center gap-3 py-12 text-center" role="alert">
      <p className="text-sm text-danger">{t(messageKey)}</p>
      {onRetry ? (
        <Button variant="secondary" onClick={onRetry}>
          {t('actions.retry', { ns: 'common' })}
        </Button>
      ) : null}
    </div>
  );
}
