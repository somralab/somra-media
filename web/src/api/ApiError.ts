import type { TFunction } from 'i18next';
import type { components } from './generated/openapi';

export type ErrorEnvelope = components['schemas']['ErrorResponse'];

export interface ApiErrorInit {
  status: number;
  code: string;
  messageKey: string;
  message: string;
  requestId?: string;
  details?: Record<string, unknown>;
  cause?: unknown;
}

/**
 * Strongly-typed error raised by the Somra HTTP client. Mirrors the
 * `ErrorResponse` envelope from `api/openapi.yaml`.
 *
 * Use `t()` to render the i18n `messageKey` through a translator, falling
 * back to the server-localized `message` when the key cannot be resolved.
 */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly messageKey: string;
  readonly requestId?: string;
  readonly details?: Record<string, unknown>;

  constructor(init: ApiErrorInit) {
    super(init.message, init.cause === undefined ? undefined : { cause: init.cause });
    this.name = 'ApiError';
    this.status = init.status;
    this.code = init.code;
    this.messageKey = init.messageKey;
    if (init.requestId !== undefined) {
      this.requestId = init.requestId;
    }
    if (init.details !== undefined) {
      this.details = init.details;
    }
  }

  /**
   * Resolve the i18n `messageKey` through the supplied translator. When the
   * translator does not have a localization for the key, falls back to the
   * server-provided localized `message` so the user always sees something.
   */
  t(translator: TFunction): string {
    const translated = translator(this.messageKey, { defaultValue: '' });
    if (typeof translated === 'string' && translated.length > 0 && translated !== this.messageKey) {
      return translated;
    }
    return this.message;
  }

  static fromEnvelope(status: number, envelope: ErrorEnvelope, cause?: unknown): ApiError {
    return new ApiError({
      status,
      code: envelope.code,
      messageKey: envelope.messageKey,
      message: envelope.message,
      ...(envelope.requestId !== undefined ? { requestId: envelope.requestId } : {}),
      ...(envelope.details !== undefined ? { details: envelope.details } : {}),
      ...(cause !== undefined ? { cause } : {}),
    });
  }

  static network(messageKey: string, message: string, cause?: unknown): ApiError {
    return new ApiError({
      status: 0,
      code: 'network_error',
      messageKey,
      message,
      ...(cause !== undefined ? { cause } : {}),
    });
  }
}

export function isApiError(value: unknown): value is ApiError {
  return value instanceof ApiError;
}
