import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  createAutomationMonitor,
  createQualityProfile,
  deleteAutomationMonitor,
  getAutomationDownload,
  getAutomationMonitor,
  getQualityProfile,
  listAutomationDownloads,
  listAutomationMonitors,
  listQualityProfiles,
  patchAutomationMonitor,
  patchQualityProfile,
  type AutomationMonitorInput,
  type AutomationMonitorPatch,
  type QualityProfileInput,
  type QualityProfilePatch,
} from '@/api/endpoints/automation';

const DOWNLOADS_KEY = ['automation', 'downloads'] as const;
const PROFILES_KEY = ['automation', 'quality-profiles'] as const;
const MONITORS_KEY = ['automation', 'monitors'] as const;

const DOWNLOAD_POLL_MS = 5_000;

export function useAutomationDownloads(enabled = true) {
  return useQuery({
    queryKey: DOWNLOADS_KEY,
    queryFn: ({ signal }) => listAutomationDownloads(signal),
    enabled,
    refetchInterval: DOWNLOAD_POLL_MS,
  });
}

export function useAutomationDownload(downloadId: number, enabled = true) {
  return useQuery({
    queryKey: [...DOWNLOADS_KEY, downloadId],
    queryFn: ({ signal }) => getAutomationDownload(downloadId, signal),
    enabled: enabled && downloadId > 0,
    refetchInterval: DOWNLOAD_POLL_MS,
  });
}

export function useQualityProfiles(enabled = true) {
  return useQuery({
    queryKey: PROFILES_KEY,
    queryFn: ({ signal }) => listQualityProfiles(signal),
    enabled,
  });
}

export function useQualityProfile(profileId: number, enabled = true) {
  return useQuery({
    queryKey: [...PROFILES_KEY, profileId],
    queryFn: ({ signal }) => getQualityProfile(profileId, signal),
    enabled: enabled && profileId > 0,
  });
}

export function useCreateQualityProfile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: QualityProfileInput) => createQualityProfile(body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: PROFILES_KEY }),
  });
}

export function usePatchQualityProfile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body: QualityProfilePatch }) =>
      patchQualityProfile(id, body),
    onSuccess: (_data, { id }) => {
      void qc.invalidateQueries({ queryKey: PROFILES_KEY });
      void qc.invalidateQueries({ queryKey: [...PROFILES_KEY, id] });
    },
  });
}

export function useAutomationMonitors(enabled = true) {
  return useQuery({
    queryKey: MONITORS_KEY,
    queryFn: ({ signal }) => listAutomationMonitors(signal),
    enabled,
  });
}

export function useAutomationMonitor(monitorId: number, enabled = true) {
  return useQuery({
    queryKey: [...MONITORS_KEY, monitorId],
    queryFn: ({ signal }) => getAutomationMonitor(monitorId, signal),
    enabled: enabled && monitorId > 0,
  });
}

export function useCreateAutomationMonitor() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: AutomationMonitorInput) => createAutomationMonitor(body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: MONITORS_KEY }),
  });
}

export function usePatchAutomationMonitor() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body: AutomationMonitorPatch }) =>
      patchAutomationMonitor(id, body),
    onSuccess: (_data, { id }) => {
      void qc.invalidateQueries({ queryKey: MONITORS_KEY });
      void qc.invalidateQueries({ queryKey: [...MONITORS_KEY, id] });
    },
  });
}

export function useDeleteAutomationMonitor() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteAutomationMonitor(id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: MONITORS_KEY }),
  });
}
