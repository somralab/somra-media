import { type ReactNode, useState } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { LanguageSwitcher } from '@/components/LanguageSwitcher';
import { ThemeSwitcher } from '@/components/ThemeSwitcher';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { useSettings, usePatchSettings } from '@/api/hooks/useSettings';

export default function SettingsPage(): ReactNode {
  const { t } = useTranslation('settings');
  const { t: tc } = useTranslation();
  const { data: settings, isLoading } = useSettings();
  const [advanced, setAdvanced] = useState(false);
  const patchGeneral = usePatchSettings('general');
  const patchLibrary = usePatchSettings('library');
  const patchPlayback = usePatchSettings('playback');
  const patchSubtitles = usePatchSettings('subtitles');

  const general = settings?.general as Record<string, string> | undefined;
  const library = settings?.library as Record<string, string> | undefined;
  const playback = settings?.playback as Record<string, number> | undefined;
  const subtitles = settings?.subtitles as Record<string, unknown> | undefined;

  return (
    <section className="mx-auto flex max-w-2xl flex-col gap-6 p-6">
      <header className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-semibold">{t('page.title')}</h1>
          <p className="text-sm text-muted">{t('page.subtitle')}</p>
        </div>
        <Button variant="ghost" size="sm" onClick={() => setAdvanced((v) => !v)}>
          {advanced ? t('page.simple') : t('page.advanced')}
        </Button>
      </header>

      <Card>
        <CardHeader>
          <CardTitle>{tc('settings.language.label')}</CardTitle>
          <CardDescription>{tc('settings.language.description')}</CardDescription>
        </CardHeader>
        <CardContent>
          <LanguageSwitcher />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{tc('settings.theme.label')}</CardTitle>
          <CardDescription>{tc('settings.theme.description')}</CardDescription>
        </CardHeader>
        <CardContent>
          <ThemeSwitcher />
        </CardContent>
      </Card>

      {!isLoading && settings ? (
        <>
          <Card>
            <CardHeader>
              <CardTitle>{t('categories.general')}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <label className="block space-y-1 text-sm">
                <span>{t('general.defaultLocale.label')}</span>
                <select
                  className="w-full rounded-md border border-border bg-surface px-3 py-2"
                  value={general?.defaultLocale ?? 'en-US'}
                  onChange={(e) =>
                    patchGeneral.mutate({ defaultLocale: e.target.value })
                  }
                >
                  <option value="en-US">en-US</option>
                  <option value="tr-TR">tr-TR</option>
                </select>
              </label>
              <p className="text-xs text-muted">{t('general.defaultLocale.description')}</p>
            </CardContent>
          </Card>

          {advanced ? (
            <>
              <Card>
                <CardHeader>
                  <CardTitle>{t('categories.library')}</CardTitle>
                </CardHeader>
                <CardContent>
                  <label className="block space-y-1 text-sm">
                    <span>{t('library.scanCron.label')}</span>
                    <Input
                      defaultValue={library?.scanCron ?? '0 3 * * *'}
                      onBlur={(e) => patchLibrary.mutate({ scanCron: e.target.value })}
                    />
                  </label>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t('categories.playback')}</CardTitle>
                </CardHeader>
                <CardContent>
                  <label className="block space-y-1 text-sm">
                    <span>{t('playback.maxConcurrentTranscodes.label')}</span>
                    <Input
                      type="number"
                      min={1}
                      max={8}
                      defaultValue={playback?.maxConcurrentTranscodes ?? 2}
                      onBlur={(e) =>
                        patchPlayback.mutate({
                          maxConcurrentTranscodes: Number(e.target.value),
                        })
                      }
                    />
                  </label>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t('categories.subtitles')}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="checkbox"
                      defaultChecked={Boolean(subtitles?.autoDownload)}
                      onChange={(e) =>
                        patchSubtitles.mutate({ autoDownload: e.target.checked })
                      }
                    />
                    {t('subtitles.autoDownload.label')}
                  </label>
                  <label className="block space-y-1 text-sm">
                    <span>{t('subtitles.preferredLanguages.label')}</span>
                    <Input
                      defaultValue={
                        Array.isArray(subtitles?.preferredLanguages)
                          ? (subtitles?.preferredLanguages as string[]).join(', ')
                          : 'en, tr'
                      }
                      onBlur={(e) =>
                        patchSubtitles.mutate({
                          preferredLanguages: e.target.value.split(',').map((s) => s.trim()),
                        })
                      }
                    />
                  </label>
                  <label className="block space-y-1 text-sm">
                    <span>{t('subtitles.apiKey.label')}</span>
                    <Input
                      type="password"
                      placeholder={subtitles?.apiKeySet ? '••••••••' : ''}
                      onBlur={(e) => {
                        if (e.target.value) patchSubtitles.mutate({ apiKey: e.target.value });
                      }}
                    />
                  </label>
                </CardContent>
              </Card>
            </>
          ) : null}

          <Card>
            <CardHeader>
              <CardTitle>{t('categories.users')}</CardTitle>
              <CardDescription>{t('users.description')}</CardDescription>
            </CardHeader>
            <CardContent>
              <Link to="/admin/users" className="text-sm text-primary hover:underline">
                {t('users.manageLink')}
              </Link>
            </CardContent>
          </Card>
        </>
      ) : null}
    </section>
  );
}
