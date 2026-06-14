import { expect, type APIRequestContext, type Page } from '@playwright/test';

export const E2E_ADMIN = {
  username: 'e2e-admin',
  password: 'E2eAdmin1',
} as const;

export const E2E_USER = {
  username: 'e2e-user',
  password: 'E2eUserPass1',
} as const;

export async function ensureAdmin(request: APIRequestContext): Promise<void> {
  const status = await request.get('/api/v1/setup/status');
  const body = (await status.json()) as {
    setupRequired: boolean;
    completed?: boolean;
    phase?: string;
  };

  let accessToken: string | undefined;

  if (body.phase === 'language' || (body.setupRequired && body.phase !== 'admin')) {
    const lang = await request.post('/api/v1/onboarding/step', {
      data: { phase: 'language', locale: 'en-US' },
    });
    if (!lang.ok()) {
      throw new Error(`onboarding language step failed: ${lang.status()}`);
    }
  }

  if (body.setupRequired) {
    const setup = await request.post('/api/v1/setup/admin', { data: E2E_ADMIN });
    if (!setup.ok()) {
      throw new Error(`setup admin failed: ${setup.status()}`);
    }
    const tok = (await setup.json()) as { accessToken: string };
    accessToken = tok.accessToken;
  }

  const onbStatus = await request.get('/api/v1/onboarding/status');
  const onb = (await onbStatus.json()) as { completed: boolean };
  if (!onb.completed) {
    if (!accessToken) {
      const login = await request.post('/api/v1/auth/login', { data: E2E_ADMIN });
      if (!login.ok()) {
        throw new Error(`login for onboarding complete failed: ${login.status()}`);
      }
      const tok = (await login.json()) as { accessToken: string };
      accessToken = tok.accessToken;
    }
    const done = await request.post('/api/v1/onboarding/complete', {
      headers: { Authorization: `Bearer ${accessToken}` },
    });
    if (!done.ok()) {
      throw new Error(`onboarding complete failed: ${done.status()}`);
    }
  }
}

export async function login(page: Page): Promise<void> {
  await page.goto('/login');
  await page.getByLabel(/username|kullanıcı/i).fill(E2E_ADMIN.username);
  await page.getByLabel(/password|şifre|parola/i).fill(E2E_ADMIN.password);
  await page.getByRole('button', { name: /sign in|giriş/i }).click();
}

export async function getAdminToken(request: APIRequestContext): Promise<string> {
  await ensureAdmin(request);
  const login = await request.post('/api/v1/auth/login', { data: E2E_ADMIN });
  if (!login.ok()) {
    throw new Error(`admin login failed: ${login.status()}`);
  }
  const tok = (await login.json()) as { accessToken: string };
  return tok.accessToken;
}

export async function ensureRegularUser(request: APIRequestContext): Promise<void> {
  const token = await getAdminToken(request);
  const users = await request.get('/api/v1/users', {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!users.ok()) {
    throw new Error(`list users failed: ${users.status()}`);
  }
  const list = (await users.json()) as Array<{ username: string }>;
  if (list.some((u) => u.username === E2E_USER.username)) {
    return;
  }
  const created = await request.post('/api/v1/users', {
    headers: { Authorization: `Bearer ${token}` },
    data: { username: E2E_USER.username, password: E2E_USER.password, roles: ['user'] },
  });
  if (!created.ok()) {
    throw new Error(`create user failed: ${created.status()}`);
  }
}

export async function loginAs(
  page: Page,
  creds: { username: string; password: string },
): Promise<void> {
  await page.goto('/login');
  await page.getByLabel(/username|kullanıcı/i).fill(creds.username);
  await page.getByLabel(/password|şifre|parola/i).fill(creds.password);
  await page.getByRole('button', { name: /sign in|giriş/i }).click();
  await expect(page).not.toHaveURL(/\/login/);
}
