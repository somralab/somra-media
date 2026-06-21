import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  createPluginInstance,
  deletePluginInstance,
  getPluginInstance,
  listPluginCatalog,
  listPluginInstances,
  patchPluginInstance,
  testPluginInstance,
} from '@/api/endpoints/plugins';
import i18n from '@/i18n';

const originalFetch = globalThis.fetch;

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('plugins endpoints', () => {
  beforeEach(async () => {
    await i18n.changeLanguage('en-US');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it('lists catalog and instances', async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(200, { catalog: [{ pluginType: 'indexer' }] }))
      .mockResolvedValueOnce(jsonResponse(200, { instances: [{ id: 1, name: 'idx' }] }))
      .mockResolvedValueOnce(jsonResponse(200, { id: 1, name: 'idx' })) as unknown as typeof fetch;

    const catalog = await listPluginCatalog();
    expect(catalog.catalog).toHaveLength(1);

    const instances = await listPluginInstances();
    expect(instances.instances[0]?.name).toBe('idx');

    await getPluginInstance(1);
  });

  it('mutates plugin instances', async () => {
    globalThis.fetch = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse(201, {
          id: 2,
          pluginType: 'indexer',
          implementation: 'stub',
          name: 'new',
          enabled: true,
        }),
      )
      .mockResolvedValueOnce(jsonResponse(200, { id: 2, name: 'renamed' }))
      .mockResolvedValueOnce(jsonResponse(200, { ok: true }))
      .mockResolvedValueOnce(new Response(null, { status: 204 })) as unknown as typeof fetch;

    const created = await createPluginInstance({
      pluginType: 'indexer',
      implementation: 'stub',
      name: 'new',
      enabled: true,
    });
    expect(created.id).toBe(2);

    const patched = await patchPluginInstance(2, { name: 'renamed' });
    expect(patched.name).toBe('renamed');

    const tested = await testPluginInstance(2);
    expect(tested).toEqual({ ok: true });

    await deletePluginInstance(2);
  });
});
