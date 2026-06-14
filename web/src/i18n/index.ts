import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import enCommon from './locales/en-US/common.json';
import enStatus from './locales/en-US/status.json';
import trCommon from './locales/tr-TR/common.json';
import trStatus from './locales/tr-TR/status.json';

export const SUPPORTED_LOCALES = ['en-US', 'tr-TR'] as const;
export type SupportedLocale = (typeof SUPPORTED_LOCALES)[number];

export const FALLBACK_LOCALE: SupportedLocale = 'en-US';
export const DEFAULT_NAMESPACE = 'common';
export const NAMESPACES = ['common', 'status'] as const;

export const resources = {
  'en-US': {
    common: enCommon,
    status: enStatus,
  },
  'tr-TR': {
    common: trCommon,
    status: trStatus,
  },
} as const;

void i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    supportedLngs: [...SUPPORTED_LOCALES],
    fallbackLng: FALLBACK_LOCALE,
    defaultNS: DEFAULT_NAMESPACE,
    fallbackNS: DEFAULT_NAMESPACE,
    ns: [...NAMESPACES],
    nonExplicitSupportedLngs: false,
    load: 'currentOnly',
    interpolation: {
      escapeValue: false,
    },
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'],
      lookupLocalStorage: 'somra.locale',
    },
    returnNull: false,
  });

i18n.on('languageChanged', (lng) => {
  const html = typeof document !== 'undefined' ? document.documentElement : null;
  if (html) {
    html.setAttribute('lang', lng);
  }
  if (typeof document !== 'undefined') {
    const appName = i18n.t('app.name');
    const tagline = i18n.t('app.tagline');
    document.title = `${appName} — ${tagline}`;
  }
});

export default i18n;
