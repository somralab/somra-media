import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  createPluginInstance,
  deletePluginInstance,
  getPluginInstance,
  listPluginCatalog,
  listPluginInstances,
  patchPluginInstance,
  testPluginInstance,
  type PluginInstanceInput,
  type PluginInstancePatch,
  type PluginType,
} from '@/api/endpoints/plugins';

const CATALOG_KEY = ['plugins', 'catalog'] as const;
const INSTANCES_KEY = ['plugins', 'instances'] as const;

export function usePluginCatalog(enabled = true) {
  return useQuery({
    queryKey: CATALOG_KEY,
    queryFn: ({ signal }) => listPluginCatalog(signal),
    enabled,
  });
}

export function usePluginInstances(enabled = true) {
  return useQuery({
    queryKey: INSTANCES_KEY,
    queryFn: ({ signal }) => listPluginInstances(signal),
    enabled,
  });
}

export function usePluginInstancesByType(pluginType: PluginType, enabled = true) {
  const query = usePluginInstances(enabled);
  const instances = (query.data?.instances ?? []).filter((i) => i.pluginType === pluginType);
  return { ...query, instances };
}

export function usePluginInstance(instanceId: number, enabled = true) {
  return useQuery({
    queryKey: [...INSTANCES_KEY, instanceId],
    queryFn: ({ signal }) => getPluginInstance(instanceId, signal),
    enabled: enabled && instanceId > 0,
  });
}

export function useCreatePluginInstance() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: PluginInstanceInput) => createPluginInstance(body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: INSTANCES_KEY }),
  });
}

export function usePatchPluginInstance() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body: PluginInstancePatch }) =>
      patchPluginInstance(id, body),
    onSuccess: (_data, { id }) => {
      void qc.invalidateQueries({ queryKey: INSTANCES_KEY });
      void qc.invalidateQueries({ queryKey: [...INSTANCES_KEY, id] });
    },
  });
}

export function useDeletePluginInstance() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deletePluginInstance(id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: INSTANCES_KEY }),
  });
}

export function useTestPluginInstance() {
  return useMutation({
    mutationFn: (id: number) => testPluginInstance(id),
  });
}
