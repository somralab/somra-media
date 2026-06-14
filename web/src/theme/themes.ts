export const THEME_IDS = ['cinematic', 'aurora', 'noir', 'minimal'] as const;

export type ThemeId = (typeof THEME_IDS)[number];

export const DEFAULT_THEME: ThemeId = 'cinematic';

export interface ThemeDescriptor {
  id: ThemeId;
  labelKey: string;
}

export const THEMES: readonly ThemeDescriptor[] = THEME_IDS.map((id) => ({
  id,
  labelKey: `settings.theme.options.${id}`,
}));

export function isThemeId(value: unknown): value is ThemeId {
  return typeof value === 'string' && (THEME_IDS as readonly string[]).includes(value);
}
