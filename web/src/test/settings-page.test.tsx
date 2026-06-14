import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import SettingsPage from '@/pages/SettingsPage';
import { ThemeProvider } from '@/theme/ThemeProvider';
import i18n from '@/i18n';

describe('<SettingsPage />', () => {
  it('renders the language, theme and locale demo cards', async () => {
    await i18n.changeLanguage('en-US');
    render(
      <I18nextProvider i18n={i18n}>
        <ThemeProvider>
          <SettingsPage />
        </ThemeProvider>
      </I18nextProvider>,
    );
    expect(screen.getByRole('heading', { name: 'Settings' })).toBeInTheDocument();
    expect(screen.getAllByText('Language').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Theme').length).toBeGreaterThan(0);
    expect(screen.getByTestId('settings-demo')).toBeInTheDocument();
  });
});
