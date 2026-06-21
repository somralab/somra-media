import { test, expect } from '@playwright/test';
import { ensureAdmin, login } from './helpers';

test.describe('automation series monitors', () => {
  test.beforeEach(async ({ request }) => {
    await ensureAdmin(request);
  });

  test('admin can open monitors page and create form', async ({ page }) => {
    await login(page);
    await expect(page).not.toHaveURL(/\/login/);

    const monitorsReady = page.waitForResponse(
      (res) => res.url().includes('/api/v1/automation/monitors') && res.ok(),
    );
    await page.goto('/settings/automation/monitors');
    await monitorsReady;

    await expect(
      page.getByRole('heading', { level: 1, name: /series monitors|dizi monitörleri/i }),
    ).toBeVisible();

    await page.getByRole('link', { name: /add monitor|monitör ekle/i }).click();
    await expect(page.getByText(/series title|dizi başlığı/i)).toBeVisible();
  });
});
