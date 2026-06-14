import { test, expect } from '@playwright/test';

async function login(page: import('@playwright/test').Page): Promise<void> {
  await page.goto('/login');
  await page.getByLabel(/username|kullanıcı/i).fill('admin');
  await page.getByLabel(/password|şifre/i).fill('AdminPass1');
  await page.getByRole('button', { name: /log in|giriş/i }).click();
}

test.describe('themes', () => {
  test('theme switcher changes data-theme', async ({ page }) => {
    await page.goto('/settings');
    const select = page.getByLabel(/theme|tema/i);
    await select.selectOption('aurora');
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'aurora');
    await select.selectOption('noir');
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'noir');
    await select.selectOption('minimal');
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'minimal');
    await select.selectOption('cinematic');
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'cinematic');
  });

  test('home respects theme after login', async ({ page }) => {
    await login(page);
    await page.goto('/settings');
    await page.getByLabel(/theme|tema/i).selectOption('noir');
    await page.goto('/');
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'noir');
  });
});
