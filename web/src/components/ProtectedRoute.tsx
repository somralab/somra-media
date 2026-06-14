import { type ReactNode } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/stores/auth';
import { getSetupStatus } from '@/api/endpoints/auth';

interface ProtectedRouteProps {
  children: ReactNode;
  adminOnly?: boolean;
}

function isOnboardingIncomplete(
  setup:
    | {
        completed?: boolean;
        phase?: string;
      }
    | undefined,
): boolean {
  if (!setup) return false;
  return setup.completed === false || (setup.phase != null && setup.phase !== 'complete');
}

export function ProtectedRoute({ children, adminOnly = false }: ProtectedRouteProps): ReactNode {
  const { t } = useTranslation('common');
  const accessToken = useAuthStore((s) => s.accessToken);
  const isAdmin = useAuthStore((s) => s.isAdmin());
  const location = useLocation();

  const setupQuery = useQuery({
    queryKey: ['setup-status'],
    queryFn: getSetupStatus,
    staleTime: 30_000,
  });

  if (!accessToken) {
    if (setupQuery.isLoading) {
      return <p className="p-6 text-muted">{t('states.loading')}</p>;
    }
    if (isOnboardingIncomplete(setupQuery.data)) {
      return <Navigate to="/setup/wizard" replace />;
    }
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  if (setupQuery.isLoading) {
    return <p className="p-6 text-muted">{t('states.loading')}</p>;
  }

  if (isOnboardingIncomplete(setupQuery.data) && location.pathname !== '/setup/wizard') {
    return <Navigate to="/setup/wizard" replace />;
  }

  if (adminOnly && !isAdmin) {
    return <Navigate to="/" replace />;
  }
  return children;
}
