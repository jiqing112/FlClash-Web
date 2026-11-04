<script lang="ts">
  // NavigationSidebar per prism_ui: 220/48 widths, 40px rows, gray selected
  // pill with a 3x16 primary indicator, onSurface labels.
  let {
    items,
    selected,
    expanded = true,
    onNavigate,
    onToggle,
  }: {
    items: { label: string; title: string; icon: string; show: boolean }[];
    selected: string;
    expanded: boolean;
    onNavigate: (label: string) => void;
    onToggle: () => void;
  } = $props();

  const shown = $derived(items.filter((item) => item.show));
</script>

<nav class="sidebar chrome-surface" class:collapsed={!expanded}>
  <button class="item toggle" onclick={onToggle} title="Toggle sidebar">
    <span class="icon">
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="16" rx="3"/><path d="M9.5 4v16" /></svg>
    </span>
    {#if expanded}<span class="label ellipsis"></span>{/if}
  </button>
  <div class="items">
    {#each shown as item (item.label)}
      <button
        class="item"
        class:selected={selected === item.label}
        onclick={() => onNavigate(item.label)}
        title={item.title}
      >
        <span class="icon">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html item.icon}</svg>
        </span>
        {#if expanded}<span class="label ellipsis">{item.title}</span>{/if}
        {#if selected === item.label}
          <span class="indicator"></span>
        {/if}
      </button>
    {/each}
  </div>
</nav>

<style>
  .sidebar {
    width: var(--sidebar-width);
    display: flex;
    flex-direction: column;
    padding: 4px 0;
    border-right: 1px solid color-mix(in srgb, var(--on-surface) 8%, transparent);
    transition: width 0.25s cubic-bezier(0.2, 0.8, 0.3, 1);
    flex-shrink: 0;
    overflow: hidden;
  }

  .sidebar.collapsed {
    width: var(--sidebar-collapsed-width);
  }

  .items {
    display: flex;
    flex-direction: column;
    gap: 4px;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .item {
    position: relative;
    display: flex;
    align-items: center;
    height: var(--sidebar-item);
    margin: 0 12px;
    padding: 0 12px;
    border-radius: 10px;
    color: var(--on-surface);
    font-size: 15px;
    text-align: left;
    white-space: nowrap;
    flex-shrink: 0;
    transition: background 0.15s ease;
  }

  .sidebar.collapsed .item {
    justify-content: flex-start;
    margin: 0 12px;
    padding: 0;
    width: calc(var(--sidebar-collapsed-width) - 24px);
  }

  .item:hover {
    background: color-mix(in srgb, var(--on-surface) 4%, transparent);
  }

  .item.selected {
    background: color-mix(in srgb, var(--on-surface) 8%, transparent);
  }

  .icon {
    width: var(--sidebar-icon);
    height: var(--sidebar-icon);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--on-surface-variant);
    flex-shrink: 0;
  }

  .item.selected .icon {
    color: var(--on-surface);
  }

  .label {
    margin-left: 12px;
    min-width: 0;
  }

  .sidebar.collapsed .label {
    display: none;
  }

  .indicator {
    position: absolute;
    left: -12px;
    top: 50%;
    transform: translateY(-50%);
    width: var(--sidebar-indicator-w);
    height: var(--sidebar-indicator-h);
    border-radius: var(--radius-full);
    background: var(--primary);
  }
</style>
