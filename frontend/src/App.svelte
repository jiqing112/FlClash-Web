<script lang="ts">
  // App shell: sidebar navigation on desktop, floating dock on mobile.
  import { onMount } from 'svelte';
  import { loadBootstrap, groups, config, showBanner, banners } from './lib/stores';
  import { startWebSocket, stopWebSocket } from './lib/ws';
  import { themeVars, isDark, applyTheme } from './lib/theme';
  import { pushConfig } from './lib/stores';
  import { t } from './lib/i18n';
  import NavSidebar from './components/NavSidebar.svelte';
import Dashboard from './pages/Dashboard.svelte';
  import Proxies from './pages/Proxies.svelte';
  import Profiles from './pages/Profiles.svelte';
  import Requests from './pages/Requests.svelte';
  import DnsQueries from './pages/DnsQueries.svelte';
  import Logs from './pages/Logs.svelte';
  import Resources from './pages/Resources.svelte';
  import Tools from './pages/Tools.svelte';

  type PageLabel = 'dashboard' | 'proxies' | 'profiles' | 'requests' | 'dns' | 'resources' | 'logs' | 'tools';

  // Icons are inlined SVG body paths (ported from the app's glyph set).
  function glyphIcon(name: string): string {
    const icons: Record<string, string> = {
      dashboard: '<rect x="3.5" y="3.5" width="7" height="8" rx="2"/><rect x="13.5" y="3.5" width="7" height="6" rx="2"/><rect x="3.5" y="14.5" width="7" height="6" rx="2"/><rect x="13.5" y="12.5" width="7" height="8" rx="2"/>',
      proxies: '<circle cx="12" cy="12" r="8.8"/><ellipse cx="12" cy="12" rx="3.8" ry="8.8"/><path d="M3.2 12h17.6"/>',
      profiles: '<rect x="5" y="3" width="14" height="18" rx="2.5"/><path d="M9 8.5h6"/><path d="M9 12.5h6"/><path d="M9 16h4"/>',
      requests: '<circle cx="12" cy="5.2" r="2.4"/><circle cx="5.2" cy="17.6" r="2.4"/><circle cx="18.8" cy="17.6" r="2.4"/><path d="M10.7 7.1 6.4 15.6M13.3 7.1l4.3 8.5M7.6 17.6h8.8"/>',
      dns: '<rect x="3.5" y="3.5" width="17" height="7" rx="2.5"/><rect x="3.5" y="13.5" width="17" height="7" rx="2.5"/><circle cx="7.2" cy="7" r="1.1" fill="currentColor" stroke="none"/><circle cx="7.2" cy="17" r="1.1" fill="currentColor" stroke="none"/>',
      resources: '<rect x="3" y="3.5" width="18" height="5" rx="1.8"/><rect x="4.5" y="10.8" width="15" height="9.7" rx="2.2"/><path d="M10 14.2h4"/>',
      logs: '<rect x="5" y="3" width="14" height="18" rx="2.5"/><path d="M9 8h6M9 12h6M9 16h4"/>',
      tools: '<path d="M14.5 6.5a4.5 4.5 0 0 1 5.6-4.4l-3 3 2.3 2.3 3-3a4.5 4.5 0 0 1-5.8 5.7L7.4 18.3a2.1 2.1 0 1 1-2.3-2.3L13.3 7.9a4.5 4.5 0 0 1 1.2-1.4z"/>',
    };
    return icons[name] ?? '';
  }

  let current = $state<PageLabel>('dashboard');
  // Bumping this remounts the page, so tapping the tab you are already on
  // returns a tools sub-page to its list instead of doing nothing.
  let navEpoch = $state(0);
  let targetProxyGroup = $state('');

  function navigate(label: PageLabel) {
    if (label === current) {
      navEpoch += 1;
      return;
    }
    current = label;
  }
  let sidebarExpanded = $state(true);
  let innerWidth = $state(1200);

  $effect(() => {
    const cfg = $config;
    if (cfg) {
      sidebarExpanded = cfg.appSettingProps.sidebarExpanded;
    }
  });

  $effect(() => {
    applyTheme($themeVars, $isDark);
  });

  const isMobile = $derived(innerWidth <= 600);
  const canExpand = $derived(innerWidth > 840);

  const items = $derived.by(() => {
    const cfg = $config;
    const list: { label: PageLabel; title: string; icon: string; show: boolean }[] = [
      { label: 'dashboard', title: t('dashboard'), icon: glyphIcon('dashboard'), show: true },
      { label: 'proxies', title: t('proxies'), icon: glyphIcon('proxies'), show: $groups.length > 0 },
      { label: 'profiles', title: t('profiles'), icon: glyphIcon('profiles'), show: true },
      { label: 'requests', title: t('liveConnections'), icon: glyphIcon('requests'), show: true },
      { label: 'dns', title: t('dnsQueries'), icon: glyphIcon('dns'), show: true },
      { label: 'resources', title: t('resources'), icon: glyphIcon('resources'), show: true },
      { label: 'logs', title: t('logs'), icon: glyphIcon('logs'), show: cfg?.appSettingProps.openLogs ?? false },
      { label: 'tools', title: t('tools'), icon: glyphIcon('tools'), show: true },
    ];
    return list;
  });

  const railItems = $derived(items.filter((item) => item.show && ['dashboard', 'proxies', 'profiles', 'requests', 'logs', 'tools'].includes(item.label)));
  const dockItems = $derived(
    items.filter((item) => item.show).filter((item) => ['dashboard', 'proxies', 'profiles', 'tools'].includes(item.label)),
  );

  function toggleSidebar() {
    sidebarExpanded = !sidebarExpanded;
    pushConfig({ appSettingProps: { sidebarExpanded } });
  }

  onMount(() => {
    loadBootstrap()
      .then(() => startWebSocket())
      .catch((error) => showBanner(String(error), 'error'));
    const onGotoProxies = (event: Event) => {
      targetProxyGroup = (event as CustomEvent<string>).detail ?? '';
      navEpoch += 1;
      current = 'proxies';
    };
    window.addEventListener('goto-proxies', onGotoProxies);
    const onResize = () => (innerWidth = window.innerWidth);
    onResize();
    window.addEventListener('resize', onResize);
    return () => {
      window.removeEventListener('goto-proxies', onGotoProxies);
      window.removeEventListener('resize', onResize);
      stopWebSocket();
    };
  });
</script>

<svelte:head>
  <title>FlClash</title>
</svelte:head>

{#if $config}
  <div class="shell page-surface">
    {#if !isMobile}
      <NavSidebar
        items={railItems}
        selected={current}
        expanded={canExpand && sidebarExpanded}
        onNavigate={(label) => navigate(label as PageLabel)}
        onToggle={toggleSidebar}
      />
    {/if}
    <main class="content" class:dock-space={isMobile}>
      {#key `${current}:${navEpoch}`}
        <div class="page fade-in">
          {#if current === 'dashboard'}
            <Dashboard />
          {:else if current === 'proxies'}
            <Proxies focusGroup={targetProxyGroup} />
          {:else if current === 'profiles'}
            <Profiles />
          {:else if current === 'requests'}
            <Requests />
          {:else if current === 'dns'}
            <DnsQueries />
          {:else if current === 'logs'}
            <Logs />
          {:else if current === 'resources'}
            <Resources />
          {:else}
            <Tools navigate={(label: string) => navigate(label as PageLabel)} />
          {/if}
        </div>
      {/key}
    </main>
  </div>

  {#if isMobile}
    <div class="dock-wrap">
      <div class="dock">
        {#each dockItems as item (item.label)}
          <button class="dock-item" class:selected={current === item.label} onclick={() => navigate(item.label)}>
            <svg width="21" height="21" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html item.icon}</svg>
            <span>{item.title}</span>
          </button>
        {/each}
      </div>
    </div>
  {/if}
{:else}
  <div class="boot page-surface">
    <div class="boot-ring"></div>
  </div>
{/if}

<!-- banners -->
<div class="banner-stack">
  {#each $banners as banner (banner.id)}
    <div class="banner popup-surface" class:error={banner.level === 'error'}>
      {banner.message}
    </div>
  {/each}
</div>

<style>
  .shell {
    display: flex;
    height: 100vh;
    min-width: 0;
  }

  .sidebar {
    width: var(--sidebar-width);
    display: flex;
    flex-direction: column;
    padding: 12px;
    gap: 4px;
    border-right: 1px solid color-mix(in srgb, var(--on-surface) 8%, transparent);
    transition: width 0.22s cubic-bezier(0.2, 0.8, 0.3, 1);
    flex-shrink: 0;
  }

  .sidebar.collapsed {
    width: var(--sidebar-collapsed-width);
  }

  .toggle {
    margin-bottom: 10px;
  }

  .nav-items {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 12px;
    height: 44px;
    padding: 0 12px;
    border-radius: var(--radius-full);
    color: var(--on-surface-variant);
    font-weight: 590;
    font-size: 0.92rem;
    text-align: left;
    transition: background 0.15s ease, color 0.15s ease;
    position: relative;
    overflow: hidden;
    white-space: nowrap;
  }

  .nav-item:hover {
    background: color-mix(in srgb, var(--on-surface) 7%, transparent);
    color: var(--on-surface);
  }

  .nav-item.selected {
    background: color-mix(in srgb, var(--primary) 13%, transparent);
    color: var(--primary);
  }

  .nav-item.selected::before {
    content: '';
    position: absolute;
    left: 0;
    top: 10px;
    bottom: 10px;
    width: 3.5px;
    border-radius: 4px;
    background: var(--primary);
  }

  .sidebar.collapsed .nav-item {
    justify-content: center;
    padding: 0;
  }

  .sidebar.collapsed .nav-label {
    display: none;
  }

  .nav-label {
    flex: 1;
  }

  .content {
    flex: 1;
    min-width: 0;
    min-height: 0;
    position: relative;
  }

  .content.dock-space {
    padding-bottom: 88px;
  }

  .page {
    height: 100%;
  }

  /* floating dock (mobile) */
  .dock-wrap {
    position: fixed;
    left: 0;
    right: 0;
    bottom: 14px;
    display: flex;
    justify-content: center;
    z-index: 50;
    pointer-events: none;
  }

  .dock {
    pointer-events: auto;
    display: flex;
    gap: 2px;
    padding: 5px;
    border-radius: var(--radius-full);
    background: color-mix(in srgb, var(--surface-container-high) 72%, transparent);
    backdrop-filter: blur(24px);
    -webkit-backdrop-filter: blur(24px);
    box-shadow: var(--shadow-e2);
    border: 1px solid color-mix(in srgb, var(--on-surface) 8%, transparent);
  }

  .dock-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    min-width: 74px;
    padding: 8px 12px;
    border-radius: var(--radius-full);
    color: var(--on-surface-variant);
    font-size: 0.72rem;
    font-weight: 590;
    transition: background 0.18s ease, color 0.18s ease;
  }

  .dock-item.selected {
    background: color-mix(in srgb, var(--primary) 15%, transparent);
    color: var(--primary);
  }

  .boot {
    height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .boot-ring {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    border: 3.5px solid color-mix(in srgb, var(--primary) 20%, transparent);
    border-top-color: var(--primary);
    animation: spin 0.9s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .banner-stack {
    position: fixed;
    top: 16px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    flex-direction: column;
    gap: 8px;
    z-index: 200;
    max-width: min(480px, 90vw);
  }

  .banner {
    padding: 12px 18px;
    font-size: 0.9rem;
    animation: fade-in 0.2s ease;
  }

  .banner.error {
    border-left: 3px solid var(--error);
  }
</style>
