import { test, expect } from '@playwright/test';
import { E2E_ADMIN, ensureAdmin, login } from './helpers';

test.describe('auth flow', () => {
  test('setup admin, login, and reach libraries', async ({ page, request }) => {
    await ensureAdmin(request);

    await page.goto('/login');
    await page.getByLabel(/username|kullanıcı/i).fill(E2E_ADMIN.username);
    await page.getByLabel(/password|şifre|parola/i).fill(E2E_ADMIN.password);
    await page.getByRole('button', { name: /sign in|giriş|create admin|yönetici/i }).click();

    await expect(page).toHaveURL(/\/libraries/);
    await expect(
      page.getByRole('heading', { level: 1, name: /libraries|kütüphane/i }),
    ).toBeVisible();
  });

  test('session persists after page reload', async ({ page, request }) => {
    await ensureAdmin(request);
    await login(page);
    await expect(page).toHaveURL(/\/libraries/);

    await page.reload();
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
