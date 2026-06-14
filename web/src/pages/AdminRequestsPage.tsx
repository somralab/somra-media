import { type FormEvent, type ReactNode, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { RequestStatusBadge } from '@/components/requests/RequestStatusBadge';
import {
  useApproveRequest,
  usePatchRequestPolicies,
  useRejectRequest,
  useRequestPolicies,
  useRequestRealtimeSync,
  useRequests,
} from '@/api/hooks/useRequests';

type AdminTab = 'queue' | 'policies';

export default function AdminRequestsPage(): ReactNode {
  const { t } = useTranslation('requests');
  const { t: tc } = useTranslation('common');
  const [tab, setTab] = useState<AdminTab>('queue');
  const [rejectNote, setRejectNote] = useState('');
  const [rejectingId, setRejectingId] = useState<number | null>(null);
  const [quota, setQuota] = useState<string>('');
  const [autoApproveRoles, setAutoApproveRoles] = useState<string>('');

  const { connected } = useRequestRealtimeSync();
  const pendingQuery = useRequests({ status: 'pending' });
  const policiesQuery = useRequestPolicies(tab === 'policies');
  const approveMutation = useApproveRequest();
  const rejectMutation = useRejectRequest();
  const patchPolicies = usePatchRequestPolicies();

  const pending = pendingQuery.data?.requests ?? [];
  const policies = policiesQuery.data;

  useEffect(() => {
    if (policies && tab === 'policies') {
      setQuota(String(policies.userQuotaPerMonth));
      setAutoApproveRoles(policies.autoApproveRoles.join(', '));
    }
  }, [policies, tab]);

  const handleSavePolicies = (e: FormEvent): void => {
    e.preventDefault();
    patchPolicies.mutate({
      userQuotaPerMonth: Number(quota),
      autoApproveRoles: autoApproveRoles
        .split(',')
        .map((r) => r.trim())
        .filter(Boolean),
    });
  };

  return (
    <div className="mx-auto max-w-4xl space-y-6 p-6">
      <header className="space-y-2">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h1 className="text-2xl font-semibold">{t('admin.title')}</h1>
          <span className="text-xs text-muted">
            {connected ? t('myRequests.live') : t('myRequests.polling')}
          </span>
        </div>
        <p className="text-sm text-muted">{t('admin.subtitle')}</p>
        <div className="flex gap-2">
          <Button
            variant={tab === 'queue' ? 'primary' : 'secondary'}
            size="sm"
            onClick={() => setTab('queue')}
          >
            {t('admin.tabs.queue')}
          </Button>
          <Button
            variant={tab === 'policies' ? 'primary' : 'secondary'}
            size="sm"
            onClick={() => setTab('policies')}
          >
            {t('admin.tabs.policies')}
          </Button>
        </div>
      </header>

      {tab === 'queue' ? (
        <>
          {pendingQuery.isLoading ? <p className="text-muted">{tc('states.loading')}</p> : null}
          {!pendingQuery.isLoading && pending.length === 0 ? (
            <p className="text-muted">{t('admin.queueEmpty')}</p>
          ) : null}
          <ul className="space-y-3">
            {pending.map((req) => (
              <li key={req.id}>
                <Card className="space-y-3 p-4">
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div>
                      <p className="font-medium">{req.title}</p>
                      <p className="text-xs text-muted">
                        {t(`mediaKind.${req.mediaKind}`)} ·{' '}
                        {t(`quality.options.${req.qualityResolution}`)}
                      </p>
                      <RequestStatusBadge status={req.status} className="mt-1" />
                    </div>
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        disabled={approveMutation.isPending}
                        onClick={() => approveMutation.mutate({ id: req.id })}
                      >
                        {t('admin.approve')}
                      </Button>
                      <Button
                        variant="danger"
                        size="sm"
                        onClick={() => {
                          setRejectingId(req.id);
                          setRejectNote('');
                        }}
                      >
                        {t('admin.reject')}
                      </Button>
                    </div>
                  </div>
                  {rejectingId === req.id ? (
                    <form
                      className="flex flex-wrap gap-2 border-t border-border pt-3"
                      onSubmit={(e) => {
                        e.preventDefault();
                        rejectMutation.mutate(
                          { id: req.id, body: rejectNote ? { adminNote: rejectNote } : {} },
                          { onSuccess: () => setRejectingId(null) },
                        );
                      }}
                    >
                      <Input
                        className="min-w-[200px] flex-1"
                        placeholder={t('admin.rejectNotePlaceholder')}
                        value={rejectNote}
                        onChange={(e) => setRejectNote(e.target.value)}
                      />
                      <Button type="submit" variant="danger" size="sm">
                        {t('admin.confirmReject')}
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => setRejectingId(null)}
                      >
                        {tc('actions.cancel')}
                      </Button>
                    </form>
                  ) : null}
                </Card>
              </li>
            ))}
          </ul>
        </>
      ) : (
        <Card className="space-y-4 p-4">
          {policiesQuery.isLoading ? <p className="text-muted">{tc('states.loading')}</p> : null}
          {policies ? (
            <form className="space-y-4" onSubmit={handleSavePolicies}>
              <label className="block space-y-1 text-sm">
                <span>{t('admin.policies.quotaLabel')}</span>
                <Input
                  type="number"
                  min={0}
                  value={quota}
                  onChange={(e) => setQuota(e.target.value)}
                />
                <span className="text-xs text-muted">{t('admin.policies.quotaDescription')}</span>
              </label>
              <label className="block space-y-1 text-sm">
                <span>{t('admin.policies.autoApproveLabel')}</span>
                <Input
                  value={autoApproveRoles}
                  onChange={(e) => setAutoApproveRoles(e.target.value)}
                  placeholder={t('admin.policies.autoApprovePlaceholder')}
                />
                <span className="text-xs text-muted">
                  {t('admin.policies.autoApproveDescription')}
                </span>
              </label>
              <Button type="submit" disabled={patchPolicies.isPending}>
                {patchPolicies.isPending ? tc('states.loading') : tc('actions.save')}
              </Button>
              {patchPolicies.isSuccess ? (
                <p className="text-sm text-emerald-600">{t('admin.policies.saved')}</p>
              ) : null}
            </form>
          ) : null}
        </Card>
      )}
    </div>
  );
}
