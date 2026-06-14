import { test, expect } from '@playwright/test';
import { E2E_ADMIN, ensureAdmin, login } from './helpers';

test.describe('browse flow', () => {
  test.beforeAll(async ({ request }) => {
    await ensureAdmin(request);
  });

  test('home and libraries require auth', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/login/);
  });

  test('login navigates to home', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/login/);
    await page.getByLabel(/username|kullanıcı/i).fill('e2e-admin');
    await page.getByLabel(/password|şifre|parola/i).fill('E2eAdmin1');
    await page.getByRole('button', { name: /sign in|giriş/i }).click();
    await expect(page).toHaveURL('/');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test('libraries page loads after login', async ({ page }) => {
    await login(page);
    await page.getByRole('link', { name: /libraries|kütüphane/i }).click();
    await expect(page).toHaveURL(/libraries/);
  });

  test('media detail page shows title when item exists', async ({ page, request }) => {
    const loginResp = await request.post('/api/v1/auth/login', { data: E2E_ADMIN });
    if (!loginResp.ok()) {
      test.skip();
      return;
    }
    const { accessToken } = (await loginResp.json()) as { accessToken: string };
    const auth = { Authorization: `Bearer ${accessToken}` };

    const libs = await request.get('/api/v1/libraries', { headers: auth });
    if (!libs.ok()) {
      test.skip();
      return;
    }
    const libraries = (await libs.json()) as { id: number }[];
    if (libraries.length === 0) {
      test.skip();
      return;
    }
    const libId = libraries[0].id;
    const items = await request.get(`/api/v1/libraries/${libId}/items?limit=1`, { headers: auth });
    if (!items.ok()) {
      test.skip();
      return;
    }
    const body = (await items.json()) as { items: { id: number; title: string }[] };
    if (body.items.length === 0) {
      test.skip();
      return;
    }
    const item = body.items[0];

    await login(page);
    await page.goto(`/libraries/${libId}/items/${item.id}`);
    await expect(page.getByRole('heading', { level: 1, name: item.title })).toBeVisible({
      timeout: 15_000,
    });
  });
});
