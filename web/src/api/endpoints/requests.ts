import { apiClient } from '../client';
import type { components } from '../generated/openapi';

export type ContentRequest = components['schemas']['ContentRequest'];
export type CreateRequestInput = components['schemas']['CreateRequestInput'];
export type RequestDiscoverResult = components['schemas']['RequestDiscoverResult'];
export type RequestPolicies = components['schemas']['RequestPolicies'];
export type RequestPoliciesPatch = components['schemas']['RequestPoliciesPatch'];
export type RequestStatus = components['schemas']['RequestStatus'];
export type RequestMediaKind = components['schemas']['RequestMediaKind'];
export type RequestQualityResolution = components['schemas']['RequestQualityResolution'];
export type RequestActionInput = components['schemas']['RequestActionInput'];

export interface ListRequestsParams {
  status?: RequestStatus;
  userId?: string;
  limit?: number;
  offset?: number;
}

export interface DiscoverParams {
  q: string;
  kind?: RequestMediaKind;
}

export function listRequests(
  params: ListRequestsParams = {},
  signal?: AbortSignal,
): Promise<{ requests: ContentRequest[] }> {
  const search = new URLSearchParams();
  if (params.status) search.set('status', params.status);
  if (params.userId) search.set('userId', params.userId);
  if (params.limit !== undefined) search.set('limit', String(params.limit));
  if (params.offset !== undefined) search.set('offset', String(params.offset));
  const qs = search.toString();
  return apiClient.fetch<{ requests: ContentRequest[] }>(
    `/requests${qs ? `?${qs}` : ''}`,
    signal !== undefined ? { signal } : {},
  );
}

export function discoverRequestableContent(
  params: DiscoverParams,
  signal?: AbortSignal,
): Promise<{ results: RequestDiscoverResult[] }> {
  const search = new URLSearchParams({ q: params.q });
  if (params.kind) search.set('kind', params.kind);
  return apiClient.fetch<{ results: RequestDiscoverResult[] }>(
    `/requests/discover?${search}`,
    signal !== undefined ? { signal } : {},
  );
}

export function createRequest(body: CreateRequestInput): Promise<ContentRequest> {
  return apiClient.fetch<ContentRequest>('/requests', { method: 'POST', body });
}

export function getRequest(requestId: number, signal?: AbortSignal): Promise<ContentRequest> {
  return apiClient.fetch<ContentRequest>(
    `/requests/${requestId}`,
    signal !== undefined ? { signal } : {},
  );
}

export function approveRequest(
  requestId: number,
  body: RequestActionInput = {},
): Promise<ContentRequest> {
  return apiClient.fetch<ContentRequest>(`/requests/${requestId}/approve`, {
    method: 'POST',
    body,
  });
}

export function rejectRequest(
  requestId: number,
  body: RequestActionInput = {},
): Promise<ContentRequest> {
  return apiClient.fetch<ContentRequest>(`/requests/${requestId}/reject`, {
    method: 'POST',
    body,
  });
}

export function cancelRequest(requestId: number): Promise<ContentRequest> {
  return apiClient.fetch<ContentRequest>(`/requests/${requestId}/cancel`, { method: 'POST' });
}

export function getRequestPolicies(signal?: AbortSignal): Promise<RequestPolicies> {
  return apiClient.fetch<RequestPolicies>(
    '/requests/policies',
    signal !== undefined ? { signal } : {},
  );
}

export function patchRequestPolicies(body: RequestPoliciesPatch): Promise<RequestPolicies> {
  return apiClient.fetch<RequestPolicies>('/requests/policies', { method: 'PATCH', body });
}
