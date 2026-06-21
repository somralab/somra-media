import { test, expect } from '@playwright/test';
import { ensureAdmin, getAdminToken, login } from './helpers';

test.describe('automation plugins admin', () => {
  test.beforeEach(async ({ request }) => {
    await ensureAdmin(request);
  });

  test('admin can add stub indexer and download client', async ({ page, request }) => {
    const token = await getAdminToken(request);

    const indexer = await request.post('/api/v1/plugins/instances', {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        pluginType: 'indexer',
        implementation: 'stub',
        name: `e2e-indexer-${Date.now()}`,
        enabled: true,
      },
    });
    expect(indexer.ok()).toBeTruthy();

    const client = await request.post('/api/v1/plugins/instances', {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        pluginType: 'download_client',
        implementation: 'stub',
        name: `e2e-dl-${Date.now()}`,
        enabled: true,
      },
    });
    expect(client.ok()).toBeTruthy();

    await login(page);
    await page.goto('/settings/automation/indexers');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
    await page.goto('/settings/automation/download-clients');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });
});
