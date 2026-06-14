import { type ReactNode, Suspense, lazy } from 'react';
import { NavLink, Route, Routes, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useMutation } from '@tanstack/react-query';
import { cn } from '@/lib/cn';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { useAuthStore, clearAuthSession } from '@/stores/auth';
import { logout } from '@/api/endpoints/auth';
import { Button } from '@/components/ui/Button';

const StatusPage = lazy(() => import('@/pages/StatusPage'));
const SettingsPage = lazy(() => import('@/pages/SettingsPage'));
const LibraryPage = lazy(() => import('@/pages/LibraryPage'));
const LibraryDetailPage = lazy(() => import('@/pages/LibraryDetailPage'));
const LoginPage = lazy(() => import('@/pages/LoginPage'));
const ProfilePage = lazy(() => import('@/pages/ProfilePage'));
const AdminUsersPage = lazy(() => import('@/pages/AdminUsersPage'));

function NavItem({ to, label }: { to: string; label: string }): ReactNode {
  return (
    <NavLink
      to={to}
      end={to === '/'}
      className={({ isActive }) =>
        cn(
          'rounded-md px-3 py-2 text-sm font-medium transition-colors',
          isActive ? 'bg-primary/15 text-primary' : 'text-muted hover:bg-surface hover:text-text',
        )
      }
    >
      {label}
    </NavLink>
  );
}

function AuthNav(): ReactNode {
  const { t } = useTranslation('auth');
  const navigate = useNavigate();
  const accessToken = useAuthStore((s) => s.accessToken);
  const isAdmin = useAuthStore((s) => s.isAdmin());
  const logoutMutation = useMutation({
    mutationFn: logout,
    onSettled: () => {
      clearAuthSession();
      void navigate('/login');
    },
  });

  if (!accessToken) {
    return <NavItem to="/login" label={t('nav.login')} />;
  }

  return (
    <>
      <NavItem to="/profile" label={t('nav.profile')} />
      {isAdmin ? <NavItem to="/admin/users" label={t('admin.nav')} /> : null}
      <Button variant="secondary" size="sm" onClick={() => logoutMutation.mutate()}>
        {t('login.logout')}
      </Button>
    </>
  );
}

export default function App(): ReactNode {
  const { t } = useTranslation();

  return (
    <div className="flex min-h-screen flex-col">
      <header className="border-b border-border bg-surface">
        <div className="mx-auto flex max-w-5xl items-center justify-between gap-4 p-4">
          <div className="flex flex-col">
            <span className="text-lg font-semibold">{t('app.name')}</span>
            <span className="text-xs text-muted">{t('app.tagline')}</span>
          </div>
          <nav aria-label="primary" className="flex items-center gap-1">
            <NavItem to="/" label={t('nav.status')} />
            <NavItem to="/libraries" label={t('nav.libraries', { ns: 'library' })} />
            <NavItem to="/settings" label={t('nav.settings')} />
            <AuthNav />
          </nav>
        </div>
      </header>
      <main className="flex-1">
        <Suspense fallback={<p className="p-6 text-muted">{t('states.loading')}</p>}>
          <Routes>
            <Route path="/" element={<StatusPage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route
              path="/libraries"
              element={
                <ProtectedRoute>
                  <LibraryPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/libraries/:id"
              element={
                <ProtectedRoute>
                  <LibraryDetailPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/profile"
              element={
                <ProtectedRoute>
                  <ProfilePage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/admin/users"
              element={
                <ProtectedRoute adminOnly>
                  <AdminUsersPage />
                </ProtectedRoute>
              }
            />
            <Route path="/settings" element={<SettingsPage />} />
          </Routes>
        </Suspense>
      </main>
    </div>
  );
}
