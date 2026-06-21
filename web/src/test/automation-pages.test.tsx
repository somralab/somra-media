import { type ReactElement } from 'react';
import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { I18nextProvider } from 'react-i18next';
import AutomationHubPage from '@/pages/automation/AutomationHubPage';
import AutomationDownloadsPage from '@/pages/automation/AutomationDownloadsPage';
import AutomationMonitorsPage from '@/pages/automation/AutomationMonitorsPage';
import QualityProfilesPage from '@/pages/automation/QualityProfilesPage';
import { PluginInstancesPage } from '@/pages/automation/PluginInstancesPage';
import { PluginInstanceForm } from '@/components/automation/PluginInstanceForm';
import { PluginInstanceList } from '@/components/automation/PluginInstanceList';
import { PluginTestButton } from '@/components/automation/PluginTestButton';
import { QualityProfilePicker } from '@/components/automation/QualityProfilePicker';
import i18n from '@/i18n';
import { TestProviders } from './testUtils';
import { useAuthStore } from '@/stores/auth';

const mockInstances = [
  {
    id: 1,
    pluginType: 'indexer' as const,
    implementation: 'stub',
    name: 'Primary Indexer',
    config: {},
    enabled: true,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  },
];

const mockProfiles = [{ id: 1, name: 'default', spec: '{}', isDefault: true }];
const mockMonitors = [
  {
    id: 1,
    title: 'Demo Show',
    provider: 'tmdb',
    externalId: 'show-1',
    qualityProfile: 'default',
    enabled: true,
    lastSeason: 0,
    lastEpisode: 0,
  },
];
const mockDownloads = [
  {
    id: 1,
    title: 'Demo Release',
    status: 'queued' as const,
    protocol: 'torrent',
    progress: 0.5,
  },
];

vi.mock('@/api/hooks/usePlugins', () => ({
  usePluginCatalog: () => ({
    data: { catalog: [{ pluginType: 'indexer', implementation: 'stub', displayName: 'Stub' }] },
    isLoading: false,
  }),
  usePluginInstances: () => ({ data: { instances: mockInstances }, isLoading: false }),
  usePluginInstancesByType: () => ({ instances: mockInstances, isLoading: false }),
  usePluginInstance: () => ({ data: mockInstances[0], isLoading: false }),
  useCreatePluginInstance: () => ({ mutate: vi.fn(), isPending: false }),
  usePatchPluginInstance: () => ({ mutate: vi.fn(), isPending: false }),
  useDeletePluginInstance: () => ({ mutate: vi.fn(), isPending: false }),
  useTestPluginInstance: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('@/api/hooks/useAutomation', () => ({
  useAutomationDownloads: () => ({ data: { downloads: mockDownloads }, isLoading: false }),
  useAutomationDownload: () => ({ data: mockDownloads[0], isLoading: false }),
  useQualityProfiles: () => ({ data: { profiles: mockProfiles }, isLoading: false }),
  useQualityProfile: () => ({ data: mockProfiles[0], isLoading: false }),
  useCreateQualityProfile: () => ({ mutate: vi.fn(), isPending: false }),
  usePatchQualityProfile: () => ({ mutate: vi.fn(), isPending: false }),
  useAutomationMonitors: () => ({ data: { monitors: mockMonitors }, isLoading: false }),
  useAutomationMonitor: () => ({ data: mockMonitors[0], isLoading: false }),
  useCreateAutomationMonitor: () => ({ mutate: vi.fn(), isPending: false }),
  usePatchAutomationMonitor: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteAutomationMonitor: () => ({ mutate: vi.fn(), isPending: false }),
}));

function renderAt(
  path: string,
  element: ReactElement,
  routePath: string,
): ReturnType<typeof render> {
  return render(
    <TestProviders>
      <MemoryRouter initialEntries={[path]}>
        <I18nextProvider i18n={i18n}>
          <Routes>
            <Route path={routePath} element={element} />
          </Routes>
        </I18nextProvider>
      </MemoryRouter>
    </TestProviders>,
  );
}

describe('AutomationHubPage', () => {
  beforeEach(() => {
    useAuthStore.setState({
      accessToken: 'token',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
      user: { id: '1', username: 'admin', roles: ['admin'], disabled: false },
    });
  });

  it('renders automation hub sections', async () => {
    await i18n.changeLanguage('en-US');
    renderAt('/settings/automation', <AutomationHubPage />, '/settings/automation');
    expect(screen.getByRole('heading', { name: /automation/i })).toBeInTheDocument();
    expect(screen.getAllByText(/indexers/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/download clients/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/quality profiles/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/series monitors/i).length).toBeGreaterThan(0);
  });
});

describe('Automation pages', () => {
  beforeEach(async () => {
    await i18n.changeLanguage('en-US');
    useAuthStore.setState({
      accessToken: 'token',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
      user: { id: '1', username: 'admin', roles: ['admin'], disabled: false },
    });
  });

  afterEach(() => {
    cleanup();
  });

  it('renders downloads list and detail', () => {
    renderAt('/automation/downloads', <AutomationDownloadsPage />, '/automation/downloads');
    expect(screen.getByRole('heading', { name: /automation downloads/i })).toBeInTheDocument();
    expect(screen.getAllByText('Demo Release').length).toBeGreaterThan(0);

    cleanup();
    renderAt('/automation/downloads/1', <AutomationDownloadsPage />, '/automation/downloads/:id');
    expect(screen.getByRole('heading', { name: /download details/i })).toBeInTheDocument();
  });

  it('renders quality profiles list and form', () => {
    renderAt(
      '/settings/automation/quality-profiles',
      <QualityProfilesPage />,
      '/settings/automation/quality-profiles',
    );
    expect(screen.getByRole('heading', { name: /quality profiles/i })).toBeInTheDocument();
    expect(screen.getByText('default')).toBeInTheDocument();

    cleanup();
    renderAt(
      '/settings/automation/quality-profiles/new',
      <QualityProfilesPage />,
      '/settings/automation/quality-profiles/new',
    );
    expect(screen.getByRole('button', { name: /create/i })).toBeInTheDocument();

    cleanup();
    renderAt(
      '/settings/automation/quality-profiles/1',
      <QualityProfilesPage />,
      '/settings/automation/quality-profiles/:id',
    );
    expect(screen.getByDisplayValue('default')).toBeInTheDocument();
  });

  it('renders monitors list and form', () => {
    renderAt(
      '/settings/automation/monitors',
      <AutomationMonitorsPage />,
      '/settings/automation/monitors',
    );
    expect(screen.getByRole('heading', { name: /series monitors/i })).toBeInTheDocument();
    expect(screen.getByText('Demo Show')).toBeInTheDocument();

    cleanup();
    renderAt(
      '/settings/automation/monitors/new',
      <AutomationMonitorsPage />,
      '/settings/automation/monitors/new',
    );
    expect(screen.getByRole('button', { name: /create/i })).toBeInTheDocument();

    cleanup();
    renderAt(
      '/settings/automation/monitors/1',
      <AutomationMonitorsPage />,
      '/settings/automation/monitors/:id',
    );
    expect(screen.getByDisplayValue('Demo Show')).toBeInTheDocument();
  });

  it('renders plugin instance pages and components', () => {
    renderAt(
      '/settings/automation/indexers',
      <PluginInstancesPage pluginType="indexer" basePath="/settings/automation/indexers" />,
      '/settings/automation/indexers',
    );
    expect(screen.getByText('Primary Indexer')).toBeInTheDocument();

    cleanup();
    renderAt(
      '/settings/automation/indexers/new',
      <PluginInstancesPage pluginType="indexer" basePath="/settings/automation/indexers" />,
      '/settings/automation/indexers/new',
    );
    expect(screen.getByRole('button', { name: /create/i })).toBeInTheDocument();

    cleanup();
    render(
      <TestProviders>
        <MemoryRouter>
          <I18nextProvider i18n={i18n}>
            <PluginInstanceList
              instances={mockInstances}
              pluginType="indexer"
              basePath="/settings/automation/indexers"
            />
          </I18nextProvider>
        </MemoryRouter>
      </TestProviders>,
    );
    expect(screen.getByText('Primary Indexer')).toBeInTheDocument();

    cleanup();
    render(
      <TestProviders>
        <I18nextProvider i18n={i18n}>
          <PluginInstanceForm pluginType="indexer" onSaved={vi.fn()} onCancel={vi.fn()} />
        </I18nextProvider>
      </TestProviders>,
    );
    expect(screen.getByRole('button', { name: /create/i })).toBeInTheDocument();

    cleanup();
    render(
      <TestProviders>
        <I18nextProvider i18n={i18n}>
          <PluginTestButton instanceId={1} />
        </I18nextProvider>
      </TestProviders>,
    );
    expect(screen.getByRole('button', { name: /test/i })).toBeInTheDocument();

    cleanup();
    render(
      <TestProviders>
        <I18nextProvider i18n={i18n}>
          <QualityProfilePicker profiles={mockProfiles} value="default" onChange={vi.fn()} />
        </I18nextProvider>
      </TestProviders>,
    );
    expect(screen.getByRole('combobox')).toBeInTheDocument();
  });
});
