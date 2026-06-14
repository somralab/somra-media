import { apiClient } from '../client';
import type { MediaItem } from './library';

export interface WatchStateBrief {
  positionMs: number;
  completed: boolean;
}

export interface MediaItemSummary extends MediaItem {
  watchState?: WatchStateBrief;
}

export interface DiscoverShelf {
  id: string;
  titleKey: string;
  items: MediaItemSummary[];
}

export interface DiscoverHome {
  shelves: DiscoverShelf[];
}

export interface SearchResult extends MediaItem {
  score?: number;
}

export interface SearchResponse {
  results: SearchResult[];
  query: string;
}

export interface PaginatedMediaItems {
  items: MediaItem[];
  total: number;
  offset: number;
  limit: number;
}

export interface BrowseFilters {
  offset?: number;
  limit?: number;
  sort?: 'title' | 'year' | 'created_at';
  genre?: string;
  year?: number;
  watchStatus?: 'all' | 'unwatched' | 'in_progress' | 'completed';
}

export interface CastMember {
  name: string;
  role: string;
  order: number;
  imageUrl?: string;
}

export interface ArtworkImage {
  kind: string;
  sourceUrl?: string;
  localPath?: string;
  width?: number;
  height?: number;
}

export interface EpisodeSummary {
  id: number;
  seasonNumber: number;
  episodeNumber: number;
  title?: string;
}

export interface SeasonDetail {
  seasonNumber: number;
  episodes: EpisodeSummary[];
}

export interface MediaDetail extends MediaItem {
  genres: string[];
  backdropUrl?: string;
  cast: CastMember[];
  images: ArtworkImage[];
  seasons?: SeasonDetail[];
  isFavorite: boolean;
  inWatchlist: boolean;
  watchState?: WatchStateBrief;
  contentRating?: string | null;
}

export function getDiscoverHome(signal?: AbortSignal): Promise<DiscoverHome> {
  return apiClient.fetch<DiscoverHome>('/discover/home', signal !== undefined ? { signal } : {});
}

export function searchMedia(q: string, limit = 20, signal?: AbortSignal): Promise<SearchResponse> {
  const params = new URLSearchParams({ q, limit: String(limit) });
  return apiClient.fetch<SearchResponse>(`/search?${params}`, signal !== undefined ? { signal } : {});
}

export function listMediaItemsPaginated(
  libraryId: number,
  filters: BrowseFilters = {},
  signal?: AbortSignal,
): Promise<PaginatedMediaItems> {
  const params = new URLSearchParams();
  if (filters.offset !== undefined) params.set('offset', String(filters.offset));
  if (filters.limit !== undefined) params.set('limit', String(filters.limit));
  if (filters.sort) params.set('sort', filters.sort);
  if (filters.genre) params.set('genre', filters.genre);
  if (filters.year !== undefined) params.set('year', String(filters.year));
  if (filters.watchStatus) params.set('watchStatus', filters.watchStatus);
  const qs = params.toString();
  return apiClient.fetch<PaginatedMediaItems>(
    `/libraries/${libraryId}/items${qs ? `?${qs}` : ''}`,
    signal !== undefined ? { signal } : {},
  );
}

export function getMediaDetail(itemId: number, signal?: AbortSignal): Promise<MediaDetail> {
  return apiClient.fetch<MediaDetail>(
    `/media-items/${itemId}/detail`,
    signal !== undefined ? { signal } : {},
  );
}

export interface WatchState {
  mediaItemId: number;
  positionMs: number;
  completed: boolean;
}

export function listWatchStates(signal?: AbortSignal): Promise<WatchState[]> {
  return apiClient.fetch<WatchState[]>('/watch-state', signal !== undefined ? { signal } : {});
}

export function upsertWatchState(
  mediaItemId: number,
  body: { positionMs: number; completed: boolean },
): Promise<WatchState> {
  return apiClient.fetch<WatchState>(`/watch-state/${mediaItemId}`, { method: 'PUT', body });
}

export function listFavorites(signal?: AbortSignal): Promise<number[]> {
  return apiClient.fetch<number[]>('/favorites', signal !== undefined ? { signal } : {});
}

export function addFavorite(mediaItemId: number): Promise<void> {
  return apiClient.fetch<void>(`/favorites/${mediaItemId}`, { method: 'POST' });
}

export function removeFavorite(mediaItemId: number): Promise<void> {
  return apiClient.fetch<void>(`/favorites/${mediaItemId}`, { method: 'DELETE' });
}

export function listWatchlist(signal?: AbortSignal): Promise<number[]> {
  return apiClient.fetch<number[]>('/watchlist', signal !== undefined ? { signal } : {});
}

export function addWatchlist(mediaItemId: number): Promise<void> {
  return apiClient.fetch<void>(`/watchlist/${mediaItemId}`, { method: 'POST' });
}

export function removeWatchlist(mediaItemId: number): Promise<void> {
  return apiClient.fetch<void>(`/watchlist/${mediaItemId}`, { method: 'DELETE' });
}
