import { type ReactNode } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import type { SearchResult } from '@/api/endpoints/browse';

export interface SearchResultsDropdownProps {
  results: SearchResult[];
  query: string;
  isLoading?: boolean;
  onClose?: () => void;
}

export function SearchResultsDropdown({
  results,
  query,
  isLoading,
  onClose,
}: SearchResultsDropdownProps): ReactNode {
  const { t } = useTranslation('search');

  if (!query.trim()) {
    return null;
  }

  return (
    <div
      className="absolute left-0 right-0 top-full z-50 mt-1 max-h-80 overflow-auto rounded-md border border-border bg-surface shadow-lg"
      role="listbox"
      aria-label={t('results.label')}
    >
      {isLoading ? (
        <p className="p-3 text-sm text-muted">{t('results.loading')}</p>
      ) : results.length === 0 ? (
        <p className="p-3 text-sm text-muted">{t('results.empty', { query })}</p>
      ) : (
        results.map((item) => (
          <Link
            key={item.id}
            to={`/libraries/${item.libraryId}/items/${item.id}`}
            className="flex items-center gap-3 border-b border-border p-3 last:border-b-0 hover:bg-bg"
            role="option"
            onClick={onClose}
          >
            {item.posterUrl ? (
              <img src={item.posterUrl} alt="" className="h-12 w-8 rounded object-cover" />
            ) : (
              <div className="flex h-12 w-8 items-center justify-center rounded bg-border text-xs text-muted">
                ?
              </div>
            )}
            <div className="min-w-0">
              <p className="truncate text-sm font-medium">{item.title}</p>
              {item.year ? <p className="text-xs text-muted">{item.year}</p> : null}
            </div>
          </Link>
        ))
      )}
    </div>
  );
}
