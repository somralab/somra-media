import { apiClient } from '../client';
import type { components } from '../generated/openapi';

export type AutomationDownload = components['schemas']['AutomationDownload'];
export type QualityProfile = components['schemas']['QualityProfile'];
export type QualityProfileInput = components['schemas']['QualityProfileInput'];
export type QualityProfilePatch = components['schemas']['QualityProfilePatch'];
export type AutomationMonitor = components['schemas']['AutomationMonitor'];
export type AutomationMonitorInput = components['schemas']['AutomationMonitorInput'];
export type AutomationMonitorPatch = components['schemas']['AutomationMonitorPatch'];
export type IndexerSearchRequest = components['schemas']['IndexerSearchRequest'];

export function listAutomationDownloads(
  signal?: AbortSignal,
): Promise<{ downloads: AutomationDownload[] }> {
  return apiClient.fetch<{ downloads: AutomationDownload[] }>(
    '/automation/downloads',
    signal !== undefined ? { signal } : {},
  );
}

export function getAutomationDownload(
  downloadId: number,
  signal?: AbortSignal,
): Promise<AutomationDownload> {
  return apiClient.fetch<AutomationDownload>(
    `/automation/downloads/${downloadId}`,
    signal !== undefined ? { signal } : {},
  );
}

export function listQualityProfiles(signal?: AbortSignal): Promise<{ profiles: QualityProfile[] }> {
  return apiClient.fetch<{ profiles: QualityProfile[] }>(
    '/automation/quality-profiles',
    signal !== undefined ? { signal } : {},
  );
}

export function getQualityProfile(
  profileId: number,
  signal?: AbortSignal,
): Promise<QualityProfile> {
  return apiClient.fetch<QualityProfile>(
    `/automation/quality-profiles/${profileId}`,
    signal !== undefined ? { signal } : {},
  );
}

export function createQualityProfile(body: QualityProfileInput): Promise<{ id: number }> {
  return apiClient.fetch<{ id: number }>('/automation/quality-profiles', {
    method: 'POST',
    body,
  });
}

export function patchQualityProfile(
  profileId: number,
  body: QualityProfilePatch,
): Promise<QualityProfile> {
  return apiClient.fetch<QualityProfile>(`/automation/quality-profiles/${profileId}`, {
    method: 'PATCH',
    body,
  });
}

export function listAutomationMonitors(
  signal?: AbortSignal,
): Promise<{ monitors: AutomationMonitor[] }> {
  return apiClient.fetch<{ monitors: AutomationMonitor[] }>(
    '/automation/monitors',
    signal !== undefined ? { signal } : {},
  );
}

export function getAutomationMonitor(
  monitorId: number,
  signal?: AbortSignal,
): Promise<AutomationMonitor> {
  return apiClient.fetch<AutomationMonitor>(
    `/automation/monitors/${monitorId}`,
    signal !== undefined ? { signal } : {},
  );
}

export function createAutomationMonitor(body: AutomationMonitorInput): Promise<AutomationMonitor> {
  return apiClient.fetch<AutomationMonitor>('/automation/monitors', { method: 'POST', body });
}

export function patchAutomationMonitor(
  monitorId: number,
  body: AutomationMonitorPatch,
): Promise<AutomationMonitor> {
  return apiClient.fetch<AutomationMonitor>(`/automation/monitors/${monitorId}`, {
    method: 'PATCH',
    body,
  });
}

export function deleteAutomationMonitor(monitorId: number): Promise<void> {
  return apiClient.fetch<void>(`/automation/monitors/${monitorId}`, { method: 'DELETE' });
}
