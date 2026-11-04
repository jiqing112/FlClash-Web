<script lang="ts">
  // Sheet: right side sheet on desktop, bottom sheet on narrow screens.
  import type { Snippet } from 'svelte';

  let {
    open = $bindable(false),
    title = '',
    children,
    footer,
    width = 480,
  }: {
    open?: boolean;
    title?: string;
    children?: Snippet;
    footer?: Snippet;
    width?: number;
  } = $props();

  let layer = $state<HTMLDivElement | null>(null);

  function onPointerDown(event: PointerEvent) {
    if (event.target === layer) open = false;
  }

  function onKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') open = false;
  }

  $effect(() => {
    if (open) {
      document.addEventListener('keydown', onKeydown);
      return () => document.removeEventListener('keydown', onKeydown);
    }
  });
</script>

{#if open}
  <div class="scrim" role="presentation"></div>
  <div class="sheet-layer" bind:this={layer} onpointerdown={onPointerDown} role="presentation">
    <aside class="sheet rise-in" style="width: min({width}px, 92vw);">
      <header>
        <h2>{title}</h2>
        <button class="icon-button small" onclick={() => (open = false)}>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"/></svg>
        </button>
      </header>
      <div class="sheet-body">
        {@render children?.()}
      </div>
      {#if footer}
        <footer>{@render footer()}</footer>
      {/if}
    </aside>
  </div>
{/if}

<style>
  .sheet-layer {
    position: fixed;
    inset: 0;
    display: flex;
    justify-content: flex-end;
    z-index: 95;
  }

  .sheet {
    height: 100%;
    background: var(--surface-bright);
    box-shadow: var(--shadow-e2);
    display: flex;
    flex-direction: column;
    animation: slide-side 0.24s cubic-bezier(0.2, 0.8, 0.3, 1);
  }

  :root.dark .sheet {
    background: var(--surface-container);
  }

  @keyframes slide-side {
    from {
      transform: translateX(40px);
      opacity: 0;
    }
    to {
      transform: translateX(0);
      opacity: 1;
    }
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 18px 20px 12px;
  }

  header h2 {
    margin: 0;
    font-size: 1.2rem;
    font-weight: 700;
  }

  .sheet-body {
    flex: 1;
    overflow-y: auto;
    padding: 4px 20px 20px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  footer {
    padding: 14px 20px;
    border-top: 1px solid color-mix(in srgb, var(--on-surface) 8%, transparent);
    display: flex;
    justify-content: flex-end;
    gap: 10px;
  }
</style>
