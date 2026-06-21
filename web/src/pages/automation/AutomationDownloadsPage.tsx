import { type ReactNode } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Card } from '@/components/ui/Card';
import { useAutomationDownload, useAutomationDownloads } from '@/api/hooks/useAutomation';

function formatProgress(progress?: number): string {
  if (progress == null) return '—';
  return `${Math.round(progress * 100)}%`;
}

export default function AutomationDownloadsPage(): ReactNode {
  const { t } = useTranslation('automation');
  const { id } = useParams();
  const downloadId = id ? Number(id) : 0;

  const listQuery = useAutomationDownloads(!downloadId);
  const detailQuery = useAutomationDownload(downloadId, downloadId > 0);

  if (downloadId > 0) {
    const download = detailQuery.data;
    return (
      <div className="mx-auto max-w-3xl space-y-6 p-6">
        <header className="space-y-2">
          <Link to="/automation/downloads" className="text-sm text-primary hover:underline">
            ← {t('downloads.detail.back')}
          </Link>
          <h1 className="text-2xl font-semibold">{t('downloads.detail.title')}</h1>
        </header>
        {detailQuery.isLoading || !download ? (
          <p className="text-sm text-muted">{t('downloads.subtitle')}</p>
        ) : (
          <Card className="space-y-2 p-4 text-sm">
            <p>
              <strong>{download.title}</strong>
            </p>
            <p>
              {t('downloads.columns.status')}:{' '}
              {t(`downloads.status.${download.status ?? 'queued'}` as 'downloads.status.queued')}
            </p>
            <p>
              {t('downloads.columns.progress')}: {formatProgress(download.progress)}
            </p>
            <p>
              {t('downloads.columns.protocol')}: {download.protocol}
            </p>
            {download.requestId ? (
              <p>
                {t('downloads.columns.request')}: #{download.requestId}
              </p>
            ) : null}
            {download.savePath ? (
              <p>
                {t('downloads.detail.savePath')}: {download.savePath}
              </p>
            ) : null}
            {download.clientDownloadId ? (
              <p>
                {t('downloads.detail.clientId')}: {download.clientDownloadId}
              </p>
            ) : null}
          </Card>
        )}
      </div>
    );
  }

  const downloads = listQuery.data?.downloads ?? [];

  return (
    <div className="mx-auto max-w-4xl space-y-6 p-6">
      <header className="space-y-2">
        <Link to="/settings/automation" className="text-sm text-primary hover:underline">
          ← {t('hub.title')}
        </Link>
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div>
            <h1 className="text-2xl font-semibold">{t('downloads.title')}</h1>
            <p className="text-sm text-muted">{t('downloads.subtitle')}</p>
          </div>
          <span className="text-xs text-muted">{t('downloads.polling')}</span>
        </div>
      </header>
      {downloads.length === 0 ? (
        <p className="text-sm text-muted">{t('downloads.empty')}</p>
      ) : (
        <Card className="overflow-x-auto p-4">
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-border text-muted">
                <th className="py-2">{t('downloads.columns.title')}</th>
                <th className="py-2">{t('downloads.columns.status')}</th>
                <th className="py-2">{t('downloads.columns.progress')}</th>
                <th className="py-2">{t('downloads.columns.protocol')}</th>
                <th className="py-2">{t('downloads.columns.request')}</th>
              </tr>
            </thead>
            <tbody>
              {downloads.map((download) => (
                <tr key={download.id} className="border-b border-border/50">
                  <td className="py-2">
                    <Link
                      to={`/automation/downloads/${download.id}`}
                      className="text-primary hover:underline"
                    >
                      {download.title}
                    </Link>
                  </td>
                  <td className="py-2">
                    {t(`downloads.status.${download.status ?? 'queued'}` as 'downloads.status.queued')}
                  </td>
                  <td className="py-2">{formatProgress(download.progress)}</td>
                  <td className="py-2">{download.protocol}</td>
                  <td className="py-2">{download.requestId ? `#${download.requestId}` : '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}
    </div>
  );
}
