import { type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';

export interface EmptyStateProps {
  titleKey?: string;
  descriptionKey?: string;
  ns?: string;
  onAction?: () => void;
  actionKey?: string;
}

export function EmptyState({
  titleKey = 'empty.title',
  descriptionKey = 'empty.description',
  ns = 'browse',
  onAction,
  actionKey,
}: EmptyStateProps): ReactNode {
  const { t } = useTranslation(ns);

  return (
    <div className="flex flex-col items-center gap-3 py-12 text-center">
      <p className="text-lg font-medium text-text">{t(titleKey)}</p>
      <p className="max-w-md text-sm text-muted">{t(descriptionKey)}</p>
      {onAction && actionKey ? (
        <Button variant="secondary" onClick={onAction}>
          {t(actionKey)}
        </Button>
      ) : null}
    </div>
  );
}
