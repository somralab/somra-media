import { type ReactNode, Suspense, lazy, useState } from 'react';
import { NavLink, Route, Routes, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useMutation } from '@tanstack/react-query';
import { cn } from '@/lib/cn';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { useAuthStore, clearAuthSession } from '@/stores/auth';
import { logout } from '@/api/endpoints/auth';
import { Button } from '@/components/ui/Button';
import { SearchBar } from '@/components/search/SearchBar';
import { SearchResultsDropdown } from '@/components/search/SearchResultsDropdown';
import { useSearchMedia } from '@/api/hooks/useBrowse';

const HomePage = lazy(() => import('@/pages/HomePage'));
const StatusPage = lazy(() => import('@/pages/StatusPage'));
const SettingsPage = lazy(() => import('@/pages/SettingsPage'));
const LibraryPage = lazy(() => import('@/pages/LibraryPage'));
const LibraryDetailPage = lazy(() => import('@/pages/LibraryDetailPage'));
const MediaDetailPage = lazy(() => import('@/pages/MediaDetailPage'));
const PlayerPage = lazy(() => import('@/pages/PlayerPage'));
const OnboardingWizardPage = lazy(() => import('@/pages/OnboardingWizardPage'));
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

function GlobalSearch(): ReactNode {
  const accessToken = useAuthStore((s) => s.accessToken);
  const [query, setQuery] = useState('');
  const [open, setOpen] = useState(false);
  const { data, isLoading } = useSearchMedia(query, Boolean(accessToken) && open);

  if (!accessToken) {
    return null;
  }

  return (
    <div className="relative hidden sm:block">
      <SearchBar
        value={query}
        onChange={(v) => {
          setQuery(v);
          setOpen(true);
        }}
        onSubmit={() => setOpen(true)}
      />
      {open && (
        <SearchResultsDropdown
          results={data?.results ?? []}
          query={query}
          isLoading={isLoading}
          onClose={() => setOpen(false)}
        />
      )}
    </div>
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
  const accessToken = useAuthStore((s) => s.accessToken);

  return (
    <div className="flex min-h-screen flex-col">
      <header className="border-b border-border bg-surface">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-4 p-4">
          <div className="flex flex-col">
            <span className="text-lg font-semibold">{t('app.name')}</span>
            <span className="text-xs text-muted">{t('app.tagline')}</span>
          </div>
          <GlobalSearch />
          <nav aria-label="primary" className="flex flex-wrap items-center gap-1">
            {accessToken ? <NavItem to="/" label={t('nav.home')} /> : null}
            {accessToken ? (
              <NavItem to="/libraries" label={t('nav.libraries', { ns: 'library' })} />
            ) : null}
            <NavItem to="/status" label={t('nav.status')} />
            <NavItem to="/settings" label={t('nav.settings')} />
            <AuthNav />
          </nav>
        </div>
      </header>
      <main className="flex-1">
        <Suspense fallback={<p className="p-6 text-muted">{t('states.loading')}</p>}>
          <Routes>
            <Route
              path="/"
              element={
                <ProtectedRoute>
                  <HomePage />
                </ProtectedRoute>
              }
            />
            <Route path="/status" element={<StatusPage />} />
            <Route path="/setup/wizard" element={<OnboardingWizardPage />} />
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
              path="/libraries/:libraryId/items/:itemId"
              element={
                <ProtectedRoute>
                  <MediaDetailPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/libraries/:libraryId/items/:itemId/play"
              element={
                <ProtectedRoute>
                  <PlayerPage />
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
