import { test, expect } from '@playwright/test';
import { ensureAdmin, login } from './helpers';

test.describe('settings regression', () => {
  test.beforeEach(async ({ request }) => {
    await ensureAdmin(request);
  });

  test('settings page loads categories', async ({ page }) => {
    await login(page);
    await page.goto('/settings');
    await expect(page.getByRole('heading', { name: /settings|ayarlar/i })).toBeVisible();
    await expect(page.getByText(/language|dil/i).first()).toBeVisible();
  });

  test('advanced settings toggle', async ({ page }) => {
    await login(page);
    await page.goto('/settings');
    await expect(page.getByText(/general|genel/i).first()).toBeVisible();
    await page.getByRole('button', { name: /advanced|gelişmiş/i }).click();
    await expect(page.getByRole('heading', { name: /library|kütüphane/i })).toBeVisible();
  });

  test('HW acceleration settings in advanced mode', async ({ page }) => {
    await login(page);
    await page.goto('/settings');
    await expect(page.getByText(/general|genel/i).first()).toBeVisible();
    await page.getByRole('button', { name: /advanced|gelişmiş/i }).click();
    await expect(
      page.getByText(/hardware acceleration|donanım hızlandırma/i).first(),
    ).toBeVisible();
  });
});

test.describe('settings API', () => {
  test('settings require auth', async ({ request }) => {
    const res = await request.get('/api/v1/settings');
    expect(res.status()).toBe(401);
  });
});
