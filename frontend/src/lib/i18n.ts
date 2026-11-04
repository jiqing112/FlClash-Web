// i18n over the app's own arb files (en + zh_CN), matching FlClash strings.
import en from '../i18n/en.json';
import zhCN from '../i18n/zh_CN.json';
import { derived } from 'svelte/store';
import { config } from './stores';

const bundles: Record<string, Record<string, string>> = {
  en: en as Record<string, string>,
  zh_CN: zhCN as Record<string, string>,
};

// Only the bundles the app ships; a locale without a bundle would silently
// render English. `''` follows the browser language.
export const locales = [
  { value: '', label: 'Default' },
  { value: 'en', label: 'English' },
  { value: 'zh_CN', label: '简体中文' },
];

function detect(): string {
  if (typeof navigator === 'undefined') return 'en';
  const lang = navigator.language || 'en';
  if (lang.toLowerCase().startsWith('zh')) return 'zh_CN';
  return 'en';
}

// The active locale: explicit setting wins, else the browser language.
export const locale = derived(config, ($c) => {
  const setting = $c?.appSettingProps?.locale;
  if (setting) return setting;
  return detect();
});

type Replacements = Record<string, string | number>;

// ICU plural: reduce `{name, plural, =1{...} other{...}}` to the branch the
// count picks, then substitute placeholders.
function resolvePlural(text: string, replacements?: Replacements): string {
  return text.replace(/\{(\w+),\s*plural,\s*([\s\S]*?)\}\s*(?=[,}.)]|$)/g, (_all, name: string, rest: string) => {
    const count = Number(replacements?.[name] ?? 0);
    // split top-level branches like =1{...} other{...}
    const branches: { key: string; body: string }[] = [];
    let depth = 0;
    let current = '';
    let currentKey = '';
    let readingKey = true;
    for (const char of rest) {
      if (char === '{') {
        depth++;
        if (depth === 1) {
          readingKey = false;
          continue;
        }
      } else if (char === '}') {
        depth--;
        if (depth === 0) {
          branches.push({ key: currentKey.trim(), body: current });
          current = '';
          currentKey = '';
          readingKey = true;
          continue;
        }
      }
      if (depth === 0) {
        if (readingKey) currentKey += char;
      } else {
        current += char;
      }
    }
    let branch = branches.find((b) => b.key === `=${count}`)?.body
      ?? branches.find((b) => b.key === 'other')?.body
      ?? '';
    return branch.replaceAll(`{${name}}`, String(count));
  });
}

function translate(bundle: Record<string, string>, key: string, replacements?: Replacements): string {
  let text = bundle[key] ?? bundles.en[key] ?? key;
  text = resolvePlural(text, replacements);
  if (replacements) {
    text = text.replace(/\{(\w+)\}/g, (all, name: string) =>
      name in replacements ? String(replacements[name]) : all,
    );
  }
  return text;
}

// t('key', {count: 1}) — placeholders use the arb `{name}` form.
export function t(key: string, replacements?: Replacements): string {
  const current = bundles[localeKey] ?? bundles.en;
  return translate(current, key, replacements);
}

let localeKey = 'en';
locale.subscribe((value) => {
  localeKey = bundles[value] ? value : 'en';
});
