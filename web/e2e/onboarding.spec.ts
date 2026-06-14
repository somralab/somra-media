import { test, expect } from '@playwright/test';

test.describe('onboarding wizard', () => {
  test('fresh install redirects to setup wizard', async ({ page }) => {
    const start = Date.now();
    await page.goto('/');
    await expect(page).toHaveURL(/\/(setup\/wizard|login)/);
    const elapsed = Date.now() - start;
    expect(elapsed).toBeLessThan(60_000);
  });

  test('language step is visible on wizard', async ({ page }) => {
    await page.goto('/setup/wizard');
    await expect(page.getByText(/welcome to somra|somra'ya hoş geldiniz/i)).toBeVisible({
      timeout: 15_000,
    });
  });
});

test.describe('onboarding flow with API', () => {
  test('setup status includes phase', async ({ request }) => {
    const res = await request.get('/api/v1/setup/status');
    expect(res.ok()).toBeTruthy();
    const body = (await res.json()) as { setupRequired: boolean; phase?: string };
    expect(body).toHaveProperty('setupRequired');
    if (body.phase) {
      expect(['language', 'admin', 'library', 'defaults', 'scan', 'complete']).toContain(
        body.phase,
      );
    }
  });

  test('system detect returns profile', async ({ request }) => {
    const res = await request.get('/api/v1/system/detect');
    expect(res.ok()).toBeTruthy();
    const body = (await res.json()) as { cpuCores: number; gpuPresent: boolean };
    expect(body.cpuCores).toBeGreaterThan(0);
    expect(typeof body.gpuPresent).toBe('boolean');
  });
});
