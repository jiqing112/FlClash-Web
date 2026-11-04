<script lang="ts">
  // Page scaffold: large title bar with actions, matching the app's Scaffold.
  import type { Snippet } from 'svelte';
  import { connected, coreConnected } from '../lib/stores';
  import { api } from '../lib/api';
  import { t } from '../lib/i18n';

  let {
    title,
    search = false,
    searchText = $bindable(''),
    searchPlaceholder = '',
    actions = [] as { icon?: string; label?: string; onClick: () => void; primary?: boolean; danger?: boolean; show?: boolean }[],
    titleTrailing,
    children,
    loading = false,
  }: {
    title: string;
    search?: boolean;
    searchText?: string;
    searchPlaceholder?: string;
    actions?: { icon?: string; label?: string; onClick: () => void; primary?: boolean; danger?: boolean; show?: boolean }[];
    titleTrailing?: Snippet;
    children?: Snippet;
    loading?: boolean;
  } = $props();

  let searchOpen = $state(false);
  let searchInput = $state<HTMLDivElement | null>(null);

  const shownActions = $derived(actions.filter((a) => a.show !== false));

  function openSearch() {
    searchOpen = true;
    setTimeout(() => searchInput?.focus(), 30);
  }

  function closeSearch() {
    searchOpen = false;
    searchText = '';
  }

  function onSearchKey(event: KeyboardEvent) {
    if (event.key === 'Escape') closeSearch();
  }

  let restarting = $state(false);

  async function confirmRestart() {
    if (restarting) return;
    if (!window.confirm($coreConnected ? t('forceRestartCoreTip') : t('restartCoreTip'))) return;
    restarting = true;
    try {
      await api.coreRestart();
    } catch {
      // banner surfaces via ws error events
    } finally {
      restarting = false;
    }
  }
</script>

<div class="scaffold page-surface">
  <header class="bar">
    {#if searchOpen}
      <div class="search-box rise-in">
        <input
          bind:this={searchInput}
          bind:value={searchText}
          placeholder={searchPlaceholder}
          onkeydown={onSearchKey}
        />
        <button class="icon-button small" onclick={closeSearch} title="Close">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"/></svg>
        </button>
      </div>
    {:else}
      <h1 class="page-title ellipsis">{title}</h1>
      <div class="spacer"></div>
      {#if search}
        <button class="icon-button" onclick={openSearch} title="Search">
          <svg width="21" height="21" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="11" cy="11" r="7"/><path d="M20 20l-3.2-3.2"/></svg>
        </button>
      {/if}
      {#if titleTrailing}{@render titleTrailing()}{/if}
      {#if title === '仪表盘' || title === 'Dashboard'}
        <button
          class="core-status"
          class:ok={$coreConnected && !restarting}
          class:bad={!$coreConnected && !restarting}
          title={$coreConnected ? t('connected') : t('disconnected')}
          onclick={confirmRestart}
        >
          {#if restarting}
            <span class="spinner"></span>
          {:else if $coreConnected}
            <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><path d="M8.6 16.6 4 12l1.4-1.4 3.2 3.2 9.2-9.2L19.2 6z"/></svg>
          {:else}
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 12a8 8 0 1 1-2.3-5.7"/><path d="M20 3v4h-4"/></svg>
          {/if}
        </button>
      {/if}
      {#each shownActions as action (action.label ?? action.icon)}
        {#if action.icon && !action.label}
          <button class="icon-button" class:primary={action.primary} onclick={action.onClick} title={action.label ?? ''}>
            <svg width="21" height="21" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html action.icon}</svg>
          </button>
        {:else}
          <button class="tonal-button small" class:filled={action.primary} class:destructive={action.danger} onclick={action.onClick}>
            {#if action.icon}
              <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html action.icon}</svg>
            {/if}
            {action.label}
          </button>
        {/if}
      {/each}
    {/if}
    {#if loading}
      <div class="bar-progress"><div></div></div>
    {/if}
  </header>
  <div class="body">
    {@render children?.()}
  </div>
</div>

<style>
  .scaffold {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-width: 0;
    position: relative;
  }

  .bar {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 20px 26px 14px;
    position: relative;
    flex-shrink: 0;
  }

  .bar h1 {
    margin: 0;
  }

  .search-box {
    display: flex;
    align-items: center;
    gap: 6px;
    flex: 1;
    max-width: 560px;
  }

  .search-box input {
    flex: 1;
    background: color-mix(in srgb, var(--on-surface) 8%, transparent);
    border: none;
    border-radius: var(--radius-full);
    padding: 10px 18px;
    outline: none;
  }

.core-status {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: var(--error-container);
    color: var(--on-error-container);
    flex-shrink: 0;
    transition: background 0.2s ease;
  }

  .core-status.ok {
    background: var(--success-container);
    color: var(--on-success-container);
  }

  .core-status.bad {
    background: var(--error-container);
    color: var(--on-error-container);
  }

  .spinner {
    width: 18px;
    height: 18px;
    border-radius: 50%;
    border: 2.5px solid color-mix(in srgb, var(--on-surface) 20%, transparent);
    border-top-color: var(--primary);
    animation: core-spin 0.9s linear infinite;
  }

  @keyframes core-spin {
    to { transform: rotate(360deg); }
  }

  .bar-progress {
    position: absolute;
    left: 26px;
    right: 26px;
    bottom: 6px;
    height: 3px;
    border-radius: 3px;
    overflow: hidden;
    background: color-mix(in srgb, var(--primary) 15%, transparent);
  }

  .bar-progress > div {
    height: 100%;
    width: 40%;
    border-radius: 3px;
    background: var(--primary);
    animation: bar-slide 1.1s ease-in-out infinite;
  }

  @keyframes bar-slide {
    0% { transform: translateX(-100%); }
    100% { transform: translateX(260%); }
  }

  .body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }
</style>
