import { create } from 'zustand';

export interface AuthUser {
  id: string;
  username: string;
  roles: string[];
  disabled: boolean;
}

interface AuthState {
  accessToken: string | null;
  expiresAt: string | null;
  user: AuthUser | null;
  setSession: (accessToken: string, expiresAt: string, user: AuthUser) => void;
  clearSession: () => void;
  isAdmin: () => boolean;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  accessToken: null,
  expiresAt: null,
  user: null,
  setSession: (accessToken, expiresAt, user) => set({ accessToken, expiresAt, user }),
  clearSession: () => set({ accessToken: null, expiresAt: null, user: null }),
  isAdmin: () => get().user?.roles.includes('admin') ?? false,
}));

export function getAccessToken(): string | null {
  return useAuthStore.getState().accessToken;
}

export function setAuthSession(accessToken: string, expiresAt: string, user: AuthUser): void {
  useAuthStore.getState().setSession(accessToken, expiresAt, user);
}

export function clearAuthSession(): void {
  useAuthStore.getState().clearSession();
}
