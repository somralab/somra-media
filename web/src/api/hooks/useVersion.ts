import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { getVersion, type VersionResponse } from '@/api/endpoints/system';
import { ApiError } from '@/api/ApiError';

export const VERSION_QUERY_KEY = ['system', 'version'] as const;

export function useVersion(): UseQueryResult<VersionResponse, ApiError> {
  return useQuery<VersionResponse, ApiError>({
    queryKey: VERSION_QUERY_KEY,
    queryFn: ({ signal }) => getVersion(signal),
    staleTime: Number.POSITIVE_INFINITY,
    gcTime: Number.POSITIVE_INFINITY,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    refetchInterval: false,
    retry: 1,
  });
}
