import { type ReactNode, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Modal, ModalContent, ModalHeader, ModalTitle, ModalClose } from '@/components/ui/Modal';
import {
  downloadSubtitle,
  listMediaSubtitles,
  searchSubtitles,
  uploadSubtitle,
} from '@/api/endpoints/subtitles';

interface SubtitleSectionProps {
  itemId: number;
}

export function SubtitleSection({ itemId }: SubtitleSectionProps): ReactNode {
  const { t } = useTranslation('subtitles');
  const qc = useQueryClient();
  const [searchOpen, setSearchOpen] = useState(false);
  const [uploadOpen, setUploadOpen] = useState(false);
  const [language, setLanguage] = useState('en');
  const [query, setQuery] = useState('');

  const { data: subtitles = [], isLoading } = useQuery({
    queryKey: ['subtitles', itemId],
    queryFn: () => listMediaSubtitles(itemId),
  });

  const searchMutation = useMutation({
    mutationFn: () => searchSubtitles({ mediaItemId: itemId, language, query }),
  });

  const downloadMutation = useMutation({
    mutationFn: (result: { provider: string; externalId: string; language: string }) =>
      downloadSubtitle({ mediaItemId: itemId, ...result }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['subtitles', itemId] }),
  });

  const uploadMutation = useMutation({
    mutationFn: (file: File) => uploadSubtitle(itemId, language, file),
    onSuccess: () => {
      setUploadOpen(false);
      void qc.invalidateQueries({ queryKey: ['subtitles', itemId] });
    },
  });

  return (
    <Card data-testid="subtitle-section">
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>{t('section.title')}</CardTitle>
        <div className="flex gap-2">
          <Button size="sm" variant="secondary" onClick={() => setSearchOpen(true)}>
            {t('actions.search')}
          </Button>
          <Button size="sm" variant="secondary" onClick={() => setUploadOpen(true)}>
            {t('actions.upload')}
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <p className="text-sm text-muted">{t('actions.refresh')}</p>
        ) : subtitles.length === 0 ? (
          <p className="text-sm text-muted">{t('section.empty')}</p>
        ) : (
          <ul className="flex flex-col gap-2 text-sm">
            {subtitles.map((sub) => (
              <li
                key={sub.id}
                className="flex items-center justify-between rounded border border-border px-3 py-2"
              >
                <span>
                  {sub.language} · {t(`source.${sub.source}`)}
                </span>
                {sub.provider ? <span className="text-muted">{sub.provider}</span> : null}
              </li>
            ))}
          </ul>
        )}
      </CardContent>

      <Modal open={searchOpen} onOpenChange={setSearchOpen}>
        <ModalContent>
          <ModalHeader>
            <ModalTitle>{t('search.title')}</ModalTitle>
            <ModalClose className="absolute right-4 top-4 text-muted">×</ModalClose>
          </ModalHeader>
          <div className="space-y-3">
            <label className="block space-y-1 text-sm">
              <span>{t('search.language')}</span>
              <Input value={language} onChange={(e) => setLanguage(e.target.value)} />
            </label>
            <label className="block space-y-1 text-sm">
              <span>{t('search.query')}</span>
              <Input value={query} onChange={(e) => setQuery(e.target.value)} />
            </label>
            <Button onClick={() => searchMutation.mutate()} disabled={searchMutation.isPending}>
              {t('actions.search')}
            </Button>
            {searchMutation.data?.length === 0 ? (
              <p className="text-sm text-muted">{t('search.noResults')}</p>
            ) : null}
            <ul className="flex flex-col gap-2">
              {searchMutation.data?.map((r) => (
                <li
                  key={r.externalId}
                  className="flex items-center justify-between rounded border border-border px-3 py-2 text-sm"
                >
                  <span>
                    {r.language} · {r.releaseName ?? r.externalId} ·{' '}
                    {t('search.score', { score: Math.round(r.score) })}
                  </span>
                  <Button
                    size="sm"
                    disabled={downloadMutation.isPending}
                    onClick={() =>
                      downloadMutation.mutate({
                        provider: r.provider,
                        externalId: r.externalId,
                        language: r.language,
                      })
                    }
                  >
                    {t('actions.download')}
                  </Button>
                </li>
              ))}
            </ul>
          </div>
        </ModalContent>
      </Modal>

      <Modal open={uploadOpen} onOpenChange={setUploadOpen}>
        <ModalContent>
          <ModalHeader>
            <ModalTitle>{t('upload.title')}</ModalTitle>
            <ModalClose className="absolute right-4 top-4 text-muted">×</ModalClose>
          </ModalHeader>
          <div className="space-y-3">
            <label className="block space-y-1 text-sm">
              <span>{t('upload.language')}</span>
              <Input value={language} onChange={(e) => setLanguage(e.target.value)} />
            </label>
            <label className="block space-y-1 text-sm">
              <span>{t('upload.file')}</span>
              <input
                type="file"
                accept=".srt,.vtt"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) uploadMutation.mutate(file);
                }}
              />
            </label>
          </div>
        </ModalContent>
      </Modal>
    </Card>
  );
}
