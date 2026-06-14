import type { APIRequestContext, Page } from '@playwright/test';

export const E2E_ADMIN = {
  username: 'e2e-admin',
  password: 'E2eAdmin1',
} as const;

export async function ensureAdmin(request: APIRequestContext): Promise<void> {
  const status = await request.get('/api/v1/setup/status');
  const body = (await status.json()) as { setupRequired: boolean };

  if (body.setupRequired) {
    const setup = await request.post('/api/v1/setup/admin', { data: E2E_ADMIN });
    if (!setup.ok()) {
      throw new Error(`setup admin failed: ${setup.status()}`);
    }
  }
}

export async function login(page: Page): Promise<void> {
  await page.goto('/login');
  await page.getByLabel(/username|kullanıcı/i).fill(E2E_ADMIN.username);
  await page.getByLabel(/password|şifre|parola/i).fill(E2E_ADMIN.password);
  await page.getByRole('button', { name: /sign in|giriş/i }).click();
}
