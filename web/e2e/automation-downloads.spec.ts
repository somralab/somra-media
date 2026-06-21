import { test, expect } from '@playwright/test';
import { ensureAdmin, login } from './helpers';

test.describe('automation downloads', () => {
  test.beforeEach(async ({ request }) => {
    await ensureAdmin(request);
  });

  test('admin downloads page loads', async ({ page }) => {
    await login(page);
    await expect(page).not.toHaveURL(/\/login/);

    const downloadsReady = page.waitForResponse(
      (res) => res.url().includes('/api/v1/automation/downloads') && res.ok(),
    );
    await page.goto('/automation/downloads');
    await downloadsReady;

    await expect(
      page.getByRole('heading', { level: 1, name: /automation downloads|otomasyon indirmeleri/i }),
    ).toBeVisible();
  });
});
