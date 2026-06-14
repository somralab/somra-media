import { useMutation, useQuery } from '@tanstack/react-query';
import {
  getWatchState,
  saveWatchState,
  startPlayback,
  stopPlayback,
  type PlayRequest,
  type PlayResponse,
} from '@/api/endpoints/streaming';

export function usePlayback(itemId: number) {
  return useMutation({
    mutationFn: (body: PlayRequest = {}) => startPlayback(itemId, body),
  });
}

export function useStopPlayback() {
  return useMutation({ mutationFn: (sessionId: string) => stopPlayback(sessionId) });
}

export function useWatchStateQuery(itemId: number, enabled = true) {
  return useQuery({
    queryKey: ['watch-state', itemId],
    queryFn: () => getWatchState(itemId),
    enabled: enabled && itemId > 0,
    retry: false,
  });
}

export function useSaveWatchProgress(itemId: number) {
  return useMutation({
    mutationFn: (positionMs: number) =>
      saveWatchState(itemId, { positionMs, completed: false }),
  });
}

export type { PlayResponse };
