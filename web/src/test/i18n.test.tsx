import { describe, expect, it } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import type { ReactElement } from 'react';
import { I18nextProvider, useTranslation } from 'react-i18next';
import i18n from '@/i18n';

function Probe(): ReactElement {
  const { t } = useTranslation(['common', 'status']);
  return (
    <div>
      <span data-testid="status-title">{t('status:title')}</span>
      <span data-testid="settings-language">{t('common:settings.language.label')}</span>
      <span data-testid="actions-save">{t('common:actions.save')}</span>
    </div>
  );
}

describe('i18n', () => {
  it('switches resolved text when changing language from en-US to tr-TR', async () => {
    await i18n.changeLanguage('en-US');

    render(
      <I18nextProvider i18n={i18n}>
        <Probe />
      </I18nextProvider>,
    );

    expect(screen.getByTestId('status-title')).toHaveTextContent('System status');
    expect(screen.getByTestId('settings-language')).toHaveTextContent('Language');
    expect(screen.getByTestId('actions-save')).toHaveTextContent('Save');

    await act(async () => {
      await i18n.changeLanguage('tr-TR');
    });

    expect(screen.getByTestId('status-title')).toHaveTextContent('Sistem durumu');
    expect(screen.getByTestId('settings-language')).toHaveTextContent('Dil');
    expect(screen.getByTestId('actions-save')).toHaveTextContent('Kaydet');

    await i18n.changeLanguage('en-US');
  });
});
