<script lang="ts">
  // Proxies: tab strip per group + node grid, or list mode; delay tests,
  // selection and card style from proxiesStyleProps.
  import { t } from '../lib/i18n';
  import Scaffold from '../components/Scaffold.svelte';
  import PopupMenu from '../components/PopupMenu.svelte';
  import { config, groups, selectedMap, showBanner, delayMap, defaultTestUrl, running, tunnelMode, pushConfig } from '../lib/stores';
  import { api } from '../lib/api';
  import { delayText, regionEmoji } from '../lib/format';
  import { delayColor } from '../lib/color';
    type Group = { name: string; all: { name: string; type: string; 'provider-name'?: string }[]; now?: string; testUrl?: string; type: string };
  type Proxy = Group['all'][number];

  let { focusGroup = '' }: { focusGroup?: string } = $props();

  let currentGroup = $state('');
  let testing = $state(false);
  let search = $state('');
  let unfoldSet = $state<Set<string>>(new Set());

  const style = $derived($config?.proxiesStyleProps ?? { type: 'tab', sortType: 'none', layout: 'standard', iconStyle: 'filled', cardType: 'standard', hideTimeoutProxies: false });

  const visibleGroups = $derived($groups.filter((group) => !group.hidden || $tunnelMode === 'global'));
  const effectiveGroup = $derived.by(() => {
    if (focusGroup && visibleGroups.some((group) => group.name === focusGroup)) {
      return focusGroup;
    }
    if (currentGroup && visibleGroups.some((group) => group.name === currentGroup)) {
      return currentGroup;
    }
    return visibleGroups[0]?.name ?? '';
  });
  const activeGroup = $derived(visibleGroups.find((group) => group.name === effectiveGroup) ?? null);
  const compact = $derived(style.cardType === 'compact');
  const loose = $derived(style.cardType === 'loose');
  const isDark = $derived(document.documentElement.classList.contains('dark'));

  // A group's delay is its real pick's: follow nested groups down.
  function delayStateOf(group: Group, proxy: Proxy): number | undefined {
    if (proxy.all) return undefined;
    let name = proxy.name;
    let url = group.testUrl || $defaultTestUrl;
    for (let depth = 0; depth < 8; depth++) {
      const nested = $groups.find((g) => g.name === name);
      if (!nested) break;
      const pick = $selectedMap[name] || nested.now || '';
      if (!pick) break;
      name = pick;
      if (nested.testUrl) url = nested.testUrl;
    }
    return $delayMap[url]?.[name];
  }

  function selectedNameOf(group: Group): string {
    return $selectedMap[group.name] || group.now || '';
  }

  function filteredIn(group: Group): Proxy[] {
    if (!search.trim()) return group.all;
    const needle = search.trim().toLowerCase();
    return group.all.filter((proxy) => proxy.name.toLowerCase().includes(needle));
  }

  async function testGroup(groupName: string) {
    testing = true;
    try {
      await api.testDelay({ group: groupName });
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    } finally {
      setTimeout(() => (testing = false), 600);
    }
  }

  let testingNodes = $state<Set<string>>(new Set());

  async function testOne(proxyName: string) {
    testingNodes.add(proxyName);
    testingNodes = new Set(testingNodes);
    try {
      await api.testDelay({ proxy: proxyName });
    } catch {
      // ignore
    } finally {
      testingNodes.delete(proxyName);
      testingNodes = new Set(testingNodes);
    }
  }


  function setStyle(patch: Partial<typeof style>) {
    pushConfig({ proxiesStyleProps: patch });
  }

  async function select(groupName: string, proxyName: string) {
    try {
      await api.selectProxy(groupName, proxyName);
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    }
  }

  function toggleUnfold(name: string) {
    const next = new Set(unfoldSet);
    if (next.has(name)) {
      next.delete(name);
    } else {
      next.add(name);
    }
    unfoldSet = next;
  }
</script>

<Scaffold
  title={t('proxies')}
  bind:searchText={search}
  search
  searchPlaceholder={t('search')}
  loading={testing}
  actions={[
    {
      icon: '<path d="M13 3L4 14h6l-1 7 9-11h-6l1-7z"/>',
      label: t('delayTest'),
      primary: true,
      show: style.type === 'tab',
      onClick: () => activeGroup && testGroup(activeGroup.name),
    },
  ]}
>
  {#snippet titleTrailing()}
    <PopupMenu align="end">
      {#snippet children()}
        <button class="icon-button" title={t('proxies')}>
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="5" cy="12" r="1.6" fill="currentColor"/><circle cx="12" cy="12" r="1.6" fill="currentColor"/><circle cx="19" cy="12" r="1.6" fill="currentColor"/></svg>
        </button>
      {/snippet}
      {#snippet menu()}
        <div class="menu-section">{t('style')}</div>
        {#each [{ value: 'tab', label: t('tab') }, { value: 'list', label: t('list') }] as option}
          <button class="popup-item" class:selected={style.type === option.value} onclick={() => setStyle({ type: option.value as any })}>
            {option.label}
          </button>
        {/each}
        <div class="menu-section">{t('sort')}</div>
        {#each [{ value: 'none', label: t('defaultText') }, { value: 'delay', label: t('delay') }, { value: 'name', label: t('name') }] as option}
          <button class="popup-item" class:selected={style.sortType === option.value} onclick={() => setStyle({ sortType: option.value as any })}>
            {option.label}
          </button>
        {/each}
        <div class="menu-section">{t('size')}</div>
        {#each [{ value: 'compact', label: t('compact') }, { value: 'standard', label: t('standard') }, { value: 'loose', label: t('loose') }] as option}
          <button class="popup-item" class:selected={style.cardType === option.value} onclick={() => setStyle({ cardType: option.value as any })}>
            {option.label}
          </button>
        {/each}
        <button class="popup-item" class:selected={style.hideTimeoutProxies} onclick={() => setStyle({ hideTimeoutProxies: !style.hideTimeoutProxies })}>
          {t('hideTimeoutProxies')}
          {#if style.hideTimeoutProxies}
            <svg class="check" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12l5 5L20 7"/></svg>
          {/if}
        </button>
      {/snippet}
    </PopupMenu>
  {/snippet}
  {#snippet children()}
    <div class="proxies">
      {#if visibleGroups.length === 0}
        <div class="empty muted">--</div>
      {:else if style.type === 'tab'}
        <div class="tabs-row">
          <div class="group-tabs">
            {#each visibleGroups as group (group.name)}
              <button class="tab" class:selected={effectiveGroup === group.name} onclick={() => (currentGroup = group.name)}>
                {group.name}
              </button>
            {/each}
          </div>
        </div>
        {#if activeGroup}
          <div class="group-head">
            <div class="summary ellipsis muted">
              {activeGroup.now ? `${activeGroup.name} · ${activeGroup.now}` : activeGroup.name}
              · {activeGroup.all.length}
            </div>
            <div class="spacer"></div>
          </div>
          <div class="node-grid" class:compact class:loose>
            {#each filteredIn(activeGroup) as proxy (proxy.name)}
              {@render ProxyCard(activeGroup, proxy)}
            {/each}
          </div>
        {/if}
      {:else}
        <div class="list-mode">
          {#each visibleGroups as group (group.name)}
            <section class="list-section">
              <div class="section-head" role="presentation">
                <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" class:rotated={!unfoldSet.has(group.name)}><path d="M6 9l6 6 6-6"/></svg>
                <span class="gname">{group.name}</span>
                <span class="muted small-text ellipsis">{selectedNameOf(group)} · {group.all.length}</span>
                <div class="spacer"></div>
                <button class="icon-button small" onclick={(event) => { event.stopPropagation(); testGroup(group.name); }}>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M13 3L4 14h6l-1 7 9-11h-6l1-7z"/></svg>
                </button>
              </div>
              {#if unfoldSet.has(group.name)}
                <div class="node-grid nested" class:compact class:loose>
                  {#each filteredIn(group) as proxy (proxy.name)}
                    {@render ProxyCard(group, proxy)}
                  {/each}
                </div>
              {/if}
            </section>
          {/each}
        </div>
      {/if}
    </div>
  {/snippet}
</Scaffold>

{#snippet ProxyCard(group: Group, proxy: Proxy)}
  {@const selected = selectedNameOf(group) === proxy.name}
  {@const delay = delayStateOf(group, proxy)}
  <div
    class="proxy-card app-card"
    class:selected
    class:compact
    role="button"
    tabindex={0}
    onclick={() => select(group.name, proxy.name)}
    onkeydown={(e) => { if (e.key === 'Enter') select(group.name, proxy.name); }}
    title={proxy.name}
  >
    <div class="node-name ellipsis">{regionEmoji(proxy.name)} {proxy.name}</div>
    <div class="node-meta">
      <span class="node-type ellipsis muted">{proxy.type}</span>
      <button
        class="delay mono"
        class:testing={testingNodes.has(proxy.name)}
        title={t('delayTest')}
        onclick={(event) => { event.stopPropagation(); testOne(proxy.name); }}
      >
        {#if testingNodes.has(proxy.name)}
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin"><path d="M20 12a8 8 0 1 1-2.3-5.7"/><path d="M20 3v4h-4"/></svg>
        {:else if delay !== undefined && delay !== 0}
          <span style="color: {delayColor(delay, isDark)}">{delayText(delay)}</span>
        {:else}
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M13 3L4 14h6l-1 7 9-11h-6l1-7z"/></svg>
        {/if}
      </button>
    </div>
    {#if selected}
      <span class="selection-badge">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12l5 5L20 7"/></svg>
      </span>
    {/if}
  </div>
{/snippet}

<style>
  .proxies {
    height: 100%;
    display: flex;
    flex-direction: column;
    min-height: 0;
    padding: 0 26px 26px;
    box-sizing: border-box;
  }

  .empty {
    padding: 40px;
    text-align: center;
  }

  .tabs-row {
    flex-shrink: 0;
  }

  .group-tabs {
    display: flex;
    gap: 8px;
    overflow-x: auto;
    padding-bottom: 12px;
  }

  .tab {
    height: 38px;
    padding: 0 18px;
    border-radius: var(--radius-full);
    background: color-mix(in srgb, var(--on-surface) 7%, transparent);
    color: var(--on-surface-variant);
    font-weight: 590;
    font-size: 0.9rem;
    white-space: nowrap;
    transition: background 0.16s ease, color 0.16s ease;
  }

  .tab.selected {
    background: var(--primary);
    color: var(--on-primary);
  }

  .group-head {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 0 12px;
  }

  .node-grid {
    flex: 1;
    overflow-y: auto;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
    gap: 10px;
    align-content: start;
    padding-bottom: 8px;
  }

  .node-grid.compact {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  }

  .node-grid.loose {
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  }

  .proxy-card {
    position: relative;
    padding: 14px 14px 12px;
    text-align: left;
    display: flex;
    flex-direction: column;
    gap: 6px;
    border-radius: var(--radius-lg);
  }

  .proxy-card.compact {
    padding: 10px 12px 8px;
  }

  .node-name {
    font-weight: 590;
    font-size: 0.92rem;
    line-height: 1.3;
    white-space: nowrap;
  }

  .node-meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    font-size: 0.8rem;
  }

  .delay {
    font-weight: 600;
    background: none;
    padding: 2px 4px;
    border-radius: 6px;
    display: inline-flex;
    align-items: center;
    gap: 4px;
    flex-shrink: 0;
  }

  .delay:hover {
    background: color-mix(in srgb, var(--on-surface) 8%, transparent);
  }

  .delay.testing .spin {
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .selection-badge {
    position: absolute;
    top: 8px;
    right: 8px;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--primary);
    color: var(--on-primary);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  /* list mode */
  .list-mode {
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .list-section {
    background: color-mix(in srgb, var(--on-surface) 3%, transparent);
    border-radius: var(--radius-xl);
    padding: 8px 12px 12px;
  }

  .section-head {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 8px 4px;
    font-weight: 590;
  }

  .rotated {
    transform: rotate(-90deg);
  }

  .node-grid.nested {
    max-height: 460px;
  }
</style>
