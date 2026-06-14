import { test, expect } from '@playwright/test';
import { ensureAdmin, login } from './helpers';

test.describe('subtitles API', () => {
  test.beforeEach(async ({ request }) => {
    await ensureAdmin(request);
  });

  test('subtitle list requires auth', async ({ request }) => {
    const res = await request.get('/api/v1/media-items/1/subtitles');
    expect(res.status()).toBe(401);
  });

  test('subtitle search requires auth', async ({ request }) => {
    const res = await request.post('/api/v1/subtitles/search', {
      data: { mediaItemId: 1, language: 'en' },
    });
    expect(res.status()).toBe(401);
  });
});

test.describe('subtitles UI', () => {
  test.beforeEach(async ({ request }) => {
    await ensureAdmin(request);
  });

  test('media detail shows subtitle section when item exists', async ({ page, request }) => {
    await login(page);
    const libs = await request.get('/api/v1/libraries', {
      headers: { Authorization: `Bearer ${await getToken(request)}` },
    });
    if (!libs.ok()) {
      test.skip();
      return;
    }
    const libraries = (await libs.json()) as { id: number }[];
    if (libraries.length === 0) {
      test.skip();
      return;
    }
    const items = await request.get(`/api/v1/libraries/${libraries[0].id}/items?limit=1`, {
      headers: { Authorization: `Bearer ${await getToken(request)}` },
    });
    if (!items.ok()) {
      test.skip();
      return;
    }
    const list = (await items.json()) as { items?: { id: number }[] };
    const item = list.items?.[0];
    if (!item) {
      test.skip();
      return;
    }
    await page.goto(`/libraries/${libraries[0].id}/items/${item.id}`);
    await expect(page.getByTestId('subtitle-section')).toBeVisible({ timeout: 10_000 });
  });
});

async function getToken(request: import('@playwright/test').APIRequestContext): Promise<string> {
  const loginRes = await request.post('/api/v1/auth/login', {
    data: { username: 'e2e-admin', password: 'E2eAdmin1' },
  });
  const body = (await loginRes.json()) as { accessToken: string };
  return body.accessToken;
}
