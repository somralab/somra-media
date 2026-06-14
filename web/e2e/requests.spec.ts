import { test, expect } from '@playwright/test';
import { E2E_ADMIN, E2E_USER, ensureAdmin, ensureRegularUser, loginAs } from './helpers';

async function signOut(page: import('@playwright/test').Page): Promise<void> {
  await page.getByRole('button', { name: /sign out|çıkış/i }).click();
  await expect(page).toHaveURL(/\/login/);
}

test.describe('request workflow', () => {
  test.beforeEach(async ({ request }) => {
    await ensureAdmin(request);
    await ensureRegularUser(request);
  });

  test('discover → create → admin approve → status update', async ({ page }) => {
    const uniqueTitle = `E2E Request ${Date.now()}`;

    await loginAs(page, E2E_USER);
    await expect(page).toHaveURL(/\/libraries/);

    await page.goto('/requests/discover');
    await expect(
      page.getByRole('heading', { level: 1, name: /request content|içerik iste/i }),
    ).toBeVisible();

    await page.getByLabel(/search by title|başlıkla ara/i).fill(uniqueTitle);
    await page.getByRole('button', { name: /^search$|ara$/i }).click();

    await expect(page.getByText(uniqueTitle)).toBeVisible();
    await page.getByRole('button', { name: /^request$|iste$/i }).click();

    await expect(
      page.getByRole('heading', { name: /confirm request|isteği onayla/i }),
    ).toBeVisible();
    await page.getByRole('button', { name: /submit request|isteği gönder/i }).click();

    await page.goto('/requests');
    await expect(page.getByText(uniqueTitle)).toBeVisible();
    await expect(
      page
        .getByRole('listitem')
        .filter({ hasText: uniqueTitle })
        .getByText(/^pending$|beklemede$/i),
    ).toBeVisible();

    await signOut(page);
    await loginAs(page, E2E_ADMIN);
    await page.goto('/admin/requests');
    await expect(
      page.getByRole('heading', { level: 1, name: /request management|istek yönetimi/i }),
    ).toBeVisible();
    await expect(page.getByText(uniqueTitle)).toBeVisible();

    const approveBtn = page.getByRole('button', { name: /^approve$|onayla$/i }).first();
    await approveBtn.click();

    await signOut(page);
    await loginAs(page, E2E_USER);
    await page.goto('/requests');
    await expect(page.getByText(uniqueTitle)).toBeVisible();
    await expect(
      page
        .getByRole('listitem')
        .filter({ hasText: uniqueTitle })
        .getByText(/^approved$|onaylandı$/i),
    ).toBeVisible();
  });
});
