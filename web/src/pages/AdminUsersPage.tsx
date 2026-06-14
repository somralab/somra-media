import { type FormEvent, type ReactNode, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createUser, listUsers, updateUser } from '@/api/endpoints/auth';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';

export default function AdminUsersPage(): ReactNode {
  const { t } = useTranslation('auth');
  const queryClient = useQueryClient();
  const usersQuery = useQuery({ queryKey: ['users'], queryFn: listUsers });
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  const createMutation = useMutation({
    mutationFn: () => createUser({ username, password, roles: ['user'] }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['users'] });
      setUsername('');
      setPassword('');
    },
  });

  const toggleMutation = useMutation({
    mutationFn: ({ id, disabled }: { id: string; disabled: boolean }) =>
      updateUser(id, { disabled }),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['users'] }),
  });

  const handleCreate = (e: FormEvent): void => {
    e.preventDefault();
    createMutation.mutate();
  };

  if (usersQuery.isLoading) return <p className="p-6 text-muted">{t('loading')}</p>;

  return (
    <div className="mx-auto max-w-3xl space-y-6 p-6">
      <h1 className="text-xl font-semibold">{t('admin.title')}</h1>
      <Card className="space-y-3 p-4">
        <h2 className="font-medium">{t('admin.create')}</h2>
        <form className="flex flex-wrap gap-2" onSubmit={handleCreate}>
          <Input
            placeholder={t('fields.username')}
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />
          <Input
            type="password"
            placeholder={t('fields.password')}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          <Button type="submit" disabled={createMutation.isPending}>
            {t('admin.createSubmit')}
          </Button>
        </form>
      </Card>
      <Card className="overflow-x-auto p-4">
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b border-border text-muted">
              <th className="py-2">{t('fields.username')}</th>
              <th className="py-2">{t('admin.roles')}</th>
              <th className="py-2">{t('admin.status')}</th>
              <th className="py-2" />
            </tr>
          </thead>
          <tbody>
            {(usersQuery.data ?? []).map((u) => (
              <tr key={u.id} className="border-b border-border/50">
                <td className="py-2">{u.username}</td>
                <td className="py-2">{u.roles.join(', ')}</td>
                <td className="py-2">{u.disabled ? t('admin.disabled') : t('admin.active')}</td>
                <td className="py-2">
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => toggleMutation.mutate({ id: u.id, disabled: !u.disabled })}
                  >
                    {u.disabled ? t('admin.enable') : t('admin.disable')}
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
