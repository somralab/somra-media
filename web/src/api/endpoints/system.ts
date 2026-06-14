import type { paths } from '@/api/generated/openapi';
import { apiFetch } from '@/api/client';

export type HealthResponse =
  paths['/health']['get']['responses']['200']['content']['application/json'];

export type VersionResponse =
  paths['/version']['get']['responses']['200']['content']['application/json'];

export function getHealth(signal?: AbortSignal): Promise<HealthResponse> {
  return apiFetch<HealthResponse>('/health', signal !== undefined ? { signal } : {});
}

export function getVersion(signal?: AbortSignal): Promise<VersionResponse> {
  return apiFetch<VersionResponse>('/version', signal !== undefined ? { signal } : {});
}
