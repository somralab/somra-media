import { test, expect } from '@playwright/test';

test.describe('playback flow', () => {
  test.skip('login library play seek resume', async ({ page }) => {
    // Requires seeded media library and running backend with test fixtures.
    await page.goto('/login');
    await expect(page).toHaveURL(/login/);
  });
});
