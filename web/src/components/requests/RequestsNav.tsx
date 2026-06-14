import { type ReactNode } from 'react';
import { NavLink } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/cn';

export interface RequestsNavProps {
  className?: string;
}

export function RequestsNav({ className }: RequestsNavProps): ReactNode {
  const { t } = useTranslation('requests');

  const links: Array<{ to: string; label: string; end?: boolean }> = [
    { to: '/requests/discover', label: t('nav.discover') },
    { to: '/requests', label: t('nav.myRequests'), end: true },
  ];

  return (
    <nav aria-label={t('nav.aria')} className={cn('flex flex-wrap gap-2', className)}>
      {links.map((link) => (
        <NavLink
          key={link.to}
          to={link.to}
          {...(link.end ? { end: true } : {})}
          className={({ isActive }) =>
            cn(
              'rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
              isActive
                ? 'bg-primary/15 text-primary'
                : 'text-muted hover:bg-surface hover:text-text',
            )
          }
        >
          {link.label}
        </NavLink>
      ))}
    </nav>
  );
}
