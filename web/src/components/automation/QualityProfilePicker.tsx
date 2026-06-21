import { type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import type { QualityProfile } from '@/api/endpoints/automation';

export interface QualityProfilePickerProps {
  profiles: QualityProfile[];
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
}

export function QualityProfilePicker({
  profiles,
  value,
  onChange,
  disabled = false,
}: QualityProfilePickerProps): ReactNode {
  const { t } = useTranslation('automation');

  return (
    <label className="block space-y-1 text-sm">
      <span>{t('profilePicker.label')}</span>
      <select
        className="w-full rounded-md border border-border bg-surface px-3 py-2"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
      >
        <option value="">{t('profilePicker.none')}</option>
        {profiles.map((profile) => (
          <option key={profile.id} value={profile.name ?? ''}>
            {profile.name}
            {profile.isDefault ? ' *' : ''}
          </option>
        ))}
      </select>
    </label>
  );
}
