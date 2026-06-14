import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import enCommon from './locales/en-US/common.json';
import enStatus from './locales/en-US/status.json';
import enLibrary from './locales/en-US/library.json';
import enAuth from './locales/en-US/auth.json';
import enStreaming from './locales/en-US/streaming.json';
import enPlayer from './locales/en-US/player.json';
import enDiscover from './locales/en-US/discover.json';
import enBrowse from './locales/en-US/browse.json';
import enSearch from './locales/en-US/search.json';
import enDetail from './locales/en-US/detail.json';
import trCommon from './locales/tr-TR/common.json';
import trStatus from './locales/tr-TR/status.json';
import trLibrary from './locales/tr-TR/library.json';
import trAuth from './locales/tr-TR/auth.json';
import trStreaming from './locales/tr-TR/streaming.json';
import trPlayer from './locales/tr-TR/player.json';
import trDiscover from './locales/tr-TR/discover.json';
import trBrowse from './locales/tr-TR/browse.json';
import trSearch from './locales/tr-TR/search.json';
import trDetail from './locales/tr-TR/detail.json';

export const SUPPORTED_LOCALES = ['en-US', 'tr-TR'] as const;
export type SupportedLocale = (typeof SUPPORTED_LOCALES)[number];

export const FALLBACK_LOCALE: SupportedLocale = 'en-US';
export const DEFAULT_NAMESPACE = 'common';
export const NAMESPACES = [
  'common',
  'status',
  'library',
  'auth',
  'streaming',
  'player',
  'discover',
  'browse',
  'search',
  'detail',
] as const;

export const resources = {
  'en-US': {
    common: enCommon,
    status: enStatus,
    library: enLibrary,
    auth: enAuth,
    streaming: enStreaming,
    player: enPlayer,
    discover: enDiscover,
    browse: enBrowse,
    search: enSearch,
    detail: enDetail,
  },
  'tr-TR': {
    common: trCommon,
    status: trStatus,
    library: trLibrary,
    auth: trAuth,
    streaming: trStreaming,
    player: trPlayer,
    discover: trDiscover,
    browse: trBrowse,
    search: trSearch,
    detail: trDetail,
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
