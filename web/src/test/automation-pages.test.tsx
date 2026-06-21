import { describe, expect, it, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { I18nextProvider } from 'react-i18next';
import AutomationHubPage from '@/pages/automation/AutomationHubPage';
import i18n from '@/i18n';
import { TestProviders } from './testUtils';
import { useAuthStore } from '@/stores/auth';

vi.mock('@/api/hooks/usePlugins', () => ({
  usePluginCatalog: () => ({ data: { catalog: [] }, isLoading: false }),
  usePluginInstances: () => ({ data: { instances: [] }, isLoading: false }),
  usePluginInstancesByType: () => ({ instances: [], isLoading: false }),
  usePluginInstance: () => ({ data: undefined, isLoading: false }),
  useCreatePluginInstance: () => ({ mutate: vi.fn(), isPending: false }),
  usePatchPluginInstance: () => ({ mutate: vi.fn(), isPending: false }),
  useDeletePluginInstance: () => ({ mutate: vi.fn(), isPending: false }),
  useTestPluginInstance: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('@/api/hooks/useAutomation', () => ({
  useAutomationDownloads: () => ({ data: { downloads: [] }, isLoading: false }),
  useAutomationDownload: () => ({ data: undefined, isLoading: false }),
  useQualityProfiles: () => ({ data: { profiles: [] }, isLoading: false }),
  useQualityProfile: () => ({ data: undefined, isLoading: false }),
  useCreateQualityProfile: () => ({ mutate: vi.fn(), isPending: false }),
  usePatchQualityProfile: () => ({ mutate: vi.fn(), isPending: false }),
  useAutomationMonitors: () => ({ data: { monitors: [] }, isLoading: false }),
  useAutomationMonitor: () => ({ data: undefined, isLoading: false }),
  useCreateAutomationMonitor: () => ({ mutate: vi.fn(), isPending: false }),
  usePatchAutomationMonitor: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteAutomationMonitor: () => ({ mutate: vi.fn(), isPending: false }),
}));

function renderHub(): ReturnType<typeof render> {
  return render(
    <TestProviders>
      <MemoryRouter>
        <I18nextProvider i18n={i18n}>
          <AutomationHubPage />
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
    renderHub();
    expect(screen.getByRole('heading', { name: /automation/i })).toBeInTheDocument();
    expect(screen.getAllByText(/indexers/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/download clients/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/quality profiles/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/series monitors/i).length).toBeGreaterThan(0);
  });
});
