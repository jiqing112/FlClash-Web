<script lang="ts">
  // ShapeBadge: an M3 expressive shape (cookie/clover/circle) with content.
  import { materialShapes } from '../../lib/shapes';

  let {
    shape = 'circle',
    size = 48,
    color = 'var(--surface-container-high)',
    rotation = 0,
    children,
  }: {
    shape?: keyof typeof materialShapes;
    size?: number;
    color?: string;
    rotation?: number;
    children?: import('svelte').Snippet;
  } = $props();
</script>

<span class="badge" style="width:{size}px;height:{size}px">
  <svg viewBox="0 0 1 1" style="width:{size}px;height:{size}px;display:block">
    <g transform="rotate({(rotation * 180) / Math.PI} 0.5 0.5)">
      <path d={materialShapes[shape]} fill={color} />
    </g>
  </svg>
  {#if children}
    <span class="content">{@render children()}</span>
  {/if}
</span>

<style>
  .badge {
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .badge svg {
    position: absolute;
    inset: 0;
  }

  .content {
    position: relative;
    z-index: 1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
</style>
