import { createServer, type Server } from 'node:http';
import { test, expect } from '@playwright/test';
import { getAdminToken, ensureAdmin } from './helpers';

type WebhookCapture = {
  url: string;
  server: Server;
  received: Array<Record<string, unknown>>;
};

async function startWebhookCapture(): Promise<WebhookCapture> {
  const received: Array<Record<string, unknown>> = [];
  const server = createServer((req, res) => {
    let body = '';
    req.on('data', (chunk: Buffer) => {
      body += chunk.toString();
    });
    req.on('end', () => {
      try {
        received.push(JSON.parse(body) as Record<string, unknown>);
      } catch {
        received.push({ raw: body });
      }
      res.writeHead(200);
      res.end('ok');
    });
  });
  await new Promise<void>((resolve) => server.listen(0, '127.0.0.1', resolve));
  const addr = server.address();
  const port = typeof addr === 'object' && addr ? addr.port : 0;
  return { url: `http://127.0.0.1:${port}/hook`, server, received };
}

test.describe('notification preferences', () => {
  let webhook: WebhookCapture;

  test.beforeEach(async ({ request }) => {
    await ensureAdmin(request);
    webhook = await startWebhookCapture();
  });

  test.afterEach(async () => {
    await new Promise<void>((resolve, reject) => {
      webhook.server.close((err) => (err ? reject(err) : resolve()));
    });
  });

  test('mock webhook: preference off means no notification', async ({ request }) => {
    const token = await getAdminToken(request);

    const policyRes = await request.patch('/api/v1/requests/policies', {
      headers: { Authorization: `Bearer ${token}` },
      data: { userQuotaPerMonth: 0 },
    });
    expect(policyRes.ok()).toBeTruthy();

    const channelRes = await request.post('/api/v1/notifications/channels', {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        channelType: 'webhook',
        name: 'E2E webhook',
        config: { url: webhook.url },
        enabled: true,
      },
    });
    expect(channelRes.ok()).toBeTruthy();
    const channel = (await channelRes.json()) as { id: number };

    const patchOff = await request.patch('/api/v1/notifications/preferences', {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        preferences: [
          {
            eventType: 'request.created',
            channelId: channel.id,
            enabled: false,
            debounceSeconds: 0,
          },
        ],
      },
    });
    expect(patchOff.ok()).toBeTruthy();

    const createOff = await request.post('/api/v1/requests', {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        mediaKind: 'movie',
        provider: 'tmdb',
        externalId: `off-${Date.now()}`,
        title: 'Notification Off Test',
      },
    });
    expect(createOff.ok()).toBeTruthy();

    await expect.poll(() => webhook.received.length, { timeout: 3000 }).toBe(0);

    const patchOn = await request.patch('/api/v1/notifications/preferences', {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        preferences: [
          {
            eventType: 'request.created',
            channelId: channel.id,
            enabled: true,
            debounceSeconds: 0,
          },
        ],
      },
    });
    expect(patchOn.ok()).toBeTruthy();

    const createOn = await request.post('/api/v1/requests', {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        mediaKind: 'movie',
        provider: 'tmdb',
        externalId: `on-${Date.now()}`,
        title: 'Notification On Test',
      },
    });
    expect(createOn.ok()).toBeTruthy();

    await expect.poll(() => webhook.received.length, { timeout: 5000 }).toBeGreaterThan(0);
    expect(webhook.received.some((p) => p.eventType === 'request.created')).toBeTruthy();
  });
});
