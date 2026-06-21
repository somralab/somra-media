import { type FormEvent, type ReactNode, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Card } from '@/components/ui/Card';
import { RequestsNav } from '@/components/requests/RequestsNav';
import { QualitySelector } from '@/components/requests/QualitySelector';
import { QualityProfilePicker } from '@/components/automation/QualityProfilePicker';
import { useQualityProfiles } from '@/api/hooks/useAutomation';
import {
  Modal,
  ModalContent,
  ModalDescription,
  ModalHeader,
  ModalTitle,
} from '@/components/ui/Modal';
import { useDiscover } from '@/api/hooks/useDiscover';
import { useCreateRequest } from '@/api/hooks/useRequests';
import type {
  RequestDiscoverResult,
  RequestMediaKind,
  RequestQualityResolution,
} from '@/api/endpoints/requests';

export default function RequestDiscoverPage(): ReactNode {
  const { t } = useTranslation('requests');
  const { t: tc } = useTranslation('common');
  const [query, setQuery] = useState('');
  const [submittedQuery, setSubmittedQuery] = useState('');
  const [kindFilter, setKindFilter] = useState<RequestMediaKind | ''>('');
  const [selected, setSelected] = useState<RequestDiscoverResult | null>(null);
  const [quality, setQuality] = useState<RequestQualityResolution>('1080p');
  const [qualityProfile, setQualityProfile] = useState('');

  const profilesQuery = useQualityProfiles(selected !== null);

  const discoverQuery = useDiscover(
    submittedQuery,
    kindFilter || undefined,
    submittedQuery.length > 0,
  );
  const createMutation = useCreateRequest();

  const handleSearch = (e: FormEvent): void => {
    e.preventDefault();
    setSubmittedQuery(query.trim());
  };

  const handleRequest = (): void => {
    if (!selected) return;
    createMutation.mutate(
      {
        mediaKind: selected.mediaKind,
        provider: selected.provider,
        externalId: selected.externalId,
        title: selected.title,
        ...(selected.posterUrl ? { posterUrl: selected.posterUrl } : {}),
        qualityResolution: quality,
        ...(qualityProfile ? { qualityProfile } : {}),
      },
      {
        onSuccess: () => {
          setSelected(null);
        },
      },
    );
  };

  return (
    <div className="mx-auto max-w-4xl space-y-6 p-6">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold">{t('discover.title')}</h1>
        <p className="text-sm text-muted">{t('discover.subtitle')}</p>
        <RequestsNav />
      </header>

      <form className="flex flex-wrap gap-2" onSubmit={handleSearch}>
        <Input
          className="min-w-[200px] flex-1"
          placeholder={t('discover.searchPlaceholder')}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          aria-label={t('discover.searchPlaceholder')}
        />
        <select
          className="rounded-md border border-border bg-surface px-3 py-2 text-sm"
          value={kindFilter}
          onChange={(e) => setKindFilter(e.target.value as RequestMediaKind | '')}
          aria-label={t('discover.kindFilter')}
        >
          <option value="">{t('discover.kindAll')}</option>
          <option value="movie">{t('mediaKind.movie')}</option>
          <option value="tv">{t('mediaKind.tv')}</option>
        </select>
        <Button type="submit">{t('discover.search')}</Button>
      </form>

      {discoverQuery.isLoading ? <p className="text-muted">{tc('states.loading')}</p> : null}

      {discoverQuery.isError ? (
        <p className="text-danger" role="alert">
          {tc('states.error')}
        </p>
      ) : null}

      {submittedQuery && !discoverQuery.isLoading && discoverQuery.data?.results.length === 0 ? (
        <p className="text-muted">{t('discover.empty')}</p>
      ) : null}

      <ul className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {(discoverQuery.data?.results ?? []).map((item) => (
          <li key={`${item.provider}-${item.externalId}`}>
            <Card className="flex h-full flex-col gap-3 p-4">
              <div className="flex gap-3">
                {item.posterUrl ? (
                  <img
                    src={item.posterUrl}
                    alt=""
                    className="h-24 w-16 shrink-0 rounded object-cover"
                  />
                ) : (
                  <div className="flex h-24 w-16 shrink-0 items-center justify-center rounded bg-surface text-xs text-muted">
                    {t('discover.noPoster')}
                  </div>
                )}
                <div className="min-w-0 flex-1">
                  <p className="font-medium">{item.title}</p>
                  <p className="text-xs text-muted">{t(`mediaKind.${item.mediaKind}`)}</p>
                  {item.inLibrary ? (
                    <p className="mt-1 text-xs text-emerald-600">{t('discover.inLibrary')}</p>
                  ) : null}
                </div>
              </div>
              <Button
                variant="secondary"
                size="sm"
                disabled={item.inLibrary}
                onClick={() => {
                  setSelected(item);
                  setQuality('1080p');
                  setQualityProfile('');
                }}
              >
                {t('discover.requestButton')}
              </Button>
            </Card>
          </li>
        ))}
      </ul>

      <Modal open={selected !== null} onOpenChange={(open) => !open && setSelected(null)}>
        <ModalContent aria-describedby="request-modal-desc">
          <ModalHeader>
            <ModalTitle>{t('discover.modalTitle')}</ModalTitle>
            <ModalDescription id="request-modal-desc">
              {selected ? t('discover.modalDescription', { title: selected.title }) : ''}
            </ModalDescription>
          </ModalHeader>
          <QualitySelector
            value={quality}
            onChange={setQuality}
            disabled={createMutation.isPending}
          />
          <QualityProfilePicker
            profiles={profilesQuery.data?.profiles ?? []}
            value={qualityProfile}
            onChange={setQualityProfile}
            disabled={createMutation.isPending}
          />
          <div className="mt-4 flex justify-end gap-2">
            <Button variant="ghost" onClick={() => setSelected(null)}>
              {tc('actions.cancel')}
            </Button>
            <Button onClick={handleRequest} disabled={createMutation.isPending}>
              {createMutation.isPending ? tc('states.loading') : t('discover.confirmRequest')}
            </Button>
          </div>
          {createMutation.isError ? (
            <p className="mt-2 text-sm text-danger" role="alert">
              {t('discover.createFailed')}
            </p>
          ) : null}
        </ModalContent>
      </Modal>
    </div>
  );
}
