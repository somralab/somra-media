import { type FormEvent, type ReactNode, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useMutation, useQuery } from '@tanstack/react-query';
import { getSetupStatus, login, setupAdmin, updateProfile } from '@/api/endpoints/auth';
import { setAuthSession } from '@/stores/auth';
import { useThemeStore } from '@/theme/ThemeProvider';
import { isThemeId } from '@/theme/themes';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Card } from '@/components/ui/Card';

export default function LoginPage(): ReactNode {
  const { t } = useTranslation('auth');
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as { from?: string } | null)?.from ?? '/libraries';
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  const setupQuery = useQuery({ queryKey: ['setup-status'], queryFn: getSetupStatus });
  const isSetup = setupQuery.data?.setupRequired ?? false;

  const authMutation = useMutation({
    mutationFn: () => (isSetup ? setupAdmin(username, password) : login(username, password)),
    onSuccess: async (data) => {
      setAuthSession(data.accessToken, data.expiresAt, data.user);
      const localTheme = useThemeStore.getState().theme;
      if (isThemeId(localTheme)) {
        try {
          await updateProfile({ theme: localTheme });
        } catch {
          /* best-effort theme sync */
        }
      }
      void navigate(from, { replace: true });
    },
  });

  const handleSubmit = (e: FormEvent): void => {
    e.preventDefault();
    authMutation.mutate();
  };

  if (setupQuery.isLoading) {
    return <p className="p-6 text-muted">{t('loading')}</p>;
  }

  return (
    <div className="mx-auto flex min-h-[60vh] max-w-md flex-col justify-center p-6">
      <Card className="space-y-4 p-6">
        <h1 className="text-xl font-semibold">{isSetup ? t('setup.title') : t('login.title')}</h1>
        <p className="text-sm text-muted">{isSetup ? t('setup.description') : t('login.description')}</p>
        <form className="space-y-3" onSubmit={handleSubmit}>
          <label className="block space-y-1">
            <span className="text-sm">{t('fields.username')}</span>
            <Input
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              autoComplete="username"
              required
            />
          </label>
          <label className="block space-y-1">
            <span className="text-sm">{t('fields.password')}</span>
            <Input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete={isSetup ? 'new-password' : 'current-password'}
              required
            />
          </label>
          {authMutation.isError ? (
            <p className="text-sm text-danger" role="alert">
              {t('login.error')}
            </p>
          ) : null}
          <Button type="submit" disabled={authMutation.isPending} className="w-full">
            {isSetup ? t('setup.submit') : t('login.submit')}
          </Button>
        </form>
        {!isSetup ? (
          <p className="text-center text-sm text-muted">
            <Link to="/" className="text-primary hover:underline">
              {t('login.back')}
            </Link>
          </p>
        ) : null}
      </Card>
    </div>
  );
}
