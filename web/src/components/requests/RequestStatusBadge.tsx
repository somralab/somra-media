import { type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/cn';
import type { RequestStatus } from '@/api/endpoints/requests';

const STATUS_STYLES: Record<RequestStatus, string> = {
  pending: 'bg-amber-500/15 text-amber-600',
  approved: 'bg-blue-500/15 text-blue-600',
  rejected: 'bg-danger/15 text-danger',
  completed: 'bg-emerald-500/15 text-emerald-600',
  cancelled: 'bg-muted/20 text-muted',
};

export interface RequestStatusBadgeProps {
  status: RequestStatus;
  className?: string;
}

export function RequestStatusBadge({ status, className }: RequestStatusBadgeProps): ReactNode {
  const { t } = useTranslation('requests');
  return (
    <span
      className={cn(
        'inline-flex rounded-full px-2 py-0.5 text-xs font-medium capitalize',
        STATUS_STYLES[status],
        className,
      )}
    >
      {t(`status.${status}`)}
    </span>
  );
}
