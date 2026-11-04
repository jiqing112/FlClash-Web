<script lang="ts">
  // DashboardCardHeader: 24px tall, glyph 18, title footnote 13 w600.
  import type { Snippet } from 'svelte';

  let {
    glyph,
    title,
    leading,
    trailing,
    chevron = false,
  }: {
    glyph?: string;
    title: string;
    leading?: Snippet;
    trailing?: Snippet;
    chevron?: boolean;
  } = $props();

  const isSvg = $derived(glyph?.startsWith('<') ?? false);
</script>

<div class="header">
  {#if glyph}
    <span class="glyph">
      {#if isSvg}
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html glyph}</svg>
      {:else}
        <span class="emoji">{glyph}</span>
      {/if}
    </span>
  {/if}
  <div class="titles">
    <span class="title ellipsis">{title}</span>
    {#if leading}{@render leading?.()}{/if}
  </div>
  <div class="end">
    {#if trailing}{@render trailing?.()}{/if}
    {#if chevron}
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M9 6l6 6-6 6"/></svg>
    {/if}
  </div>
</div>

<style>
  .header {
    height: 24px;
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .glyph {
    color: var(--on-surface-variant);
    display: inline-flex;
  }

  .emoji {
    font-size: 15px;
    line-height: 1;
  }

  .titles {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .title {
    font-size: 13px;
    font-weight: 600;
    color: var(--on-surface-variant);
    white-space: nowrap;
  }

  .end {
    display: flex;
    align-items: center;
    gap: 4px;
    color: var(--on-surface-variant);
    flex-shrink: 0;
  }
</style>
