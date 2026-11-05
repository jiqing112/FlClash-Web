<script lang="ts">
  // Anchored popup menu, restyled to prism_ui/lib/popup.dart: card radius 24
  // (item 16 + inset 8), content inset 8, rows 44px with 16px-radius hover
  // pills, width 160-280, sections as 13px titles, check at the row end.
  // The children snippet is the trigger: clicking it toggles the menu unless
  // it carries its own click handler. Closes on an outside click or Escape.
  import type { Snippet } from 'svelte';

  type MenuItem = {
    label?: string;
    subtitle?: string;
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

  const shown = $derived(items.filter((item) => item.show !== false));

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
  <div class="trigger" onclick={onToggleClick} role="presentation">
    {@render children?.()}
  </div>
  {#if open}
    <div class="menu rise-in" bind:this={menu} style="{align}: 0">
      {#if menuSnippet}
        {@render menuSnippet()}
      {:else}
        {#each shown as item}
          {#if item.section}
            <div class="menu-section">{@render item.section()}</div>
          {:else}
            <button
              class="popup-item"
              class:danger={item.danger}
              class:with-check={item.selected}
              onclick={() => run(item)}
            >
              {#if item.icon}
                <span class="item-icon">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html item.icon}</svg>
                </span>
              {/if}
              <span class="item-titles">
                <span class="item-label ellipsis">{item.label}</span>
                {#if item.subtitle}
                  <span class="item-subtitle ellipsis">{item.subtitle}</span>
                {/if}
              </span>
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

  /* The menu card: radius = item 16 + inset 8, content inset 8, width 160-280,
     height fits 7 rows plus a half-row cut that tells it scrolls. */
  .menu {
    position: absolute;
    top: calc(100% + 8px);
    z-index: 80;
    min-width: 200px;
    max-width: 280px;
    padding: 8px;
    border-radius: 24px;
    background: var(--surface-bright);
    box-shadow: var(--shadow-e2);
    max-height: calc(16px + 7.5 * 44px);
    overflow-y: auto;
    overscroll-behavior: contain;
    scrollbar-width: none;
  }

  :root.dark .menu {
    background: var(--surface-container-high);
  }

  .menu::-webkit-scrollbar {
    width: 0;
    height: 0;
  }

  /* rows and section titles are styled globally in app.css so snippet-built
     menus share the same look */

</style>
