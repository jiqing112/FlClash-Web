// Theme: computes the color scheme from themeProps and writes CSS variables.
import { derived } from 'svelte/store';
import { config } from './stores';
import {
  seededColorScheme,
  toPureBlack,
  successColor,
  successContainerColor,
  onSuccessContainerColor,
  warningColor,
  warningContainerColor,
  onWarningContainerColor,
} from './color';
import type { SchemeColors } from './color';

export const systemDark =
  typeof window !== 'undefined' && window.matchMedia
    ? window.matchMedia('(prefers-color-scheme: dark)')
    : null;

// isDark resolves the theme mode against the system preference.
export const isDark = derived(config, ($c) => {
  const mode = $c?.themeProps?.themeMode ?? 'system';
  if (mode === 'dark') return true;
  if (mode === 'light') return false;
  return systemDark?.matches ?? false;
});

function varsOf(
  colors: SchemeColors,
  success: string,
  successContainer: string,
  onSuccessContainer: string,
  warning: string,
  warningContainer: string,
  onWarningContainer: string,
  textScale: number,
): Record<string, string> {
  const vars: Record<string, string> = {};
  for (const [role, value] of Object.entries(colors)) {
    if (role === 'brightness') continue;
    vars[`--${camelToKebab(role)}`] = value;
  }
  vars['--success'] = success;
  vars['--success-container'] = successContainer;
  vars['--on-success-container'] = onSuccessContainer;
  vars['--warning'] = warning;
  vars['--warning-container'] = warningContainer;
  vars['--on-warning-container'] = onWarningContainer;
  vars['--text-scale'] = String(textScale);
  vars['--font-scale'] = String(textScale);
  return vars;
}

function camelToKebab(input: string): string {
  return input.replace(/[A-Z]/g, (m) => `-${m.toLowerCase()}`);
}

// themeVars builds the CSS variable map for the current theme state.
export const themeVars = derived([config, isDark], ([$c, $dark]) => {
  const theme = $c?.themeProps;
  const seed = theme?.primaryColor ?? 0xff007aff;
  const style = theme?.colorStyle ?? 'vivid';
  let colors = seededColorScheme(seed, $dark, style);
  if (theme?.pureBlack) colors = toPureBlack(colors);
  const scale = theme?.textScale?.enable ? theme.textScale.scale : 1;
  return varsOf(
    colors,
    successColor($dark),
    successContainerColor($dark),
    onSuccessContainerColor($dark),
    warningColor($dark),
    warningContainerColor($dark),
    onWarningContainerColor($dark),
    scale,
  );
});

// applyTheme writes the variables onto the root element.
export function applyTheme(vars: Record<string, string>, dark: boolean) {
  const root = document.documentElement;
  for (const [name, value] of Object.entries(vars)) {
    root.style.setProperty(name, value);
  }
  root.classList.toggle('dark', dark);
  root.dataset.dark = dark ? 'true' : 'false';
}

if (systemDark) {
  systemDark.addEventListener('change', () => {
    // derived store recomputes via config; force a tick
    config.update((c) => c);
  });
}
