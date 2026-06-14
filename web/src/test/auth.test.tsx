import { describe, expect, it } from 'vitest';
import { useAuthStore } from '@/stores/auth';

describe('auth store', () => {
  it('tracks admin role', () => {
    useAuthStore.setState({
      accessToken: 'tok',
      expiresAt: new Date().toISOString(),
      user: { id: '1', username: 'admin', roles: ['admin'], disabled: false },
    });
    expect(useAuthStore.getState().isAdmin()).toBe(true);
    useAuthStore.getState().clearSession();
    expect(useAuthStore.getState().accessToken).toBeNull();
  });
});
