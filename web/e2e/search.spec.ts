import { test, expect } from '@playwright/test';
import { ensureAdmin, login } from './helpers';

test.describe('search', () => {
  test.beforeAll(async ({ request }) => {
    await ensureAdmin(request);
  });

  test('search input visible when logged in', async ({ page }) => {
    await login(page);
    await expect(page).toHaveURL(/\/(libraries)?/);
    await expect(page.getByRole('searchbox')).toBeVisible();
  });
});
