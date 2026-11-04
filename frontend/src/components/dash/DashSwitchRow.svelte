<script lang="ts">
  // DashboardSwitchRow: icon (cookie6 when on / circle off) + titles + switch.
  import type { Snippet } from 'svelte';
  import ShapeBadge from './ShapeBadge.svelte';

  let {
    glyph,
    title,
    subtitle,
    value = false,
    iconSize = 40,
    pad = '0 16px',
    radius = 20,
    color,
    onclick,
    onchange,
    children,
    showIcon = true,
  }: {
    glyph?: string;
    title: string;
    subtitle?: string;
    value?: boolean;
    iconSize?: number;
    pad?: string;
    radius?: number;
    color?: string;
    onclick?: () => void;
    onchange?: (value: boolean) => void;
    children?: Snippet;
    showIcon?: boolean;
  } = $props();

  function toggle() {
    onchange?.(!value);
  }
</script>

<div
  class="switch-row"
  style="border-radius: {radius}px; background: {color ?? 'var(--surface-container)'}; padding: {pad};"
  onclick={onclick}
  role={onclick ? 'button' : undefined}
  tabindex={onclick ? 0 : undefined}
>
  {#if showIcon && glyph}
    <span class="leading" style="width:{iconSize}px;height:{iconSize}px">
      {#if value}
        <ShapeBadge shape="cookie6Sided" size={iconSize} color="var(--primary)">
          <svg width={iconSize * 0.5} height={iconSize * 0.5} viewBox="0 0 24 24" fill="none" stroke="var(--on-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html glyph}</svg>
        </ShapeBadge>
      {:else}
        <ShapeBadge shape="circle" size={iconSize} color="var(--surface-bright)">
          <svg width={iconSize * 0.5} height={iconSize * 0.5} viewBox="0 0 24 24" fill="none" stroke="var(--on-surface-variant)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html glyph}</svg>
        </ShapeBadge>
      {/if}
    </span>
  {/if}
  <div class="titles">
    <div class="title ellipsis">{title}</div>
    {#if subtitle}<div class="subtitle ellipsis">{subtitle}</div>{/if}
    {#if children}{@render children()}{/if}
  </div>
  <button
    class="switch"
    class:on={value}
    role="switch"
    aria-checked={value}
    onclick={(event) => { event.stopPropagation(); toggle(); }}
  ></button>
</div>

<style>
  .switch-row {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
    width: 100%;
    min-height: 56px;
  }

  :root.dark .switch-row {
    background: color-mix(in srgb, var(--surface-container-high) 55%, transparent);
  }

  .switch-row[role='button'] {
    cursor: pointer;
  }

  .leading {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }

  .titles {
    flex: 1;
    min-width: 0;
  }

  .title {
    font-size: 15px;
    font-weight: 600;
    line-height: 20px;
  }

  .subtitle {
    font-size: 12px;
    line-height: 16px;
    color: var(--on-surface-variant);
    margin-top: 2px;
  }

  .switch {
    position: relative;
    width: 46px;
    height: 28px;
    border-radius: var(--radius-full);
    background: color-mix(in srgb, var(--on-surface) 18%, transparent);
    transition: background 0.2s ease;
    flex-shrink: 0;
  }

  .switch.on {
    background: var(--primary);
  }

  .switch::after {
    content: '';
    position: absolute;
    top: 3px;
    left: 3px;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.25);
    transition: transform 0.2s ease;
  }

  .switch.on::after {
    transform: translateX(18px);
  }
</style>
