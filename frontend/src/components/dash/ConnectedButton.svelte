<script lang="ts">
  // DashboardConnectedButton: M3 Expressive connected pill. At rest the inner
  // corners stay tight; the selected one becomes a full pill filled primary.
  import type { Snippet } from 'svelte';

  let {
    first = false,
    last = false,
    selected = false,
    outer = 24,
    inner = 10,
    onclick,
    children,
  }: {
    first?: boolean;
    last?: boolean;
    selected?: boolean;
    outer?: number;
    inner?: number;
    onclick?: () => void;
    children?: Snippet;
  } = $props();

  const radius = $derived.by(() => {
    const start = first ? outer : inner;
    const end = last ? outer : inner;
    return selected
      ? `${outer}px`
      : `${start}px ${end}px ${end}px ${start}px`;
  });
</script>

<button class="connected" class:selected style="border-radius: {radius};" onclick={onclick}>
  {@render children?.()}
</button>

<style>
  .connected {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--surface-container);
    color: var(--on-surface-variant);
    transition: background 0.24s cubic-bezier(0.2, 0.8, 0.3, 1), transform 0.16s cubic-bezier(0.2, 0.8, 0.3, 1);
    cursor: pointer;
    -webkit-tap-highlight-color: transparent;
  }

  :root.dark .connected {
    background: color-mix(in srgb, var(--surface-container-high) 55%, transparent);
  }

  .connected:hover {
    box-shadow: inset 0 0 0 99px rgba(0, 0, 0, 0.04);
  }

  .connected.selected {
    background: var(--primary);
    color: var(--on-primary);
  }

  .connected:active {
    transform: scale(0.97);
  }
</style>
