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
import { useSetupStatus } from '@/api/hooks/useOnboarding';

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
const RequestDiscoverPage = lazy(() => import('@/pages/RequestDiscoverPage'));
const MyRequestsPage = lazy(() => import('@/pages/MyRequestsPage'));
const AdminRequestsPage = lazy(() => import('@/pages/AdminRequestsPage'));
const NotificationSettingsPage = lazy(() => import('@/pages/NotificationSettingsPage'));
const AutomationHubPage = lazy(() => import('@/pages/automation/AutomationHubPage'));
const IndexersPage = lazy(() =>
  import('@/pages/automation/PluginInstancesPage').then((m) => ({ default: m.default })),
);
const DownloadClientsPage = lazy(() =>
  import('@/pages/automation/PluginInstancesPage').then((m) => ({
    default: m.DownloadClientsPage,
  })),
);
const QualityProfilesPage = lazy(() => import('@/pages/automation/QualityProfilesPage'));
const AutomationDownloadsPage = lazy(() => import('@/pages/automation/AutomationDownloadsPage'));
const AutomationMonitorsPage = lazy(() => import('@/pages/automation/AutomationMonitorsPage'));

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
      {isAdmin ? <NavItem to="/admin/requests" label={t('nav.admin', { ns: 'requests' })} /> : null}
      <Button variant="secondary" size="sm" onClick={() => logoutMutation.mutate()}>
        {t('login.logout')}
      </Button>
    </>
  );
}

export default function App(): ReactNode {
  const { t } = useTranslation();
  const accessToken = useAuthStore((s) => s.accessToken);
  const setupQuery = useSetupStatus();
  const onboardingActive =
    setupQuery.data?.completed === false ||
    (setupQuery.data?.phase != null && setupQuery.data.phase !== 'complete');

  return (
    <div className="flex min-h-screen flex-col">
      <header className="border-b border-border bg-surface">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-4 p-4">
          <div className="flex flex-col">
            <span className="text-lg font-semibold">{t('app.name')}</span>
            <span className="text-xs text-muted">{t('app.tagline')}</span>
          </div>
          {!onboardingActive ? <GlobalSearch /> : null}
          <nav aria-label="primary" className="flex flex-wrap items-center gap-1">
            {!onboardingActive && accessToken ? <NavItem to="/" label={t('nav.home')} /> : null}
            {!onboardingActive && accessToken ? (
              <NavItem to="/libraries" label={t('nav.libraries', { ns: 'library' })} />
            ) : null}
            {!onboardingActive && accessToken ? (
              <NavItem to="/requests/discover" label={t('nav.requests')} />
            ) : null}
            {!onboardingActive ? <NavItem to="/status" label={t('nav.status')} /> : null}
            {!onboardingActive ? <NavItem to="/settings" label={t('nav.settings')} /> : null}
            {!onboardingActive ? <AuthNav /> : null}
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
            <Route
              path="/requests/discover"
              element={
                <ProtectedRoute>
                  <RequestDiscoverPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/requests"
              element={
                <ProtectedRoute>
                  <MyRequestsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/admin/requests"
              element={
                <ProtectedRoute adminOnly>
                  <AdminRequestsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings/notifications"
              element={
                <ProtectedRoute>
                  <NotificationSettingsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings/automation"
              element={
                <ProtectedRoute adminOnly>
                  <AutomationHubPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings/automation/indexers"
              element={
                <ProtectedRoute adminOnly>
                  <IndexersPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings/automation/indexers/new"
              element={
                <ProtectedRoute adminOnly>
                  <IndexersPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings/automation/indexers/:id"
              element={
                <ProtectedRoute adminOnly>
                  <IndexersPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings/automation/download-clients"
              element={
                <ProtectedRoute adminOnly>
                  <DownloadClientsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings/automation/download-clients/new"
              element={
                <ProtectedRoute adminOnly>
                  <DownloadClientsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings/automation/download-clients/:id"
              element={
                <ProtectedRoute adminOnly>
                  <DownloadClientsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings/automation/quality-profiles"
              element={
                <ProtectedRoute adminOnly>
                  <QualityProfilesPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings/automation/quality-profiles/new"
              element={
                <ProtectedRoute adminOnly>
                  <QualityProfilesPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings/automation/quality-profiles/:id"
              element={
                <ProtectedRoute adminOnly>
                  <QualityProfilesPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings/automation/monitors"
              element={
                <ProtectedRoute adminOnly>
                  <AutomationMonitorsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings/automation/monitors/new"
              element={
                <ProtectedRoute adminOnly>
                  <AutomationMonitorsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings/automation/monitors/:id"
              element={
                <ProtectedRoute adminOnly>
                  <AutomationMonitorsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/automation/downloads"
              element={
                <ProtectedRoute adminOnly>
                  <AutomationDownloadsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/automation/downloads/:id"
              element={
                <ProtectedRoute adminOnly>
                  <AutomationDownloadsPage />
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
