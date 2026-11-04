<script lang="ts">
  // Anchored popup menu. The children snippet is the trigger: clicking it
  // toggles the menu unless it carries its own click handler. Closes on an
  // outside click or Escape.
  import type { Snippet } from 'svelte';

  type MenuItem = {
    label?: string;
    icon?: string;
    onClick?: () => void;
    selected?: boolean;
    danger?: boolean;
    show?: boolean;
    section?: Snippet;
    keepOpen?: boolean;
  };

  let {
    items = [] as MenuItem[],
    open = $bindable(false),
    children,
    menu: menuSnippet,
    align = 'end',
  }: {
    items?: MenuItem[];
    open?: boolean;
    children?: Snippet;
    menu?: Snippet;
    align?: 'start' | 'end';
  } = $props();

  let anchor = $state<HTMLDivElement | null>(null);
  let menu = $state<HTMLDivElement | null>(null);
  let trigger = $state<HTMLDivElement | null>(null);

  const shown = $derived(items.filter((item) => item.show !== false));

  // The trigger wrapper handles the toggle; a click handler on the trigger's
  // own button would also bubble here, so stopping it in onToggleClick is not
  // needed — the document-level closer ignores clicks inside the anchor.
  function onToggleClick(event: MouseEvent) {
    const target = event.target as HTMLElement;
    if (target.closest('[data-menu-no-toggle]')) return;
    open = !open;
  }

  function onDocumentClick(event: MouseEvent) {
    if (!open) return;
    const target = event.target as Node;
    if (!anchor?.contains(target) && !menu?.contains(target)) {
      open = false;
    }
  }

  function onKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') open = false;
  }

  function run(item: MenuItem) {
    if (item.keepOpen) {
      item.onClick?.();
      return;
    }
    item.onClick?.();
    open = false;
  }

  $effect(() => {
    if (open) {
      document.addEventListener('mousedown', onDocumentClick, true);
      document.addEventListener('keydown', onKeydown);
      return () => {
        document.removeEventListener('mousedown', onDocumentClick, true);
        document.removeEventListener('keydown', onKeydown);
      };
    }
  });
</script>

<div class="anchor" bind:this={anchor}>
  <div class="trigger" bind:this={trigger} onclick={onToggleClick} role="presentation">
    {@render children?.()}
  </div>
  {#if open}
    <div class="popup-surface menu rise-in" bind:this={menu} style="{align}: 0">
      {#if menuSnippet}
        {@render menuSnippet()}
      {:else}
        {#each shown as item}
          {#if item.section}
            <div class="section">{@render item.section()}</div>
          {:else}
            <button class="popup-item" class:selected={item.selected} class:danger={item.danger} onclick={() => run(item)}>
              {#if item.icon}
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html item.icon}</svg>
              {/if}
              <span class="ellipsis">{item.label}</span>
              {#if item.selected}
                <svg class="check" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12l5 5L20 7"/></svg>
              {/if}
            </button>
          {/if}
        {/each}
      {/if}
    </div>
  {/if}
</div>

<style>
  .anchor {
    position: relative;
    display: inline-flex;
  }

  .trigger {
    display: inline-flex;
    cursor: pointer;
  }

  .menu {
    position: absolute;
    top: calc(100% + 6px);
    z-index: 80;
    max-height: 60vh;
    overflow-y: auto;
    overscroll-behavior: contain;
  }

  .section {
    padding: 4px 0;
  }

  .section + .section {
    border-top: 1px solid color-mix(in srgb, var(--on-surface) 8%, transparent);
  }

  .check {
    margin-left: auto;
  }
</style>
