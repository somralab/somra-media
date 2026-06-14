import { useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useServerEvents } from '@/api/events';
import {
  approveRequest,
  cancelRequest,
  createRequest,
  getRequestPolicies,
  listRequests,
  patchRequestPolicies,
  rejectRequest,
  type CreateRequestInput,
  type ListRequestsParams,
  type RequestActionInput,
  type RequestPoliciesPatch,
} from '@/api/endpoints/requests';

const REQUESTS_KEY = ['requests'] as const;
const POLICIES_KEY = ['requests', 'policies'] as const;

const REQUEST_EVENT_NAMES = [
  'request.created',
  'request.approved',
  'request.rejected',
  'request.completed',
  'request.cancelled',
] as const;

export function useRequests(params: ListRequestsParams = {}, enabled = true) {
  return useQuery({
    queryKey: [...REQUESTS_KEY, params],
    queryFn: ({ signal }) => listRequests(params, signal),
    enabled,
    refetchInterval: (query) => (query.state.data !== undefined ? 30_000 : false),
  });
}

export function useCreateRequest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateRequestInput) => createRequest(body),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: REQUESTS_KEY });
    },
  });
}

export function useApproveRequest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body?: RequestActionInput }) =>
      approveRequest(id, body ?? {}),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: REQUESTS_KEY });
    },
  });
}

export function useRejectRequest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body?: RequestActionInput }) =>
      rejectRequest(id, body ?? {}),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: REQUESTS_KEY });
    },
  });
}

export function useCancelRequest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => cancelRequest(id),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: REQUESTS_KEY });
    },
  });
}

export function useRequestPolicies(enabled = true) {
  return useQuery({
    queryKey: POLICIES_KEY,
    queryFn: ({ signal }) => getRequestPolicies(signal),
    enabled,
  });
}

export function usePatchRequestPolicies() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: RequestPoliciesPatch) => patchRequestPolicies(body),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: POLICIES_KEY });
    },
  });
}

/**
 * Subscribes to request lifecycle SSE events and invalidates the requests cache
 * when updates arrive. Falls back to polling via useRequests refetchInterval.
 */
export function useRequestRealtimeSync(enabled = true): { connected: boolean } {
  const qc = useQueryClient();
  const { connected, lastEvent } = useServerEvents({
    enabled,
    eventNames: REQUEST_EVENT_NAMES,
  });

  useEffect(() => {
    if (!lastEvent) return;
    if (!REQUEST_EVENT_NAMES.includes(lastEvent.type as (typeof REQUEST_EVENT_NAMES)[number])) {
      return;
    }
    void qc.invalidateQueries({ queryKey: REQUESTS_KEY });
  }, [lastEvent, qc]);

  return { connected };
}
