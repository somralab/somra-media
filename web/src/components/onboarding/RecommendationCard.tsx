import { type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { cn } from '@/lib/cn';

interface RecommendationCardProps {
  title: string;
  description?: string;
  applied?: boolean;
  children?: ReactNode;
  className?: string;
}

export function RecommendationCard({
  title,
  description,
  applied = false,
  children,
  className,
}: RecommendationCardProps): ReactNode {
  const { t } = useTranslation('onboarding');

  return (
    <Card className={cn(className)} data-testid="recommendation-card">
      <CardHeader className="flex flex-row items-start justify-between gap-2">
        <div>
          <CardTitle className="text-base">{title}</CardTitle>
          {description ? <CardDescription>{description}</CardDescription> : null}
        </div>
        <span
          className={cn(
            'shrink-0 rounded-full px-2 py-0.5 text-xs font-medium',
            applied ? 'bg-primary/15 text-primary' : 'bg-surface text-muted',
          )}
        >
          {applied ? t('recommendation.applied') : t('recommendation.recommended')}
        </span>
      </CardHeader>
      {children ? <CardContent>{children}</CardContent> : null}
    </Card>
  );
}
