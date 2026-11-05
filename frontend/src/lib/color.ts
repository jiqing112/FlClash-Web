// Port of prism_ui/lib/color.dart `seededColorScheme`, on top of the official
// material-color-utilities implementation, so the palette matches the desktop
// app exactly.
import {
  Hct,
  DynamicScheme,
  SchemeContent,
  SchemeTonalSpot,
  SchemeMonochrome,
  MaterialDynamicColors,
  TonalPalette,
  argbFromHex,
  hexFromArgb,
} from '@material/material-color-utilities';

export type ColorStyle = 'vivid' | 'tinted' | 'soft' | 'monochrome';

const SURFACE_CHROMA = 3.0;
const PAPER_CHROMA = 1.5;
const TINT_DIVISOR = 6.0;
const MAX_TINT_CHROMA = 10.0;
const PRIMARY_CONTAINER_CHROMA = 48.0;
const DARK_ON_SURFACE_TONE = 97;
const DARK_CARD_TONE = 15.0;
const DARK_BLOCK_TONE = 21.0;
const DARK_INSET_TONE = 27.0;
const MIN_ACCENT_TONE = 35.0;
const MAX_ACCENT_TONE = 60.0;
const DARK_ACCENT_LIFT = 3.0;
const MIN_DARK_ACCENT_TONE = 50.0;
const MAX_DARK_ACCENT_TONE = 75.0;

const SUCCESS_PALETTE = TonalPalette.fromHct(Hct.fromInt(0xFF4CAF50));
const WARNING_PALETTE = TonalPalette.fromHct(Hct.fromInt(0xFFFF9800));

const clamp = (v: number, lo: number, hi: number) => Math.min(hi, Math.max(lo, v));

const paletteTone = (p: TonalPalette, tone: number) => hexFromArgb(p.tone(tone));

function schemeOf(source: Hct, dark: boolean, style: ColorStyle): DynamicScheme {
  switch (style) {
    case 'vivid':
    case 'tinted':
      return new SchemeContent(source, dark, 0.0);
    case 'soft':
      return new SchemeTonalSpot(source, dark, 0.0);
    case 'monochrome':
      return new SchemeMonochrome(source, dark, 0.0);
  }
}

function accent(seedColor: number, seed: Hct, dark: boolean): { primary: string; onPrimary: string } {
  const tone = dark
    ? clamp(seed.tone + DARK_ACCENT_LIFT, MIN_DARK_ACCENT_TONE, MAX_DARK_ACCENT_TONE)
    : clamp(seed.tone, MIN_ACCENT_TONE, MAX_ACCENT_TONE);
  const primary =
    tone === seed.tone
      ? hexFromArgb(seedColor)
      : hexFromArgb(Hct.from(seed.hue, seed.chroma, tone).toInt());
  const onPrimary =
    tone <= MAX_ACCENT_TONE
      ? '#ffffff'
      : paletteTone(TonalPalette.fromHueAndChroma(seed.hue, seed.chroma), 10);
  return { primary, onPrimary };
}

export interface SchemeColors {
  [role: string]: string;
  brightness: 'light' | 'dark';
}

// seededColorScheme: every role the app's ColorScheme carries.
export function seededColorScheme(seedColor: number, dark: boolean, style: ColorStyle): SchemeColors {
  const source = Hct.fromInt(seedColor);
  const scheme = schemeOf(source, dark, style);
  type DynamicColorLike = { getArgb(s: DynamicScheme): number; getTone(s: DynamicScheme): number; getHct(s: DynamicScheme): import('@material/material-color-utilities').Hct };
  const role = (dc: DynamicColorLike): string => hexFromArgb(dc.getArgb(scheme));

  const primaryPalette = scheme.primaryPalette;
  const containerHue = primaryPalette.hue;
  const containerChroma = Math.min(primaryPalette.chroma, PRIMARY_CONTAINER_CHROMA);
  const container = TonalPalette.fromHueAndChroma(containerHue, containerChroma);
  const tint =
    style === 'tinted' ? Math.min(source.chroma / TINT_DIVISOR, MAX_TINT_CHROMA) : null;

  const neutral = (
    dc: DynamicColorLike,
    maxChroma: number,
    darkTone?: number,
  ): string => {
    const hct = dc.getHct(scheme);
    const tone = dark && darkTone != null ? darkTone : dc.getTone(scheme);
    return hexFromArgb(Hct.from(hct.hue, tint ?? Math.min(hct.chroma, maxChroma), tone).toInt());
  };

  const paperChroma = dark ? SURFACE_CHROMA : PAPER_CHROMA;
  const useAccent = style === 'vivid' || style === 'tinted';
  const acc = useAccent ? accent(seedColor, source, dark) : null;
  const primary = acc ? acc.primary : role(MaterialDynamicColors.primary);
  const onPrimary = acc ? acc.onPrimary : role(MaterialDynamicColors.onPrimary);

  const colors: SchemeColors = {
    brightness: dark ? 'dark' : 'light',
    primary,
    onPrimary,
    surfaceTint: primary,
    primaryContainer: paletteTone(container, dark ? 30 : 90),
    onPrimaryContainer: paletteTone(container, dark ? 90 : 30),
    secondary: role(MaterialDynamicColors.secondary),
    onSecondary: role(MaterialDynamicColors.onSecondary),
    secondaryContainer: role(MaterialDynamicColors.secondaryContainer),
    onSecondaryContainer: role(MaterialDynamicColors.onSecondaryContainer),
    tertiary: role(MaterialDynamicColors.tertiary),
    onTertiary: role(MaterialDynamicColors.onTertiary),
    tertiaryContainer: role(MaterialDynamicColors.tertiaryContainer),
    onTertiaryContainer: role(MaterialDynamicColors.onTertiaryContainer),
    error: role(MaterialDynamicColors.error),
    onError: role(MaterialDynamicColors.onError),
    errorContainer: role(MaterialDynamicColors.errorContainer),
    onErrorContainer: role(MaterialDynamicColors.onErrorContainer),
    surface: neutral(MaterialDynamicColors.surface as any, paperChroma),
    surfaceBright: neutral(MaterialDynamicColors.surfaceBright as any, paperChroma),
    surfaceDim: neutral(MaterialDynamicColors.surfaceDim as any, SURFACE_CHROMA),
    surfaceContainerLowest: neutral(MaterialDynamicColors.surfaceContainerLowest as any, SURFACE_CHROMA),
    surfaceContainerLow: neutral(MaterialDynamicColors.surfaceContainerLow as any, SURFACE_CHROMA),
    surfaceContainer: neutral(MaterialDynamicColors.surfaceContainer as any, SURFACE_CHROMA, DARK_CARD_TONE),
    surfaceContainerHigh: neutral(MaterialDynamicColors.surfaceContainerHigh as any, SURFACE_CHROMA, DARK_BLOCK_TONE),
    surfaceContainerHighest: neutral(MaterialDynamicColors.surfaceContainerHighest as any, SURFACE_CHROMA, DARK_INSET_TONE),
    onSurface:
      dark
        ? hexFromArgb(Hct.from(
            (MaterialDynamicColors.onSurface as any).getHct(scheme).hue,
            (MaterialDynamicColors.onSurface as any).getHct(scheme).chroma,
            DARK_ON_SURFACE_TONE,
          ).toInt())
        : role(MaterialDynamicColors.onSurface),
    onSurfaceVariant: role(MaterialDynamicColors.onSurfaceVariant),
    outline: role(MaterialDynamicColors.outline),
    inverseSurface: role(MaterialDynamicColors.inverseSurface),
    onInverseSurface: role(MaterialDynamicColors.inverseOnSurface),
    inversePrimary: role(MaterialDynamicColors.inversePrimary),
  };
  // outlineVariant with chroma 4
  const ov = MaterialDynamicColors.outlineVariant as any;
  const ovHct = ov.getHct(scheme);
  colors.outlineVariant = hexFromArgb(Hct.from(ovHct.hue, 4, ov.getTone(scheme)).toInt());
  return colors;
}

// toPureBlack flattens dark surfaces to black, mirroring the Dart extension.
export function toPureBlack(colors: SchemeColors): SchemeColors {
  if (colors.brightness !== 'dark') return colors;
  const shift = toneOf(colors.surface);
  const lower = (hexColor: string) => shiftTone(hexColor, -shift);
  return {
    ...colors,
    surface: '#000000',
    surfaceDim: '#000000',
    surfaceContainerLowest: '#000000',
    surfaceContainerLow: lower(colors.surfaceContainerLow),
    surfaceContainer: lower(colors.surfaceContainer),
    surfaceContainerHigh: lower(colors.surfaceContainerHigh),
    surfaceContainerHighest: lower(colors.surfaceContainerHighest),
    surfaceBright: lower(colors.surfaceBright),
  };
}

export function successColor(dark: boolean): string {
  return paletteTone(SUCCESS_PALETTE, dark ? 80 : 40);
}

export function successContainerColor(dark: boolean): string {
  return paletteTone(SUCCESS_PALETTE, dark ? 30 : 90);
}

export function onSuccessContainerColor(dark: boolean): string {
  return paletteTone(SUCCESS_PALETTE, dark ? 90 : 30);
}

export function warningColor(dark: boolean): string {
  return paletteTone(WARNING_PALETTE, dark ? 80 : 40);
}

export function warningContainerColor(dark: boolean): string {
  return paletteTone(WARNING_PALETTE, dark ? 30 : 90);
}

export function onWarningContainerColor(dark: boolean): string {
  return paletteTone(WARNING_PALETTE, dark ? 90 : 30);
}

// delayColor: the app's green→red scale for latency values.
export function delayColor(delay: number, dark: boolean): string {
  if (delay < 0) {
    return dark ? '#8e8e93' : '#8e8e93';
  }
  if (delay <= 200) return successColor(dark);
  if (delay <= 400) return paletteTone(SUCCESS_PALETTE, dark ? 70 : 45);
  if (delay <= 600) return paletteTone(WARNING_PALETTE, dark ? 75 : 45);
  if (delay <= 1000) return paletteTone(TonalPalette.fromHueAndChroma(25, 90), dark ? 75 : 45);
  return paletteTone(TonalPalette.fromHueAndChroma(4, 90), dark ? 75 : 45);
}

// ---- tone helpers (HCT round trip) ----

function toneOf(hexColor: string): number {
  return Hct.fromInt(argbFromHex(hexColor)).tone;
}

function shiftTone(hexColor: string, delta: number): string {
  const hct = Hct.fromInt(argbFromHex(hexColor));
  return hexFromArgb(Hct.from(hct.hue, hct.chroma, clamp(hct.tone + delta, 0, 100)).toInt());
}

export { argbFromHex, hexFromArgb };
