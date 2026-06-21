import { type FormEvent, type ReactNode, useEffect, useState } from 'react';
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import {
  useCreateQualityProfile,
  usePatchQualityProfile,
  useQualityProfile,
  useQualityProfiles,
} from '@/api/hooks/useAutomation';

const BASE = '/settings/automation/quality-profiles';

export default function QualityProfilesPage(): ReactNode {
  const { t } = useTranslation('automation');
  const { id } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const isNew = location.pathname.endsWith('/new');
  const profileId = id ? Number(id) : 0;

  const profilesQuery = useQualityProfiles(!profileId && !isNew);
  const profileQuery = useQualityProfile(profileId, profileId > 0);
  const createMutation = useCreateQualityProfile();
  const patchMutation = usePatchQualityProfile();

  const [name, setName] = useState('');
  const [spec, setSpec] = useState('{}');
  const [isDefault, setIsDefault] = useState(false);

  useEffect(() => {
    if (profileQuery.data) {
      setName(profileQuery.data.name ?? '');
      setSpec(profileQuery.data.spec ?? '{}');
      setIsDefault(Boolean(profileQuery.data.isDefault));
    }
  }, [profileQuery.data]);

  if (isNew || profileId > 0) {
    const handleSubmit = (e: FormEvent): void => {
      e.preventDefault();
      if (profileId > 0) {
        patchMutation.mutate(
          { id: profileId, body: { name, spec, isDefault } },
          { onSuccess: () => void navigate(BASE) },
        );
        return;
      }
      createMutation.mutate(
        { name, spec, isDefault },
        { onSuccess: () => void navigate(BASE) },
      );
    };

    return (
      <div className="mx-auto max-w-2xl space-y-6 p-6">
        <header className="space-y-2">
          <Link to={BASE} className="text-sm text-primary hover:underline">
            ← {t('qualityProfiles.title')}
          </Link>
          <h1 className="text-2xl font-semibold">
            {isNew ? t('qualityProfiles.add') : t('qualityProfiles.edit')}
          </h1>
        </header>
        <Card className="space-y-4 p-4">
          <form className="space-y-4" onSubmit={handleSubmit}>
            <label className="block space-y-1 text-sm">
              <span>{t('qualityProfiles.name')}</span>
              <Input value={name} onChange={(e) => setName(e.target.value)} required />
            </label>
            <label className="block space-y-1 text-sm">
              <span>{t('qualityProfiles.spec')}</span>
              <textarea
                className="min-h-32 w-full rounded-md border border-border bg-surface px-3 py-2 font-mono text-xs"
                value={spec}
                onChange={(e) => setSpec(e.target.value)}
                spellCheck={false}
              />
            </label>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={isDefault}
                onChange={(e) => setIsDefault(e.target.checked)}
                className="accent-primary"
              />
              {t('qualityProfiles.isDefault')}
            </label>
            <Button type="submit" disabled={createMutation.isPending || patchMutation.isPending}>
              {isNew ? t('qualityProfiles.create') : t('qualityProfiles.save')}
            </Button>
          </form>
        </Card>
      </div>
    );
  }

  const profiles = profilesQuery.data?.profiles ?? [];

  return (
    <div className="mx-auto max-w-2xl space-y-6 p-6">
      <header className="flex flex-wrap items-center justify-between gap-2">
        <div className="space-y-1">
          <Link to="/settings/automation" className="text-sm text-primary hover:underline">
            ← {t('hub.title')}
          </Link>
          <h1 className="text-2xl font-semibold">{t('qualityProfiles.title')}</h1>
          <p className="text-sm text-muted">{t('qualityProfiles.subtitle')}</p>
        </div>
        <Link to={`${BASE}/new`}>
          <Button>{t('qualityProfiles.add')}</Button>
        </Link>
      </header>
      {profiles.length === 0 ? (
        <p className="text-sm text-muted">{t('qualityProfiles.empty')}</p>
      ) : (
        <div className="space-y-3">
          {profiles.map((profile) => (
            <Card key={profile.id} className="flex items-center justify-between p-4">
              <div>
                <p className="font-medium">{profile.name}</p>
                {profile.isDefault ? (
                  <span className="text-xs text-primary">{t('qualityProfiles.isDefault')}</span>
                ) : null}
              </div>
              <Link to={`${BASE}/${profile.id}`}>
                <Button variant="secondary" size="sm">
                  {t('qualityProfiles.edit')}
                </Button>
              </Link>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
