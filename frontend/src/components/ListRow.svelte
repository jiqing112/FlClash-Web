<script lang="ts">
  // Config list row: leading glyph + titles + trailing control.
  import type { Snippet } from 'svelte';

  let {
    glyph,
    title,
    subtitle,
    value,
    clickable = false,
    selected = false,
    onclick,
    children,
    trailing,
  }: {
    glyph?: string;
    title?: string;
    subtitle?: string;
    value?: string;
    clickable?: boolean;
    selected?: boolean;
    onclick?: (event: MouseEvent) => void;
    children?: Snippet;
    trailing?: Snippet;
  } = $props();

  const isSvg = $derived(glyph?.startsWith('<') ?? false);
</script>

<div class="list-row" class:clickable={clickable || onclick} class:selected role={onclick ? 'button' : undefined} tabindex={onclick ? 0 : undefined} onclick={onclick} onkeydown={(e) => { if (onclick && (e.key === 'Enter' || e.key === ' ')) onclick(e as unknown as MouseEvent); }}>
  {#if glyph}
    <span class="leading">
      {#if isSvg}
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html glyph}</svg>
      {:else if glyph}
        <span class="emoji">{glyph}</span>
      {/if}
    </span>
  {/if}
  <div class="titles">
    {#if title}<div class="title">{title}</div>{/if}
    {#if subtitle}<div class="subtitle">{subtitle}</div>{/if}
    {@render children?.()}
  </div>
  {#if value}
    <span class="trailing"><span class="value">{value}</span></span>
  {/if}
  {@render trailing?.()}
</div>

<style>
  .value {
    font-size: 0.9rem;
  }

  .emoji {
    font-size: 19px;
    line-height: 1;
    display: inline-flex;
  }

  .list-row[role='button'] {
    cursor: pointer;
  }

  .list-row.selected {
    background: color-mix(in srgb, var(--primary) 10%, transparent);
  }

  .list-row.selected .title {
    color: var(--primary);
  }
</style>
