<script lang="ts">
  // Desktop dashboard grid per dashboard.dart: rows 208 / 64 / 176, tables
  // flex 3:2, gap 16 (desktop) with 440px max columns; mobile column + editor.
  import { t } from '../lib/i18n';
  import { config, pushConfig } from '../lib/stores';
  import Scaffold from '../components/Scaffold.svelte';
  import CurrentOutboundCard from './dashboard/CurrentOutboundCard.svelte';
  import SpeedCard from './dashboard/SpeedCard.svelte';
  import OutboundModeBar from './dashboard/OutboundModeBar.svelte';
  import NetworkTakeoverTile from './dashboard/NetworkTakeoverTile.svelte';
  import IntranetIpCard from './dashboard/IntranetIpCard.svelte';
  import ServiceStatusCard from './dashboard/ServiceStatusCard.svelte';
  import ProfilesCard from './dashboard/ProfilesCard.svelte';
  import ProxyGroupsCard from './dashboard/ProxyGroupsCard.svelte';
  import ConnectionStatsCard from './dashboard/ConnectionStatsCard.svelte';
  import ConnectionsTableCard from './dashboard/ConnectionsTableCard.svelte';
  import MemoryInfoCard from './dashboard/MemoryInfoCard.svelte';

  type Widget = string;

  let editMode = $state(false);
  let innerWidth = $state(1440);
  let innerHeight = $state(900);

  $effect(() => {
    const onResize = () => {
      innerWidth = window.innerWidth;
      innerHeight = window.innerHeight;
    };
    onResize();
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  });

  const widgets = $derived($config?.appSettingProps.dashboardWidgets ?? []);
  const isMobile = $derived(innerWidth <= 600);
  const gap = $derived(innerWidth > 840 ? 16 : innerWidth > 600 ? 12 : 8);
  const columns = $derived(innerWidth >= 860 ? 4 : 2);
  const density = $derived(
    innerWidth < 640 ? 'mobile' : columns === 4 && innerWidth < 1040 ? 'compact' : 'standard',
  );

  const topHeight = $derived(density === 'compact' ? 176 : density === 'mobile' ? 208 : 208);
  const controlHeight = $derived(density === 'compact' ? 56 : 64);
  const middleHeight = $derived(density === 'compact' ? 232 : 236);
  const summaryHeight = $derived(density === 'compact' ? 112 : 120);
  const tablesMinHeight = 152;

  function saveWidgets(list: Widget[]) {
    pushConfig({ appSettingProps: { dashboardWidgets: list } });
  }

  function removeWidget(name: Widget) {
    saveWidgets(widgets.filter((item) => item !== name));
  }

  function moveWidget(name: Widget, delta: number) {
    const index = widgets.indexOf(name);
    const target = index + delta;
    if (target < 0 || target >= widgets.length) return;
    const next = [...widgets];
    next.splice(index, 1);
    next.splice(target, 0, name);
    saveWidgets(next);
  }

  function addWidget(name: Widget) {
    if (!widgets.includes(name)) saveWidgets([...widgets, name]);
  }

  const has = $derived((name: Widget) => widgets.includes(name));
</script>

<Scaffold title={t('dashboard')} actions={[{ icon: '<path d="M4 20h4L20.5 7.5a2.1 2.1 0 0 0-3-3L5 17l-1 4z"/><path d="M14.5 6l3 3"/>', label: editMode ? t('done') : undefined, primary: editMode, show: isMobile, onClick: () => (editMode = !editMode) }]}>
  {#snippet children()}
    <div class="dash-scroll">
      {#if isMobile}
        <div class="mobile-list" style="gap: {gap}px">
          {#each widgets as widget (widget)}
            <div class="mobile-card" class:editing={editMode}>
              {#if widget === 'currentOutbound'}
                <CurrentOutboundCard mobile />
              {:else if widget === 'networkSpeed'}
                <SpeedCard direction="upload" mobile />
                <div style="height: {gap}px"></div>
                <SpeedCard direction="download" mobile />
              {:else if widget === 'outboundMode'}
                <OutboundModeBar />
              {:else if widget === 'networkTakeover'}
                <NetworkTakeoverTile kind="systemProxy" />
                <div style="height: {gap}px"></div>
                <NetworkTakeoverTile kind="tun" />
              {:else if widget === 'intranetIp'}
                <IntranetIpCard mobile />
              {:else if widget === 'serviceStatus'}
                <ServiceStatusCard />
              {:else if widget === 'profiles'}
                <ProfilesCard mobile />
              {:else if widget === 'proxyGroups'}
                <ProxyGroupsCard mobile />
              {:else if widget === 'connectionStats'}
                <ConnectionStatsCard mobile />
              {:else if widget === 'memoryInfo'}
                <MemoryInfoCard mobile />
              {/if}
              {#if editMode}
                <div class="edit-tools">
                  <button class="icon-button small" title="Up" onclick={() => moveWidget(widget, -1)}>↑</button>
                  <button class="icon-button small" title="Down" onclick={() => moveWidget(widget, 1)}>↓</button>
                  {#if widget !== 'currentOutbound'}
                    <button class="icon-button small" title="Remove" onclick={() => removeWidget(widget)}>✕</button>
                  {/if}
                </div>
              {/if}
            </div>
          {/each}
        </div>
      {:else if columns === 4}
        <div class="grid" style="gap: {gap}px">
          {#if has('currentOutbound')}
            <div class="row" style="height: {topHeight}px; gap: {gap}px">
              <div class="cell span2"><CurrentOutboundCard /></div>
              <div class="cell"><SpeedCard direction="upload" /></div>
              <div class="cell"><SpeedCard direction="download" /></div>
            </div>
          {/if}
          {#if has('outboundMode') || has('networkTakeover')}
            <div class="row" style="height: {controlHeight}px; gap: {gap}px">
              {#if has('outboundMode')}
                <div class="cell span2"><OutboundModeBar /></div>
              {/if}
              {#if has('networkTakeover')}
                <div class="cell"><NetworkTakeoverTile kind="systemProxy" /></div>
                <div class="cell"><NetworkTakeoverTile kind="tun" /></div>
              {/if}
            </div>
          {/if}
          {#if has('intranetIp') || has('profiles') || has('serviceStatus') || has('memoryInfo')}
            <div class="row" style="height: {middleHeight}px; gap: {gap}px">
              {#if has('intranetIp')}<div class="cell"><IntranetIpCard /></div>{/if}
              {#if has('profiles')}<div class="cell"><ProfilesCard /></div>{/if}
              {#if has('serviceStatus')}<div class="cell"><ServiceStatusCard /></div>{/if}
              {#if has('memoryInfo')}<div class="cell"><MemoryInfoCard /></div>{/if}
            </div>
          {/if}
          {#if has('connectionStats') || has('proxyGroups')}
            <div class="tables" style="gap: {gap}px; min-height: {tablesMinHeight}px">
              {#if has('connectionStats')}
                <div class="cell" style="flex: 3"><ConnectionsTableCard /></div>
              {/if}
              {#if has('proxyGroups')}
                <div class="cell" style="flex: 2"><ProxyGroupsCard /></div>
              {/if}
            </div>
          {/if}
        </div>
      {:else}
        <div class="two-col" style="gap: {gap}px">
          {#if has('currentOutbound')}
            <div class="row" style="height: {topHeight}px; gap: {gap}px"><CurrentOutboundCard /></div>
          {/if}
          {#if has('networkSpeed')}
            <div class="row" style="height: {topHeight}px; gap: {gap}px">
              <div class="cell"><SpeedCard direction="upload" /></div>
              <div class="cell"><SpeedCard direction="download" /></div>
            </div>
          {/if}
          {#if has('outboundMode')}
            <div class="row" style="height: {controlHeight}px"><OutboundModeBar /></div>
          {/if}
          {#if has('networkTakeover')}
            <div class="row" style="height: {controlHeight}px; gap: {gap}px">
              <div class="cell"><NetworkTakeoverTile kind="systemProxy" /></div>
              <div class="cell"><NetworkTakeoverTile kind="tun" /></div>
            </div>
          {/if}
          {#if has('intranetIp')}
            <div class="row" style="height: {middleHeight}px"><IntranetIpCard /></div>
          {/if}
          {#if has('profiles')}
            <div class="row" style="height: {middleHeight}px"><ProfilesCard /></div>
          {/if}
          {#if has('serviceStatus')}
            <div class="row" style="height: {middleHeight}px"><ServiceStatusCard /></div>
          {/if}
          {#if has('memoryInfo')}
            <div class="row" style="height: {middleHeight}px"><MemoryInfoCard /></div>
          {/if}
          {#if has('connectionStats')}
            <div class="row" style="height: {summaryHeight}px"><ConnectionStatsCard mobile /></div>
          {/if}
          {#if has('proxyGroups')}
            <div class="row" style="height: {summaryHeight}px"><ProxyGroupsCard mobile /></div>
          {/if}
        </div>
      {/if}
    </div>
  {/snippet}
</Scaffold>

<style>
  .dash-scroll {
    height: 100%;
    overflow-y: auto;
    padding: 4px 24px 24px 16px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
  }

  .grid,
  .two-col {
    display: flex;
    flex-direction: column;
    max-width: calc(440px * 4 + 16px * 3 + 48px);
    /* bound to the viewport so the tables row flexes like the original:
       rows keep their fixed heights, the feed card absorbs the rest */
    flex: 1 1 auto;
    min-height: min(100%, 0px);
  }

  .row {
    display: flex;
    align-items: stretch;
  }

  .cell {
    flex: 1;
    min-width: 0;
    display: flex;
  }

  .cell :global(.dash-card) {
    flex: 1;
    width: 100%;
  }

  .span2 {
    flex: 2;
  }

  .tables {
    display: flex;
    align-items: stretch;
    flex: 1;
    min-height: 152px;
    overflow: hidden;
  }

  .mobile-list {
    display: flex;
    flex-direction: column;
    max-width: 560px;
    margin: 0 auto;
    width: 100%;
    min-width: 0;
  }

  .mobile-card {
    position: relative;
    min-width: 0;
  }

  .mobile-card :global(.dash-card) {
    box-shadow: var(--shadow-e1);
  }

  .mobile-card.editing {
    animation: sway 2.4s ease-in-out infinite;
  }

  @keyframes sway {
    0%, 100% { transform: rotate(-0.5deg); }
    50% { transform: rotate(0.5deg); }
  }

  .edit-tools {
    position: absolute;
    top: 10px;
    right: 12px;
    display: flex;
    gap: 4px;
    background: var(--surface-bright);
    border-radius: var(--radius-full);
    padding: 2px;
    box-shadow: var(--shadow-e1);
    z-index: 5;
  }
</style>
