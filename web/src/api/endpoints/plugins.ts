import { apiClient } from '../client';
import type { components } from '../generated/openapi';

export type PluginCatalogEntry = components['schemas']['PluginCatalogEntry'];
export type PluginInstance = components['schemas']['PluginInstance'];
export type PluginInstanceInput = components['schemas']['PluginInstanceInput'];
export type PluginInstancePatch = components['schemas']['PluginInstancePatch'];
export type PluginTestResult = components['schemas']['PluginTestResult'];
export type PluginType = components['schemas']['PluginType'];

export function listPluginCatalog(signal?: AbortSignal): Promise<{ catalog: PluginCatalogEntry[] }> {
  return apiClient.fetch<{ catalog: PluginCatalogEntry[] }>(
    '/plugins/catalog',
    signal !== undefined ? { signal } : {},
  );
}

export function listPluginInstances(signal?: AbortSignal): Promise<{ instances: PluginInstance[] }> {
  return apiClient.fetch<{ instances: PluginInstance[] }>(
    '/plugins/instances',
    signal !== undefined ? { signal } : {},
  );
}

export function getPluginInstance(instanceId: number, signal?: AbortSignal): Promise<PluginInstance> {
  return apiClient.fetch<PluginInstance>(
    `/plugins/instances/${instanceId}`,
    signal !== undefined ? { signal } : {},
  );
}

export function createPluginInstance(body: PluginInstanceInput): Promise<PluginInstance> {
  return apiClient.fetch<PluginInstance>('/plugins/instances', { method: 'POST', body });
}

export function patchPluginInstance(
  instanceId: number,
  body: PluginInstancePatch,
): Promise<PluginInstance> {
  return apiClient.fetch<PluginInstance>(`/plugins/instances/${instanceId}`, {
    method: 'PATCH',
    body,
  });
}

export function deletePluginInstance(instanceId: number): Promise<void> {
  return apiClient.fetch<void>(`/plugins/instances/${instanceId}`, { method: 'DELETE' });
}

export function testPluginInstance(instanceId: number): Promise<PluginTestResult> {
  return apiClient.fetch<PluginTestResult>(`/plugins/instances/${instanceId}/test`, {
    method: 'POST',
  });
}
