import { test, expect } from '@playwright/test';

async function login(page: import('@playwright/test').Page): Promise<void> {
  await page.goto('/login');
  await page.getByLabel(/username|kullanıcı/i).fill('admin');
  await page.getByLabel(/password|şifre/i).fill('AdminPass1');
  await page.getByRole('button', { name: /log in|giriş/i }).click();
  await expect(page).toHaveURL('/');
}

test.describe('search', () => {
  test('search input visible when logged in', async ({ page }) => {
    await login(page);
    await expect(page.getByRole('searchbox')).toBeVisible();
  });
});
