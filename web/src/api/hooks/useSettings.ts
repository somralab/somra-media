import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getSettings, patchSettingsCategory } from '@/api/endpoints/settings';

export function useSettings(enabled = true) {
  return useQuery({
    queryKey: ['settings'],
    queryFn: getSettings,
    enabled,
  });
}

export function usePatchSettings(category: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (patch: Record<string, unknown>) => patchSettingsCategory(category, patch),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['settings'] });
    },
  });
}
