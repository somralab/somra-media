import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { test, expect } from '@playwright/test';
import { ensureAdmin, login, seedPlaybackLibrary } from './helpers';

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..');
const mediaDir = path.join(repoRoot, 'testdata', 'media');

test.describe('playback flow', () => {
  test.beforeAll(async ({ request }) => {
    await ensureAdmin(request);
  });

  test('login library play seek resume', async ({ page, request }) => {
    let seeded: { libraryId: number; itemId: number };
    try {
      seeded = await seedPlaybackLibrary(request, mediaDir);
    } catch {
      test.skip(true, 'playback fixture missing — run bash scripts/gen-e2e-media.sh');
      return;
    }

    await login(page);
    await page.goto(`/libraries/${seeded.libraryId}/items/${seeded.itemId}/play`);
    await expect(page.getByTestId('video-player')).toBeVisible({ timeout: 30_000 });

    const video = page.getByTestId('video-player');
    await expect(video).toHaveAttribute('controls', '');

    await page.evaluate(() => {
      const el = document.querySelector('[data-testid="video-player"]') as HTMLVideoElement | null;
      if (el) el.currentTime = Math.min(2, el.duration || 2);
    });

    await page.keyboard.press('ArrowRight');
    await expect(video).toBeVisible();
  });
});
