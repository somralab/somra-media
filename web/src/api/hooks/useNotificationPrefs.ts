import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  getNotificationPreferences,
  patchNotificationPreferences,
  type NotificationPreferencesPatch,
} from '@/api/endpoints/notifications';

const PREFS_KEY = ['notifications', 'preferences'] as const;

export function useNotificationPrefs(enabled = true) {
  return useQuery({
    queryKey: PREFS_KEY,
    queryFn: ({ signal }) => getNotificationPreferences(signal),
    enabled,
  });
}

export function usePatchNotificationPrefs() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: NotificationPreferencesPatch) => patchNotificationPreferences(body),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: PREFS_KEY });
    },
  });
}
