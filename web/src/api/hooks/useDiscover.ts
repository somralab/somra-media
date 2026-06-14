import { useQuery } from '@tanstack/react-query';
import {
  discoverRequestableContent,
  type DiscoverParams,
  type RequestMediaKind,
} from '@/api/endpoints/requests';

export function useDiscover(query: string, kind?: RequestMediaKind, enabled = true) {
  const trimmed = query.trim();
  return useQuery({
    queryKey: ['requests', 'discover', trimmed, kind ?? 'all'],
    queryFn: ({ signal }) =>
      discoverRequestableContent(
        { q: trimmed, ...(kind ? { kind } : {}) } satisfies DiscoverParams,
        signal,
      ),
    enabled: enabled && trimmed.length > 0,
  });
}
