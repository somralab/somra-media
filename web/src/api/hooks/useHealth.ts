import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { getHealth, type HealthResponse } from '@/api/endpoints/system';
import { ApiError } from '@/api/ApiError';

export const HEALTH_QUERY_KEY = ['system', 'health'] as const;

const DEFAULT_REFETCH_INTERVAL_MS = 15_000;

export interface UseHealthOptions {
  /** Override the default 15s polling interval. Pass `false` to disable. */
  refetchInterval?: number | false;
  enabled?: boolean;
  retry?: number | boolean;
}

export function useHealth(
  options: UseHealthOptions = {},
): UseQueryResult<HealthResponse, ApiError> {
  const { refetchInterval = DEFAULT_REFETCH_INTERVAL_MS, enabled = true, retry = 1 } = options;
  return useQuery<HealthResponse, ApiError>({
    queryKey: HEALTH_QUERY_KEY,
    queryFn: ({ signal }) => getHealth(signal),
    refetchInterval,
    refetchOnWindowFocus: true,
    retry,
    enabled,
  });
}
