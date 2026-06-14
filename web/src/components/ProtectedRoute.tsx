import { type ReactNode } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/auth';
import { getSetupStatus } from '@/api/endpoints/auth';

interface ProtectedRouteProps {
  children: ReactNode;
  adminOnly?: boolean;
}

export function ProtectedRoute({ children, adminOnly = false }: ProtectedRouteProps): ReactNode {
  const accessToken = useAuthStore((s) => s.accessToken);
  const isAdmin = useAuthStore((s) => s.isAdmin());
  const location = useLocation();

  const setupQuery = useQuery({
    queryKey: ['setup-status'],
    queryFn: getSetupStatus,
    enabled: Boolean(accessToken),
    staleTime: 30_000,
  });

  if (!accessToken) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  const onboardingIncomplete =
    setupQuery.data?.completed === false ||
    (setupQuery.data?.phase != null && setupQuery.data.phase !== 'complete');

  if (onboardingIncomplete && location.pathname !== '/setup/wizard') {
    return <Navigate to="/setup/wizard" replace />;
  }

  if (adminOnly && !isAdmin) {
    return <Navigate to="/" replace />;
  }
  return children;
}
