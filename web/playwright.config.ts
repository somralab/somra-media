import { defineConfig, devices } from '@playwright/test';

const port = Number(process.env.SOMRA_E2E_PORT ?? 8080);
const baseURL = process.env.SOMRA_E2E_BASE_URL ?? `http://127.0.0.1:${port}`;

const repoRoot = new URL('..', import.meta.url).pathname;

const sharedEnv = {
  SOMRA_HTTP_ADDR: `:${port}`,
  SOMRA_LOG_FORMAT: 'text',
  SOMRA_LOG_LEVEL: 'info',
  SOMRA_DATA_DIR: process.env.SOMRA_E2E_DATA_DIR ?? '/tmp/somra-e2e-data',
  SOMRA_WEB_DIR: process.env.SOMRA_E2E_WEB_DIR ?? `${repoRoot}/web/dist`,
  SOMRA_USE_TEST_METADATA: '1',
  PATH: process.env.PATH ?? '',
} satisfies Record<string, string>;

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  expect: { timeout: 5_000 },
  fullyParallel: false,
  retries: 0,
  workers: 1,
  reporter: [['list']],
  use: {
    baseURL,
    trace: 'off',
    video: 'off',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: {
    command: process.env.SOMRA_E2E_BACKEND_CMD ?? `go run ./cmd/somra`,
    cwd: repoRoot,
    url: `${baseURL}/api/v1/health`,
    reuseExistingServer: true,
    timeout: 120_000,
    env: sharedEnv,
    stdout: 'pipe',
    stderr: 'pipe',
  },
});
