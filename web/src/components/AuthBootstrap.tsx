import { type ReactNode, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { refreshSession } from '@/api/endpoints/auth';
import { clearAuthSession, setAuthSession } from '@/stores/auth';

interface AuthBootstrapProps {
  children: ReactNode;
}

/**
 * Restores the access token from the HTTP-only refresh cookie on first load.
 * Without this, Zustand auth state is lost on every full page reload.
 */
export function AuthBootstrap({ children }: AuthBootstrapProps): ReactNode {
  const { t } = useTranslation('common');
  const [ready, setReady] = useState(false);

  useEffect(() => {
    let cancelled = false;

    void (async () => {
      try {
        const data = await refreshSession();
        if (!cancelled) {
          setAuthSession(data.accessToken, data.expiresAt, data.user);
        }
      } catch {
        if (!cancelled) {
          clearAuthSession();
        }
      } finally {
        if (!cancelled) {
          setReady(true);
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, []);

  if (!ready) {
    return <p className="p-6 text-muted">{t('states.loading')}</p>;
  }

  return children;
}
