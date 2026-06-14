import { type ReactNode, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getProfile, updateProfile } from '@/api/endpoints/auth';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { ThemeSwitcher } from '@/components/ThemeSwitcher';
import { LanguageSwitcher } from '@/components/LanguageSwitcher';
import { useThemeStore } from '@/theme/ThemeProvider';
import { isThemeId, type ThemeId } from '@/theme/themes';
import i18n, { SUPPORTED_LOCALES, type SupportedLocale } from '@/i18n';

const RATINGS = ['G', 'PG', 'PG-13', 'R', 'NC-17'] as const;

export default function ProfilePage(): ReactNode {
  const { t } = useTranslation('auth');
  const queryClient = useQueryClient();
  const setTheme = useThemeStore((s) => s.setTheme);
  const profileQuery = useQuery({ queryKey: ['profile'], queryFn: getProfile });

  const saveMutation = useMutation({
    mutationFn: updateProfile,
    onSuccess: (data) => {
      void queryClient.setQueryData(['profile'], data);
    },
  });

  useEffect(() => {
    if (profileQuery.data && isThemeId(profileQuery.data.theme)) {
      setTheme(profileQuery.data.theme as ThemeId);
    }
    if (profileQuery.data?.locale && SUPPORTED_LOCALES.includes(profileQuery.data.locale as SupportedLocale)) {
      void i18n.changeLanguage(profileQuery.data.locale);
    }
  }, [profileQuery.data, setTheme]);

  if (profileQuery.isLoading) return <p className="p-6 text-muted">{t('loading')}</p>;
  if (!profileQuery.data) return <p className="p-6 text-muted">{t('profile.loadError')}</p>;

  const profile = profileQuery.data;

  return (
    <div className="mx-auto max-w-lg space-y-4 p-6">
      <h1 className="text-xl font-semibold">{t('profile.title')}</h1>
      <Card className="space-y-4 p-4">
        <div>
          <p className="text-sm text-muted">{t('fields.username')}</p>
          <p className="font-medium">{profile.userId}</p>
        </div>
        <div className="space-y-2">
          <p className="text-sm font-medium">{t('profile.locale')}</p>
          <LanguageSwitcher />
        </div>
        <div className="space-y-2">
          <p className="text-sm font-medium">{t('profile.theme')}</p>
          <ThemeSwitcher />
        </div>
        <div className="space-y-2">
          <p className="text-sm font-medium">{t('profile.parental')}</p>
          <select
            className="w-full rounded-md border border-border bg-surface px-3 py-2 text-sm"
            value={profile.maxContentRating ?? ''}
            onChange={(e) => {
              const val = e.target.value || null;
              saveMutation.mutate({ maxContentRating: val });
            }}
          >
            <option value="">{t('profile.noRatingLimit')}</option>
            {RATINGS.map((r) => (
              <option key={r} value={r}>
                {r}
              </option>
            ))}
          </select>
        </div>
        <Button
          variant="secondary"
          onClick={() => saveMutation.mutate({ locale: i18n.language, theme: useThemeStore.getState().theme })}
        >
          {t('profile.save')}
        </Button>
      </Card>
    </div>
  );
}
