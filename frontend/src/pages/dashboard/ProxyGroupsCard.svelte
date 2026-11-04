<script lang="ts">
  // ProxyGroupsCard: group tiles (radius 20, monogram icon, selected name) in
  // 1-2 columns, fading past the card; header count chip + chevron.
  import { t } from '../../lib/i18n';
  import { groups, selectedMap, running, tunnelMode, showBanner } from '../../lib/stores';
  import { api } from '../../lib/api';
  import DashCard from '../../components/dash/DashCard.svelte';
  import DashHeader from '../../components/dash/DashHeader.svelte';
  import DashChip from '../../components/dash/DashChip.svelte';

  let { mobile = false }: { mobile?: boolean } = $props();

  const layersGlyph = '<path d="M12 3l8 4.5-8 4.5-8-4.5z"/><path d="M4 12l8 4.5 8-4.5"/><path d="M4 16.5L12 21l8-4.5"/>';
  const chevronGlyph = '<path d="M9 6l6 6-6 6"/>';

  const visible = $derived($groups.filter((group) => !group.hidden || $tunnelMode === 'global'));

  function selectedOf(group: { name: string; now?: string }): string {
    return $selectedMap[group.name] || group.now || '';
  }

  function monogramOf(name: string): string {
    const clean = name.trim();
    return clean ? clean.slice(0, 1).toUpperCase() : '·';
  }

  async function openGroup(name: string) {
    // navigate to the proxies page focused on the group (via event for now)
    try {
      window.dispatchEvent(new CustomEvent('goto-proxies', { detail: name }));
    } catch {
      showBanner(name, 'info');
    }
  }
</script>

<DashCard pad={false}>
  <div class="pad">
    <DashHeader glyph={layersGlyph} title={t('proxyGroup')}>
      {#snippet leading()}
        {#if visible.length > 0}
          <DashChip label={String(visible.length)} />
        {/if}
      {/snippet}
      {#snippet trailing()}
        <button class="card-icon" disabled={visible.length === 0} title={t('proxyGroup')}>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html chevronGlyph}</svg>
        </button>
      {/snippet}
    </DashHeader>
  </div>
  {#if visible.length === 0}
    <div class="empty">
      <div class="empty-icon">
        <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html layersGlyph}</svg>
      </div>
      <div class="empty-title">{t('nullTip', { label: t('proxyGroup') })}</div>
      <div class="empty-sub muted">{t('proxyGroupsAfterStart')}</div>
    </div>
  {:else}
    <div class="tiles" style="gap: 8px; padding: 0 14px 14px">
      {#each visible as group (group.name)}
        <button class="tile" onclick={() => openGroup(group.name)}>
          <span class="monogram">{monogramOf(group.name)}</span>
          <span class="titles">
            <span class="gname ellipsis">{group.name}</span>
            <span class="sname ellipsis">{selectedOf(group) || '—'}</span>
          </span>
        </button>
      {/each}
    </div>
  {/if}
</DashCard>

<style>
  .pad {
    padding: 20px 22px 0;
  }

  .card-icon {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--on-surface-variant);
  }

  .card-icon:disabled {
    opacity: 0.4;
  }

  .empty {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 12px 24px 24px;
    text-align: center;
  }

  .empty-icon {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    background: var(--surface-container);
    color: var(--on-surface-variant);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 6px;
  }

  .empty-title {
    font-size: 15px;
    font-weight: 600;
  }

  .empty-sub {
    font-size: 13px;
  }

  .tiles {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(172px, 1fr));
    flex: 1;
    align-content: start;
    overflow-y: auto;
    min-height: 0;
    box-sizing: border-box;
    width: 100%;
  }

  .tile {
    display: flex;
    align-items: center;
    gap: 12px;
    height: 60px;
    background: var(--surface-container);
    border-radius: var(--dash-block-radius);
    padding: 0 14px 0 10px;
    text-align: left;
    cursor: pointer;
    min-width: 0;
  }

  :root.dark .tile {
    background: color-mix(in srgb, var(--surface-container-high) 55%, transparent);
  }

  .tile:hover {
    box-shadow: inset 0 0 0 99px rgba(0, 0, 0, 0.04);
  }

  .monogram {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    background: var(--surface-container-high);
    color: var(--on-surface-variant);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
    font-weight: 600;
    flex-shrink: 0;
  }

  .titles {
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .gname {
    font-size: 15px;
    font-weight: 600;
    line-height: 20px;
  }

  .sname {
    font-size: 12px;
    line-height: 16px;
    color: var(--on-surface-variant);
  }
</style>
