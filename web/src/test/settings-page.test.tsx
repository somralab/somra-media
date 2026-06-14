import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import SettingsPage from '@/pages/SettingsPage';
import { ThemeProvider } from '@/theme/ThemeProvider';
import i18n from '@/i18n';
import { TestProviders } from './testUtils';

vi.mock('@/api/hooks/useSettings', () => ({
  useSettings: () => ({ data: undefined, isLoading: true }),
  usePatchSettings: () => ({ mutate: vi.fn(), isPending: false }),
}));

describe('<SettingsPage />', () => {
  it('renders the language and theme cards', async () => {
    await i18n.changeLanguage('en-US');
    render(
      <TestProviders>
        <I18nextProvider i18n={i18n}>
          <ThemeProvider>
            <SettingsPage />
          </ThemeProvider>
        </I18nextProvider>
      </TestProviders>,
    );
    expect(screen.getByRole('heading', { name: /settings/i })).toBeInTheDocument();
    expect(screen.getAllByText(/language/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/theme/i).length).toBeGreaterThan(0);
  });
});
