import { type FormEvent, type ReactNode, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { useCreateLibrary, useLibraries } from '@/api/hooks/useLibraries';
import type { LibraryKind } from '@/api/endpoints/library';

export default function LibraryPage(): ReactNode {
  const { t } = useTranslation('library');
  const { data: libraries, isLoading, error } = useLibraries();
  const create = useCreateLibrary();
  const [name, setName] = useState('');
  const [kind, setKind] = useState<LibraryKind>('movie');
  const [pathsText, setPathsText] = useState('');

  const paths = useMemo(
    () =>
      pathsText
        .split('\n')
        .map((p) => p.trim())
        .filter(Boolean),
    [pathsText],
  );

  function onSubmit(e: FormEvent): void {
    e.preventDefault();
    create.mutate({ name, kind, paths, watchEnabled: true });
  }

  return (
    <section className="mx-auto flex max-w-3xl flex-col gap-6 p-6">
      <header className="flex flex-col gap-1">
        <h1 className="text-2xl font-semibold">{t('title')}</h1>
        <p className="text-sm text-muted">{t('subtitle')}</p>
      </header>

      <Card>
        <CardHeader>
          <CardTitle>{t('list.create')}</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="flex flex-col gap-4" onSubmit={onSubmit}>
            <label className="flex flex-col gap-1 text-sm">
              <span>{t('form.name')}</span>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder={t('form.namePlaceholder')}
                required
              />
            </label>
            <label className="flex flex-col gap-1 text-sm">
              <span>{t('form.kind')}</span>
              <select
                className="rounded-md border border-border bg-surface px-3 py-2"
                value={kind}
                onChange={(e) => setKind(e.target.value as LibraryKind)}
              >
                <option value="movie">{t('kinds.movie')}</option>
                <option value="tv">{t('kinds.tv')}</option>
                <option value="music">{t('kinds.music')}</option>
              </select>
            </label>
            <label className="flex flex-col gap-1 text-sm">
              <span>{t('form.paths')}</span>
              <span className="text-xs text-muted">{t('form.pathsHelp')}</span>
              <textarea
                className="min-h-24 rounded-md border border-border bg-surface px-3 py-2 font-mono text-sm"
                value={pathsText}
                onChange={(e) => setPathsText(e.target.value)}
                placeholder={t('form.pathsPlaceholder')}
                required
              />
            </label>
            <Button type="submit" disabled={create.isPending}>
              {t('form.submitCreate')}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t('nav.libraries')}</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading && <p className="text-muted">{t('states.loading', { ns: 'common' })}</p>}
          {error && <p className="text-danger">{t('states.error', { ns: 'common' })}</p>}
          {!isLoading && libraries?.length === 0 && <p className="text-muted">{t('list.empty')}</p>}
          <ul className="flex flex-col gap-2">
            {libraries?.map((lib) => (
              <li key={lib.id}>
                <Link
                  className="block rounded-md border border-border p-3 hover:bg-surface"
                  to={`/libraries/${lib.id}`}
                >
                  <div className="font-medium">{lib.name}</div>
                  <div className="text-xs text-muted">
                    {t(`kinds.${lib.kind}`)} · {lib.paths.length} {t('list.paths').toLowerCase()}
                  </div>
                </Link>
              </li>
            ))}
          </ul>
        </CardContent>
      </Card>
    </section>
  );
}
