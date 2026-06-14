import { type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import type { RequestQualityResolution } from '@/api/endpoints/requests';

const QUALITY_OPTIONS: RequestQualityResolution[] = ['1080p', '720p', 'any'];

export interface QualitySelectorProps {
  value: RequestQualityResolution;
  onChange: (value: RequestQualityResolution) => void;
  disabled?: boolean;
}

export function QualitySelector({
  value,
  onChange,
  disabled = false,
}: QualitySelectorProps): ReactNode {
  const { t } = useTranslation('requests');

  return (
    <fieldset className="space-y-2" disabled={disabled}>
      <legend className="text-sm font-medium">{t('quality.label')}</legend>
      <div className="flex flex-wrap gap-2">
        {QUALITY_OPTIONS.map((option) => (
          <label
            key={option}
            className="flex cursor-pointer items-center gap-2 rounded-md border border-border px-3 py-2 text-sm has-[:checked]:border-primary has-[:checked]:bg-primary/10"
          >
            <input
              type="radio"
              name="qualityResolution"
              value={option}
              checked={value === option}
              onChange={() => onChange(option)}
              className="accent-primary"
            />
            {t(`quality.options.${option}`)}
          </label>
        ))}
      </div>
    </fieldset>
  );
}
