import { test, expect } from '@playwright/test';

// End-to-end smoke for the M1 status dashboard. Runs against the Go
// binary (`go run ./cmd/somra`) which serves the built SPA from
// `SOMRA_WEB_DIR=web/dist` — Playwright handles the webServer in
// `playwright.config.ts`. The spec is intentionally minimal: it asserts
// the three things the user can verify by eye on a working install.
test.describe('M1 status dashboard', () => {
  test('shows non-empty version, health label and at least one SSE event', async ({ page }) => {
    await page.goto('/status');

    const versionValue = page.getByTestId('status-version-value');
    await expect(versionValue).toBeVisible({ timeout: 10_000 });
    await expect(versionValue).not.toHaveText('');

    const healthValue = page.getByTestId('status-health-value');
    await expect(healthValue).toBeVisible();
    await expect(healthValue).toHaveAttribute('data-status', /ok|degraded|unavailable/);

    const connection = page.getByTestId('status-events-connection');
    await expect(connection).toBeVisible();

    const lastEvent = page.getByTestId('status-events-last');
    await expect(lastEvent).toBeVisible({ timeout: 10_000 });
  });
});
