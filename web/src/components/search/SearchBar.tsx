import { type FormEvent, type ReactNode, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Input } from '@/components/ui/Input';
import { cn } from '@/lib/cn';

export interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
  onSubmit?: (value: string) => void;
  className?: string;
  debounceMs?: number;
}

export function SearchBar({
  value,
  onChange,
  onSubmit,
  className,
  debounceMs = 300,
}: SearchBarProps): ReactNode {
  const { t } = useTranslation('search');
  const [local, setLocal] = useState(value);

  useEffect(() => {
    setLocal(value);
  }, [value]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      if (local !== value) {
        onChange(local);
      }
    }, debounceMs);
    return () => window.clearTimeout(timer);
  }, [local, value, onChange, debounceMs]);

  function handleSubmit(e: FormEvent): void {
    e.preventDefault();
    onSubmit?.(local);
  }

  return (
    <form onSubmit={handleSubmit} className={cn('relative w-full max-w-xl', className)} role="search">
      <Input
        type="search"
        value={local}
        onChange={(e) => setLocal(e.target.value)}
        placeholder={t('placeholder')}
        aria-label={t('label')}
        className="w-full"
      />
    </form>
  );
}
