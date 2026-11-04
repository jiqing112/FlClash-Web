<script lang="ts">
  import {
    glyphs,
    glyphSvgPaths,
    GLYPH_SIZE,
    GLYPH_STROKE_WIDTH,
    type GlyphDef,
    type GlyphName,
  } from '../lib/glyphs';

  interface Props {
    /** Glyph definition, or the name of a glyph in the `glyphs` record. */
    glyph: GlyphDef | GlyphName;
    /** Rendered edge length in px; the 24-unit grid scales to fit, so strokes scale proportionally. */
    size?: number;
    /** Stroke width (in 24-grid user units) for body and line shapes. */
    strokeWidth?: number;
    /** Stroke width for detail shapes; defaults to `strokeWidth` (the Dart painter uses the same 1.8 at rest). */
    detailWidth?: number;
    class?: string;
  }

  let {
    glyph,
    size = GLYPH_SIZE,
    strokeWidth = GLYPH_STROKE_WIDTH,
    detailWidth = strokeWidth,
    class: className = '',
  }: Props = $props();

  const def = $derived(typeof glyph === 'string' ? glyphs[glyph] : glyph);
  const paths = $derived(glyphSvgPaths(def));
</script>

<svg
  viewBox="0 0 {GLYPH_SIZE} {GLYPH_SIZE}"
  width={size}
  height={size}
  class={className}
  fill="none"
  stroke="currentColor"
  stroke-linecap="round"
  stroke-linejoin="round"
  aria-hidden="true"
>
  {#if paths.body}
    <path d={paths.body} stroke-width={strokeWidth} />
  {/if}
  {#if paths.detail}
    <path d={paths.detail} stroke-width={detailWidth} />
  {/if}
  {#if paths.line}
    <path d={paths.line} stroke-width={strokeWidth} />
  {/if}
  {#each paths.solid as d, i (i)}
    <path {d} fill="currentColor" stroke="none" />
  {/each}
</svg>
