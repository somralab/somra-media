import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  autoMatchLibrary,
  createLibrary,
  deleteLibrary,
  getLibrary,
  listLibraries,
  listMatchCandidates,
  listMediaItems,
  listScans,
  rematchItem,
  triggerScan,
  updateLibrary,
  type LibraryInput,
} from '@/api/endpoints/library';

export function useLibraries() {
  return useQuery({ queryKey: ['libraries'], queryFn: ({ signal }) => listLibraries(signal) });
}

export function useLibrary(id: number) {
  return useQuery({
    queryKey: ['libraries', id],
    queryFn: ({ signal }) => getLibrary(id, signal),
    enabled: id > 0,
  });
}

export function useCreateLibrary() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: LibraryInput) => createLibrary(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['libraries'] }),
  });
}

export function useUpdateLibrary(id: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: LibraryInput) => updateLibrary(id, input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['libraries'] });
      qc.invalidateQueries({ queryKey: ['libraries', id] });
    },
  });
}

export function useDeleteLibrary() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteLibrary(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['libraries'] }),
  });
}

export function useTriggerScan(libraryId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (type: 'full' | 'incremental') => triggerScan(libraryId, type),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['libraries', libraryId, 'scans'] }),
  });
}

export function useScanHistory(libraryId: number) {
  return useQuery({
    queryKey: ['libraries', libraryId, 'scans'],
    queryFn: ({ signal }) => listScans(libraryId, signal),
    enabled: libraryId > 0,
  });
}

export function useMediaItems(libraryId: number) {
  return useQuery({
    queryKey: ['libraries', libraryId, 'items'],
    queryFn: ({ signal }) => listMediaItems(libraryId, signal),
    enabled: libraryId > 0,
  });
}

export function useMatchCandidates(itemId: number) {
  return useQuery({
    queryKey: ['media-items', itemId, 'candidates'],
    queryFn: ({ signal }) => listMatchCandidates(itemId, signal),
    enabled: itemId > 0,
  });
}

export function useRematchItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      itemId,
      provider,
      externalId,
    }: {
      itemId: number;
      provider: string;
      externalId: string;
    }) => rematchItem(itemId, provider, externalId),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['media-items', vars.itemId] });
      qc.invalidateQueries({ queryKey: ['libraries'] });
    },
  });
}

export function useAutoMatch(libraryId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => autoMatchLibrary(libraryId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['libraries', libraryId, 'items'] }),
  });
}
