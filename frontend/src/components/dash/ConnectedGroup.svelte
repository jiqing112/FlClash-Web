<script lang="ts">
  // DashboardConnectedButton group: M3 Expressive connected pills. Inner
  // corners stay tight at rest; the selected one is a full pill in primary.
  import type { Snippet } from 'svelte';

  let {
    outer = 24,
    inner = 10,
    selected = -1,
    children,
  }: {
    outer?: number;
    inner?: number;
    selected?: number;
    children?: Snippet<{ index: number; selected: boolean }>;
  } = $props();

  let count = 0;
  $inspect(count);
  // count children via snippet props loop below
</script>

<div class="group" style="--outer: {outer}px; --inner: {inner}px">
  {#each Array(0) as _, i (i)}
    {#snippet child()}{/snippet}
  {/each}
  {@render children?.()}
</div>

<style>
  .group {
    display: flex;
    gap: 3px;
    align-items: stretch;
    min-width: 0;
  }
</style>
