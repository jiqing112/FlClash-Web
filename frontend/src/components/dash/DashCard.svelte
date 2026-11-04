<script lang="ts">
  // DashboardCard: radius 32, e1 shadow, tone colors from style.dart.
  import type { Snippet } from 'svelte';
  import ShapeBadge from './ShapeBadge.svelte';

  let {
    pad = true,
    clip = false,
    selected = false,
    onclick,
    children,
  }: {
    pad?: boolean;
    clip?: boolean;
    selected?: boolean;
    onclick?: () => void;
    children?: Snippet;
  } = $props();
</script>

<div
  class="dash-card"
  class:pad
  class:clip
  class:clickable={onclick != null}
  class:selected
  onclick={onclick}
  role={onclick ? 'button' : undefined}
  tabindex={onclick ? 0 : undefined}
>
  {@render children?.()}
</div>

<style>
  .dash-card {
    background: var(--surface-bright);
    border-radius: var(--dash-radius);
    box-shadow: var(--shadow-e1);
    height: 100%;
    min-width: 0;
    min-height: 0;
    position: relative;
    display: flex;
    flex-direction: column;
  }

  :root.dark .dash-card {
    background: var(--surface-container);
  }

  .dash-card.pad {
    padding: 20px 22px;
  }

  .dash-card.clip {
    overflow: hidden;
  }

  .dash-card.clickable {
    cursor: pointer;
    transition: transform 0.16s cubic-bezier(0.2, 0.8, 0.3, 1);
  }

  .dash-card.clickable:active {
    transform: scale(0.985);
  }
</style>
