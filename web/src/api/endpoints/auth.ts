import { apiClient } from '../client';
import type { AuthUser } from '@/stores/auth';

export interface TokenResponse {
  accessToken: string;
  expiresAt: string;
  user: AuthUser;
}

export interface SetupStatus {
  setupRequired: boolean;
}

export interface Profile {
  userId: string;
  locale: string;
  theme: string;
  avatarUrl?: string;
  maxContentRating?: string | null;
  isChild: boolean;
}

export async function getSetupStatus(): Promise<SetupStatus> {
  return apiClient.fetch<SetupStatus>('/setup/status');
}

export async function setupAdmin(username: string, password: string): Promise<TokenResponse> {
  return apiClient.fetch<TokenResponse>('/setup/admin', {
    method: 'POST',
    body: { username, password },
    credentials: 'include',
  });
}

export async function login(username: string, password: string, deviceLabel?: string): Promise<TokenResponse> {
  return apiClient.fetch<TokenResponse>('/auth/login', {
    method: 'POST',
    body: { username, password, deviceLabel },
    credentials: 'include',
  });
}

export async function refreshSession(): Promise<{ accessToken: string; expiresAt: string }> {
  return apiClient.fetch<{ accessToken: string; expiresAt: string }>('/auth/refresh', {
    method: 'POST',
    credentials: 'include',
  });
}

export async function logout(): Promise<void> {
  await apiClient.fetch<void>('/auth/logout', { method: 'POST', credentials: 'include' });
}

export async function getProfile(): Promise<Profile> {
  return apiClient.fetch<Profile>('/profile');
}

export async function updateProfile(body: Partial<Profile>): Promise<Profile> {
  return apiClient.fetch<Profile>('/profile', { method: 'PUT', body });
}

export async function listUsers(): Promise<AuthUser[]> {
  return apiClient.fetch<AuthUser[]>('/users');
}

export async function createUser(body: {
  username: string;
  password: string;
  roles: string[];
}): Promise<AuthUser> {
  return apiClient.fetch<AuthUser>('/users', { method: 'POST', body });
}

export async function updateUser(
  userId: string,
  body: { disabled?: boolean; roles?: string[]; password?: string },
): Promise<AuthUser> {
  return apiClient.fetch<AuthUser>(`/users/${userId}`, { method: 'PUT', body });
}
