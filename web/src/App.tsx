import { type ReactNode, Suspense, lazy } from 'react';
import { NavLink, Route, Routes } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/cn';

const StatusPage = lazy(() => import('@/pages/StatusPage'));
const SettingsPage = lazy(() => import('@/pages/SettingsPage'));
const LibraryPage = lazy(() => import('@/pages/LibraryPage'));
const LibraryDetailPage = lazy(() => import('@/pages/LibraryDetailPage'));

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
          </nav>
        </div>
      </header>
      <main className="flex-1">
        <Suspense fallback={<p className="p-6 text-muted">{t('states.loading')}</p>}>
          <Routes>
            <Route path="/" element={<StatusPage />} />
            <Route path="/libraries" element={<LibraryPage />} />
            <Route path="/libraries/:id" element={<LibraryDetailPage />} />
            <Route path="/settings" element={<SettingsPage />} />
          </Routes>
        </Suspense>
      </main>
    </div>
  );
}
