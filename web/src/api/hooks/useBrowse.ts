import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  addFavorite,
  addWatchlist,
  getDiscoverHome,
  getMediaDetail,
  listMediaItemsPaginated,
  removeFavorite,
  removeWatchlist,
  searchMedia,
  type BrowseFilters,
} from '@/api/endpoints/browse';

export function useDiscoverHome() {
  return useQuery({
    queryKey: ['discover', 'home'],
    queryFn: ({ signal }) => getDiscoverHome(signal),
  });
}

export function useSearchMedia(query: string, enabled = true) {
  return useQuery({
    queryKey: ['search', query],
    queryFn: ({ signal }) => searchMedia(query, 20, signal),
    enabled: enabled && query.trim().length > 0,
  });
}

export function useBrowseItems(libraryId: number, filters: BrowseFilters) {
  return useQuery({
    queryKey: ['libraries', libraryId, 'items', filters],
    queryFn: ({ signal }) => listMediaItemsPaginated(libraryId, filters, signal),
    enabled: libraryId > 0,
  });
}

export function useMediaDetail(itemId: number) {
  return useQuery({
    queryKey: ['media-items', itemId, 'detail'],
    queryFn: ({ signal }) => getMediaDetail(itemId, signal),
    enabled: itemId > 0,
  });
}

export function useFavoriteToggle(itemId: number, isFavorite: boolean) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => (isFavorite ? removeFavorite(itemId) : addFavorite(itemId)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['media-items', itemId, 'detail'] });
      qc.invalidateQueries({ queryKey: ['favorites'] });
    },
  });
}

export function useWatchlistToggle(itemId: number, inWatchlist: boolean) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => (inWatchlist ? removeWatchlist(itemId) : addWatchlist(itemId)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['media-items', itemId, 'detail'] });
      qc.invalidateQueries({ queryKey: ['watchlist'] });
    },
  });
}
