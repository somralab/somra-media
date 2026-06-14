import { describe, expect, it, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { I18nextProvider } from 'react-i18next';
import SettingsPage from '@/pages/SettingsPage';
import { ThemeProvider } from '@/theme/ThemeProvider';
import i18n from '@/i18n';
import { TestProviders } from './testUtils';
import { useAuthStore } from '@/stores/auth';

const patchMutate = vi.fn();

const settingsData = {
  general: { defaultLocale: 'en-US' },
  library: { scanCron: '0 3 * * *' },
  playback: {
    maxConcurrentTranscodes: 2,
    hwMode: 'auto',
    hwAccelerator: 'auto',
    maxHWTranscodes: 2,
  },
  subtitles: { autoDownload: false, preferredLanguages: ['en'], apiKeySet: false },
};

vi.mock('@/api/hooks/useSettings', () => ({
  useSettings: (enabled = true) => ({
    data: enabled ? settingsData : undefined,
    isLoading: false,
    isError: false,
  }),
  usePatchSettings: () => ({ mutate: patchMutate, isPending: false }),
}));

function renderSettings(): ReturnType<typeof render> {
  return render(
    <TestProviders>
      <MemoryRouter>
        <I18nextProvider i18n={i18n}>
          <ThemeProvider>
            <SettingsPage />
          </ThemeProvider>
        </I18nextProvider>
      </MemoryRouter>
    </TestProviders>,
  );
}

describe('<SettingsPage />', () => {
  beforeEach(() => {
    useAuthStore.setState({
      accessToken: 'token',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
      user: { id: '1', username: 'admin', roles: ['admin'], disabled: false },
    });
  });

  it('renders the language and theme cards', async () => {
    await i18n.changeLanguage('en-US');
    renderSettings();
    expect(screen.getByRole('heading', { name: /settings/i })).toBeInTheDocument();
    expect(screen.getAllByText(/language/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/theme/i).length).toBeGreaterThan(0);
  });

  it('shows advanced categories when toggled', async () => {
    await i18n.changeLanguage('en-US');
    renderSettings();

    fireEvent.click(screen.getByRole('button', { name: /advanced/i }));
    expect(screen.getByRole('heading', { name: 'Library' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Playback' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Subtitles' })).toBeInTheDocument();
  });

  it('patches general locale on change', async () => {
    await i18n.changeLanguage('en-US');
    renderSettings();

    fireEvent.change(screen.getByDisplayValue('en-US'), { target: { value: 'tr-TR' } });
    expect(patchMutate).toHaveBeenCalledWith({ defaultLocale: 'tr-TR' });
  });

  it('shows HW acceleration controls in advanced mode', async () => {
    await i18n.changeLanguage('en-US');
    renderSettings();
    fireEvent.click(screen.getByRole('button', { name: /advanced/i }));
    expect(screen.getByText(/hardware acceleration/i)).toBeInTheDocument();
    expect(screen.getByText(/acceleration mode/i)).toBeInTheDocument();
  });

  it('renders users management link', async () => {
    await i18n.changeLanguage('en-US');
    renderSettings();
    expect(screen.getByRole('link', { name: /manage users/i })).toHaveAttribute(
      'href',
      '/admin/users',
    );
  });

  it('shows auth required banner without a session', async () => {
    useAuthStore.getState().clearSession();
    await i18n.changeLanguage('en-US');
    renderSettings();
    expect(
      screen.getByText(/sign in as an administrator to view and change server settings/i),
    ).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'Library' })).not.toBeInTheDocument();
  });
});
