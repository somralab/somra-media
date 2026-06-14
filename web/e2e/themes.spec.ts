import { test, expect } from '@playwright/test';
import { ensureAdmin, login } from './helpers';

test.describe('themes', () => {
  test.beforeAll(async ({ request }) => {
    await ensureAdmin(request);
  });

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
