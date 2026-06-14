import { describe, expect, it } from 'vitest';
import { act, render } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import { ThemeProvider, useThemeStore } from '@/theme/ThemeProvider';
import { ThemeSwitcher } from '@/components/ThemeSwitcher';
import { DEFAULT_THEME } from '@/theme/themes';
import i18n from '@/i18n';

describe('theme', () => {
  it('mounts with the default theme on <html data-theme>', () => {
    act(() => {
      useThemeStore.setState({ theme: DEFAULT_THEME });
    });

    render(
      <I18nextProvider i18n={i18n}>
        <ThemeProvider>
          <ThemeSwitcher />
        </ThemeProvider>
      </I18nextProvider>,
    );

    expect(document.documentElement.getAttribute('data-theme')).toBe(DEFAULT_THEME);
  });

  it('updates <html data-theme> and persists the choice in localStorage', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <ThemeProvider>
          <ThemeSwitcher />
        </ThemeProvider>
      </I18nextProvider>,
    );

    act(() => {
      useThemeStore.getState().setTheme('noir');
    });

    expect(document.documentElement.getAttribute('data-theme')).toBe('noir');

    const persisted = window.localStorage.getItem('somra.theme');
    expect(persisted).not.toBeNull();
    expect(JSON.parse(persisted ?? '{}')).toMatchObject({ state: { theme: 'noir' } });
  });
});
