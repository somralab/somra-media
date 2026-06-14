import { useCallback, type ChangeEvent, type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { useThemeStore } from '@/theme/ThemeProvider';
import { THEME_IDS, isThemeId } from '@/theme/themes';
import { cn } from '@/lib/cn';

export interface ThemeSwitcherProps {
  className?: string;
}

export function ThemeSwitcher({ className }: ThemeSwitcherProps): ReactNode {
  const { t } = useTranslation();
  const theme = useThemeStore((s) => s.theme);
  const setTheme = useThemeStore((s) => s.setTheme);

  const handleChange = useCallback(
    (event: ChangeEvent<HTMLSelectElement>) => {
      const value = event.target.value;
      if (isThemeId(value)) {
        setTheme(value);
      }
    },
    [setTheme],
  );

  return (
    <label className={cn('flex flex-col gap-1 text-sm', className)}>
      <span className="font-medium">{t('settings.theme.label')}</span>
      <select
        value={theme}
        onChange={handleChange}
        className="h-10 rounded-md border border-border bg-surface px-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
        aria-label={t('settings.theme.label')}
        data-testid="theme-switcher"
      >
        {THEME_IDS.map((id) => (
          <option key={id} value={id}>
            {t(`settings.theme.options.${id}`)}
          </option>
        ))}
      </select>
    </label>
  );
}
