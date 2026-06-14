import { test, expect } from '@playwright/test';
import { ensureAdmin, login } from './helpers';

test.describe('browse flow', () => {
  test.beforeAll(async ({ request }) => {
    await ensureAdmin(request);
  });

  test('home and libraries require auth', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/login/);
  });

  test('login navigates to home', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/login/);
    await page.getByLabel(/username|kullanıcı/i).fill('e2e-admin');
    await page.getByLabel(/password|şifre|parola/i).fill('E2eAdmin1');
    await page.getByRole('button', { name: /sign in|giriş/i }).click();
    await expect(page).toHaveURL('/');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test('libraries page loads after login', async ({ page }) => {
    await login(page);
    await page.getByRole('link', { name: /libraries|kütüphane/i }).click();
    await expect(page).toHaveURL(/libraries/);
  });
});
