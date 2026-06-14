import { apiClient } from '../client';
import type { components } from '../generated/openapi';

export type NotificationPreference = components['schemas']['NotificationPreference'];
export type NotificationPreferencesPatch = components['schemas']['NotificationPreferencesPatch'];
export type NotificationChannel = components['schemas']['NotificationChannel'];

export function getNotificationPreferences(
  signal?: AbortSignal,
): Promise<{ preferences: NotificationPreference[] }> {
  return apiClient.fetch<{ preferences: NotificationPreference[] }>(
    '/notifications/preferences',
    signal !== undefined ? { signal } : {},
  );
}

export function patchNotificationPreferences(
  body: NotificationPreferencesPatch,
): Promise<{ preferences: NotificationPreference[] }> {
  return apiClient.fetch<{ preferences: NotificationPreference[] }>('/notifications/preferences', {
    method: 'PATCH',
    body,
  });
}

export function listNotificationChannels(
  signal?: AbortSignal,
): Promise<{ channels: NotificationChannel[] }> {
  return apiClient.fetch<{ channels: NotificationChannel[] }>(
    '/notifications/channels',
    signal !== undefined ? { signal } : {},
  );
}
