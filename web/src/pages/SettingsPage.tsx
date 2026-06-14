import { useMemo, type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { LanguageSwitcher } from '@/components/LanguageSwitcher';
import { ThemeSwitcher } from '@/components/ThemeSwitcher';

function useLocaleDemo(locale: string): {
  today: string;
  now: string;
  number: string;
} {
  return useMemo(() => {
    const date = new Date();
    return {
      today: new Intl.DateTimeFormat(locale, {
        weekday: 'long',
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      }).format(date),
      now: new Intl.DateTimeFormat(locale, {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      }).format(date),
      number: new Intl.NumberFormat(locale).format(1234567.89),
    };
  }, [locale]);
}

export default function SettingsPage(): ReactNode {
  const { t, i18n } = useTranslation();
  const demo = useLocaleDemo(i18n.resolvedLanguage ?? 'en-US');

  return (
    <section className="mx-auto flex max-w-2xl flex-col gap-6 p-6">
      <header className="flex flex-col gap-1">
        <h1 className="text-2xl font-semibold">{t('settings.title')}</h1>
        <p className="text-sm text-muted">{t('settings.subtitle')}</p>
      </header>

      <Card>
        <CardHeader>
          <CardTitle>{t('settings.language.label')}</CardTitle>
          <CardDescription>{t('settings.language.description')}</CardDescription>
        </CardHeader>
        <CardContent>
          <LanguageSwitcher />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t('settings.theme.label')}</CardTitle>
          <CardDescription>{t('settings.theme.description')}</CardDescription>
        </CardHeader>
        <CardContent>
          <ThemeSwitcher />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t('settings.demo.title')}</CardTitle>
          <CardDescription>{t('settings.demo.description')}</CardDescription>
        </CardHeader>
        <CardContent>
          <dl
            className="grid grid-cols-[max-content,1fr] gap-x-4 gap-y-2 text-sm"
            data-testid="settings-demo"
          >
            <dt className="text-muted">{t('settings.demo.todayLabel')}</dt>
            <dd>{demo.today}</dd>
            <dt className="text-muted">{t('settings.demo.now')}</dt>
            <dd>{demo.now}</dd>
            <dt className="text-muted">{t('settings.demo.numberLabel')}</dt>
            <dd>{demo.number}</dd>
          </dl>
        </CardContent>
      </Card>
    </section>
  );
}
