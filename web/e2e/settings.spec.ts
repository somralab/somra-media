import { test, expect, type Page } from '@playwright/test';
import { ensureAdmin, login } from './helpers';

async function openSettingsWithGeneral(page: Page): Promise<void> {
  await login(page);
  await expect(page).not.toHaveURL(/\/login/);
  const settingsReady = page.waitForResponse(
    (res) => res.url().includes('/api/v1/settings') && res.ok(),
  );
  await page.goto('/settings');
  await settingsReady;
  await expect(page.getByRole('heading', { name: /general|genel/i })).toBeVisible();
}

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
    await openSettingsWithGeneral(page);
    await page.getByRole('button', { name: /advanced|gelişmiş/i }).click();
    await expect(page.getByRole('heading', { name: /library|kütüphane/i })).toBeVisible();
  });

  test('HW acceleration settings in advanced mode', async ({ page }) => {
    await openSettingsWithGeneral(page);
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
