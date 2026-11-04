<script lang="ts">
  // DashboardBlock: a tone block inside a card with its own radius.
  import type { Snippet } from 'svelte';

  let {
    radius = 20,
    color,
    pad = '',
    onclick,
    children,
  }: {
    radius?: number;
    color?: string;
    pad?: string;
    onclick?: () => void;
    children?: Snippet;
  } = $props();
</script>

<div
  class="dash-block"
  style="border-radius: {radius}px; {color ? `background: ${color};` : ''} {pad ? `padding: ${pad};` : ''}"
  onclick={onclick}
  role={onclick ? 'button' : undefined}
  tabindex={onclick ? 0 : undefined}
>
  {@render children?.()}
</div>

<style>
  .dash-block {
    background: var(--surface-container);
    min-width: 0;
    min-height: 0;
  }

  :root.dark .dash-block {
    background: color-mix(in srgb, var(--surface-container-high) 55%, transparent);
  }

  .dash-block[role='button'] {
    cursor: pointer;
    transition: transform 0.15s cubic-bezier(0.2, 0.8, 0.3, 1);
  }

  .dash-block[role='button']:active {
    transform: scale(0.985);
  }
</style>
