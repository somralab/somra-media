import { apiClient } from '../client';
import type { components } from '../generated/openapi';

export type OnboardingPhase = components['schemas']['OnboardingPhase'];
export type OnboardingState = components['schemas']['OnboardingState'];
export type SystemProfile = components['schemas']['SystemProfile'];
export type SmartDefaults = components['schemas']['SmartDefaults'];
export type SettingsSnapshot = components['schemas']['SettingsSnapshot'];

export interface SetupStatus {
  setupRequired: boolean;
  completed: boolean;
  phase: OnboardingPhase;
}

export async function getOnboardingStatus(): Promise<OnboardingState> {
  return apiClient.fetch<OnboardingState>('/onboarding/status');
}

export async function advanceOnboardingStep(body: {
  phase: OnboardingPhase;
  locale?: string;
  applyDefaults?: boolean;
  libraryId?: number;
}): Promise<OnboardingState> {
  return apiClient.fetch<OnboardingState>('/onboarding/step', { method: 'POST', body });
}

export async function completeOnboarding(): Promise<void> {
  await apiClient.fetch<void>('/onboarding/complete', { method: 'POST' });
}

export async function detectSystem(paths?: string): Promise<SystemProfile> {
  const qs = paths ? `?paths=${encodeURIComponent(paths)}` : '';
  return apiClient.fetch<SystemProfile>(`/system/detect${qs}`);
}

export async function getSettings(): Promise<SettingsSnapshot> {
  return apiClient.fetch<SettingsSnapshot>('/settings');
}

export async function patchSettingsCategory(
  category: string,
  patch: Record<string, unknown>,
): Promise<Record<string, unknown>> {
  return apiClient.fetch<Record<string, unknown>>(`/settings/${category}`, {
    method: 'PATCH',
    body: patch,
  });
}
