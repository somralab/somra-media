import { test, expect } from '@playwright/test';

test.describe('browse flow', () => {
  test('home and libraries require auth', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/login/);
  });

  test('login navigates to home', async ({ page }) => {
    await page.goto('/login');
    await page.getByLabel(/username|kullanıcı/i).fill('admin');
    await page.getByLabel(/password|şifre/i).fill('AdminPass1');
    await page.getByRole('button', { name: /log in|giriş/i }).click();
    await expect(page).toHaveURL('/');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test('libraries page loads after login', async ({ page }) => {
    await page.goto('/login');
    await page.getByLabel(/username|kullanıcı/i).fill('admin');
    await page.getByLabel(/password|şifre/i).fill('AdminPass1');
    await page.getByRole('button', { name: /log in|giriş/i }).click();
    await page.getByRole('link', { name: /libraries|kütüphane/i }).click();
    await expect(page).toHaveURL(/libraries/);
  });
});
