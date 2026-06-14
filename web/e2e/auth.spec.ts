import { test, expect } from '@playwright/test';

test.describe('auth flow', () => {
  test('setup admin, login, and reach libraries', async ({ page, request }) => {
    const status = await request.get('/api/v1/setup/status');
    const body = (await status.json()) as { setupRequired: boolean };

    if (body.setupRequired) {
      const setup = await request.post('/api/v1/setup/admin', {
        data: { username: 'e2e-admin', password: 'E2eAdmin1' },
      });
      expect(setup.ok()).toBeTruthy();
    }

    await page.goto('/login');
    await page.getByLabel(/username|kullanıcı/i).fill('e2e-admin');
    await page.getByLabel(/password|parola/i).fill('E2eAdmin1');
    await page.getByRole('button', { name: /sign in|giriş|create admin|yönetici/i }).click();

    await expect(page).toHaveURL(/\/libraries/);
    await expect(
      page.getByRole('heading', { level: 1, name: /libraries|kütüphane/i }),
    ).toBeVisible();
  });

  test('unauthenticated libraries API returns 401', async ({ request }) => {
    const resp = await request.get('/api/v1/libraries');
    expect(resp.status()).toBe(401);
  });
});
