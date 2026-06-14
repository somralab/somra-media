import { type FormEvent, type ReactNode, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import i18n, { SUPPORTED_LOCALES, type SupportedLocale } from '@/i18n';
import { setupAdmin } from '@/api/endpoints/auth';
import { createLibrary, triggerScan, type LibraryKind } from '@/api/endpoints/library';
import {
  useAdvanceOnboarding,
  useCompleteOnboarding,
  useOnboardingStatus,
  useSystemDetect,
} from '@/api/hooks/useOnboarding';
import { subscribeScanProgress, type ScanProgressEvent } from '@/api/scanEvents';
import { setAuthSession } from '@/stores/auth';
import { WizardShell } from '@/components/onboarding/WizardShell';
import { RecommendationCard } from '@/components/onboarding/RecommendationCard';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Card } from '@/components/ui/Card';
import type { OnboardingPhase } from '@/api/endpoints/settings';

export default function OnboardingWizardPage(): ReactNode {
  const { t } = useTranslation('onboarding');
  const { t: ta } = useTranslation('auth');
  const navigate = useNavigate();
  const statusQuery = useOnboardingStatus();
  const detectQuery = useSystemDetect();
  const advance = useAdvanceOnboarding();
  const complete = useCompleteOnboarding();

  const phase = (statusQuery.data?.phase ?? 'language') as OnboardingPhase;
  const [locale, setLocale] = useState<SupportedLocale>(
    (i18n.language as SupportedLocale) || 'en-US',
  );
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [libName, setLibName] = useState('My Library');
  const [libPath, setLibPath] = useState('/media');
  const [libKind, setLibKind] = useState<LibraryKind>('movie');
  const [libraryId, setLibraryId] = useState<number | null>(null);
  const [scanProgress, setScanProgress] = useState<ScanProgressEvent | null>(null);
  const [defaultsApplied, setDefaultsApplied] = useState(false);

  const adminMutation = useMutation({
    mutationFn: () => setupAdmin(username, password),
    onSuccess: (data) => {
      setAuthSession(data.accessToken, data.expiresAt, data.user);
    },
  });

  const libraryMutation = useMutation({
    mutationFn: () =>
      createLibrary({ name: libName, kind: libKind, paths: [libPath], watchEnabled: true }),
    onSuccess: (lib) => {
      setLibraryId(lib.id);
    },
  });

  useEffect(() => {
    if (phase !== 'scan' || libraryId == null) return;
    void triggerScan(libraryId, 'full').catch(() => undefined);
    return subscribeScanProgress(libraryId, setScanProgress);
  }, [phase, libraryId]);

  if (statusQuery.isLoading) {
    return <p className="p-6 text-muted">{t('wizard.subtitle')}</p>;
  }

  if (statusQuery.data?.completed) {
    void navigate('/libraries', { replace: true });
    return null;
  }

  const handleLanguage = async (): Promise<void> => {
    await i18n.changeLanguage(locale);
    await advance.mutateAsync({ phase: 'language', locale });
  };

  const handleAdmin = async (e: FormEvent): Promise<void> => {
    e.preventDefault();
    await adminMutation.mutateAsync();
  };

  const handleLibrary = async (e: FormEvent): Promise<void> => {
    e.preventDefault();
    const lib = await libraryMutation.mutateAsync();
    await advance.mutateAsync({ phase: 'library', libraryId: lib.id });
  };

  const handleDefaults = async (): Promise<void> => {
    await advance.mutateAsync({ phase: 'defaults', applyDefaults: true });
    setDefaultsApplied(true);
  };

  const handleScanNext = async (): Promise<void> => {
    await advance.mutateAsync({ phase: 'scan' });
  };

  const handleFinish = async (): Promise<void> => {
    await complete.mutateAsync();
    void navigate('/libraries', { replace: true });
  };

  const profile = detectQuery.data;
  const memoryGb = profile?.memoryBytes
    ? (profile.memoryBytes / (1024 * 1024 * 1024)).toFixed(1)
    : '?';

  return (
    <WizardShell currentStep={phase}>
      {phase === 'language' && (
        <Card className="space-y-4 p-6">
          <h2 className="text-lg font-semibold">{t('language.title')}</h2>
          <p className="text-sm text-muted">{t('language.description')}</p>
          <label className="block space-y-1">
            <span className="text-sm">{t('language.systemDefault')}</span>
            <select
              className="w-full rounded-md border border-border bg-surface px-3 py-2"
              value={locale}
              onChange={(e) => setLocale(e.target.value as SupportedLocale)}
            >
              {SUPPORTED_LOCALES.map((lng) => (
                <option key={lng} value={lng}>
                  {lng}
                </option>
              ))}
            </select>
          </label>
          <Button onClick={() => void handleLanguage()} disabled={advance.isPending}>
            {t('wizard.next')}
          </Button>
        </Card>
      )}

      {phase === 'admin' && (
        <Card className="space-y-4 p-6">
          <h2 className="text-lg font-semibold">{t('admin.title')}</h2>
          <p className="text-sm text-muted">{t('admin.description')}</p>
          <form className="space-y-3" onSubmit={(e) => void handleAdmin(e)}>
            <label className="block space-y-1">
              <span className="text-sm">{ta('fields.username')}</span>
              <Input value={username} onChange={(e) => setUsername(e.target.value)} required />
            </label>
            <label className="block space-y-1">
              <span className="text-sm">{ta('fields.password')}</span>
              <Input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </label>
            <Button type="submit" disabled={adminMutation.isPending}>
              {t('wizard.next')}
            </Button>
          </form>
        </Card>
      )}

      {phase === 'library' && (
        <Card className="space-y-4 p-6">
          <h2 className="text-lg font-semibold">{t('library.title')}</h2>
          <p className="text-sm text-muted">{t('library.description')}</p>
          <form className="space-y-3" onSubmit={(e) => void handleLibrary(e)}>
            <label className="block space-y-1">
              <span className="text-sm">{t('library.name')}</span>
              <Input value={libName} onChange={(e) => setLibName(e.target.value)} required />
            </label>
            <label className="block space-y-1">
              <span className="text-sm">{t('library.path')}</span>
              <Input value={libPath} onChange={(e) => setLibPath(e.target.value)} required />
            </label>
            <label className="block space-y-1">
              <span className="text-sm">{t('library.kind')}</span>
              <select
                className="w-full rounded-md border border-border bg-surface px-3 py-2"
                value={libKind}
                onChange={(e) => setLibKind(e.target.value as LibraryKind)}
              >
                <option value="movie">{t('library.kinds.movie')}</option>
                <option value="tv">{t('library.kinds.tv')}</option>
                <option value="music">{t('library.kinds.music')}</option>
              </select>
            </label>
            <Button type="submit" disabled={libraryMutation.isPending}>
              {t('wizard.next')}
            </Button>
          </form>
        </Card>
      )}

      {phase === 'defaults' && (
        <div className="flex flex-col gap-4">
          <h2 className="text-lg font-semibold">{t('defaults.title')}</h2>
          <p className="text-sm text-muted">{t('defaults.description')}</p>
          {profile ? (
            <>
              <RecommendationCard
                title={t('defaults.cpuCores', { count: profile.cpuCores })}
                description={t('defaults.memory', { gb: memoryGb })}
              />
              <RecommendationCard
                title={
                  profile.gpuPresent ? t('defaults.gpu') : t('defaults.noGpu')
                }
                description={t('defaults.transcodeConcurrency', {
                  count: statusQuery.data?.smartDefaults?.maxConcurrentTranscodes ?? 2,
                })}
                applied={defaultsApplied}
              />
              <RecommendationCard
                title={t('defaults.scanSchedule')}
                applied={defaultsApplied}
              />
            </>
          ) : null}
          <Button onClick={() => void handleDefaults()} disabled={advance.isPending}>
            {t('defaults.apply')}
          </Button>
        </div>
      )}

      {phase === 'scan' && (
        <Card className="space-y-4 p-6">
          <h2 className="text-lg font-semibold">{t('scan.title')}</h2>
          <p className="text-sm text-muted">{t('scan.description')}</p>
          {scanProgress ? (
            <p className="text-sm">
              {t('scan.progress', {
                done: scanProgress.filesDone,
                total: scanProgress.filesTotal,
              })}
            </p>
          ) : (
            <p className="text-sm text-muted">{t('scan.waiting')}</p>
          )}
          <Button onClick={() => void handleScanNext()}>{t('wizard.next')}</Button>
        </Card>
      )}

      {phase === 'complete' && (
        <Card className="space-y-4 p-6 text-center">
          <h2 className="text-lg font-semibold">{t('complete.title')}</h2>
          <p className="text-sm text-muted">{t('complete.description')}</p>
          <div className="flex justify-center gap-2">
            <Button onClick={() => void handleFinish()}>{t('wizard.finish')}</Button>
            <Button variant="secondary" onClick={() => void navigate('/libraries')}>
              {t('complete.goLibraries')}
            </Button>
          </div>
        </Card>
      )}
    </WizardShell>
  );
}
