import { useCallback, type ChangeEvent, type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { SUPPORTED_LOCALES, type SupportedLocale } from '@/i18n';
import { cn } from '@/lib/cn';

export interface LanguageSwitcherProps {
  className?: string;
}

export function LanguageSwitcher({ className }: LanguageSwitcherProps): ReactNode {
  const { t, i18n } = useTranslation();
  const current = (SUPPORTED_LOCALES as readonly string[]).includes(i18n.resolvedLanguage ?? '')
    ? (i18n.resolvedLanguage as SupportedLocale)
    : 'en-US';

  const handleChange = useCallback(
    (event: ChangeEvent<HTMLSelectElement>) => {
      void i18n.changeLanguage(event.target.value);
    },
    [i18n],
  );

  return (
    <label className={cn('flex flex-col gap-1 text-sm', className)}>
      <span className="font-medium">{t('settings.language.label')}</span>
      <select
        value={current}
        onChange={handleChange}
        className="h-10 rounded-md border border-border bg-surface px-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
        aria-label={t('settings.language.label')}
      >
        {SUPPORTED_LOCALES.map((locale) => (
          <option key={locale} value={locale}>
            {t(`settings.language.options.${locale}`)}
          </option>
        ))}
      </select>
    </label>
  );
}
