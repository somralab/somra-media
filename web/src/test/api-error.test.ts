import { describe, expect, it } from 'vitest';
import { ApiError } from '@/api/ApiError';
import i18n from '@/i18n';

describe('ApiError', () => {
  it('exposes typed envelope fields', () => {
    const err = new ApiError({
      status: 503,
      code: 'service_unavailable',
      messageKey: 'system.health.unavailable',
      message: 'Service is temporarily unavailable.',
      requestId: 'req-1',
      details: { database: 'down' },
    });

    expect(err.name).toBe('ApiError');
    expect(err.status).toBe(503);
    expect(err.code).toBe('service_unavailable');
    expect(err.messageKey).toBe('system.health.unavailable');
    expect(err.message).toBe('Service is temporarily unavailable.');
    expect(err.requestId).toBe('req-1');
    expect(err.details).toEqual({ database: 'down' });
  });

  it('builds from a backend ErrorResponse envelope', () => {
    const err = ApiError.fromEnvelope(404, {
      code: 'not_found',
      messageKey: 'common.errors.not_found',
      message: 'Not found',
      requestId: 'abc',
    });
    expect(err.status).toBe(404);
    expect(err.requestId).toBe('abc');
  });

  describe('t()', () => {
    it('resolves the messageKey through the translator when available', async () => {
      await i18n.changeLanguage('en-US');
      const err = new ApiError({
        status: 0,
        code: 'network_error',
        messageKey: 'api.errors.network',
        message: 'Server-supplied fallback.',
      });
      expect(err.t(i18n.t.bind(i18n))).toBe(
        'Network request failed. Check your connection and try again.',
      );
    });

    it('resolves to the localized tr-TR text when language is tr-TR', async () => {
      await i18n.changeLanguage('tr-TR');
      const err = new ApiError({
        status: 0,
        code: 'network_error',
        messageKey: 'api.errors.network',
        message: 'Server-supplied fallback.',
      });
      expect(err.t(i18n.t.bind(i18n))).toBe(
        'Ağ isteği başarısız oldu. Bağlantınızı kontrol edin ve tekrar deneyin.',
      );
      await i18n.changeLanguage('en-US');
    });

    it('falls back to the server message when the messageKey is missing', () => {
      const err = new ApiError({
        status: 500,
        code: 'server_error',
        messageKey: 'definitely.not.defined',
        message: 'Server-supplied fallback.',
      });
      expect(err.t(i18n.t.bind(i18n))).toBe('Server-supplied fallback.');
    });
  });
});
