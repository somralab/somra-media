#!/usr/bin/env node
/*
 * i18n parity check.
 * Asserts that every translation locale has the same key set as the source locale (en-US),
 * across all namespaces under src/i18n/locales.
 *
 * Exits non-zero if any key is missing or extra.
 */

import { readdir, readFile, stat } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const localesRoot = path.resolve(__dirname, '..', 'src', 'i18n', 'locales');
const SOURCE_LOCALE = 'en-US';

function flatten(obj, prefix = '') {
  const out = [];
  for (const [key, value] of Object.entries(obj)) {
    const next = prefix ? `${prefix}.${key}` : key;
    if (value !== null && typeof value === 'object' && !Array.isArray(value)) {
      out.push(...flatten(value, next));
    } else {
      out.push(next);
    }
  }
  return out;
}

async function listLocales(dir) {
  const entries = await readdir(dir);
  const locales = [];
  for (const entry of entries) {
    const full = path.join(dir, entry);
    const st = await stat(full);
    if (st.isDirectory()) locales.push(entry);
  }
  return locales.sort();
}

async function loadNamespaces(localeDir) {
  const files = (await readdir(localeDir)).filter((f) => f.endsWith('.json')).sort();
  const ns = {};
  for (const file of files) {
    const name = path.basename(file, '.json');
    const raw = await readFile(path.join(localeDir, file), 'utf8');
    ns[name] = JSON.parse(raw);
  }
  return ns;
}

function diffKeys(reference, candidate) {
  const refSet = new Set(reference);
  const candSet = new Set(candidate);
  const missing = reference.filter((k) => !candSet.has(k));
  const extra = candidate.filter((k) => !refSet.has(k));
  return { missing, extra };
}

async function main() {
  const locales = await listLocales(localesRoot);
  if (!locales.includes(SOURCE_LOCALE)) {
    console.error(`[i18n-parity] source locale ${SOURCE_LOCALE} not found under ${localesRoot}`);
    process.exit(1);
  }

  const source = await loadNamespaces(path.join(localesRoot, SOURCE_LOCALE));
  const sourceKeys = {};
  for (const [ns, content] of Object.entries(source)) {
    sourceKeys[ns] = flatten(content).sort();
  }

  let failed = false;
  for (const locale of locales) {
    if (locale === SOURCE_LOCALE) continue;
    const target = await loadNamespaces(path.join(localesRoot, locale));

    const sourceNs = new Set(Object.keys(sourceKeys));
    const targetNs = new Set(Object.keys(target));
    const missingNs = [...sourceNs].filter((n) => !targetNs.has(n));
    const extraNs = [...targetNs].filter((n) => !sourceNs.has(n));
    if (missingNs.length || extraNs.length) {
      failed = true;
      if (missingNs.length) {
        console.error(`[i18n-parity] ${locale}: missing namespace(s): ${missingNs.join(', ')}`);
      }
      if (extraNs.length) {
        console.error(`[i18n-parity] ${locale}: unexpected namespace(s): ${extraNs.join(', ')}`);
      }
    }

    for (const [ns, refKeys] of Object.entries(sourceKeys)) {
      if (!target[ns]) continue;
      const tgtKeys = flatten(target[ns]).sort();
      const { missing, extra } = diffKeys(refKeys, tgtKeys);
      if (missing.length || extra.length) {
        failed = true;
        if (missing.length) {
          console.error(`[i18n-parity] ${locale}/${ns}.json missing keys:`);
          for (const k of missing) console.error(`  - ${k}`);
        }
        if (extra.length) {
          console.error(`[i18n-parity] ${locale}/${ns}.json extra keys:`);
          for (const k of extra) console.error(`  + ${k}`);
        }
      }
    }
  }

  if (failed) {
    process.exit(1);
  }
  console.log(`[i18n-parity] OK — ${locales.length} locale(s) in parity with ${SOURCE_LOCALE}.`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
