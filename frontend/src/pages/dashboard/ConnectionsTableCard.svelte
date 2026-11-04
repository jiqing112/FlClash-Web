<script lang="ts">
  // ConnectionsTableCard per connections_table.dart: feed tabs hugging their
  // content (outer = height/2, inner 8), plain-text counts with grouped
  // digits, and the 来源/目标/出口/下载/上传 table with fading overflow.
  import { t } from '../../lib/i18n';
  import { liveCount, liveRequests, recentRequests, dnsQueries, showBanner } from '../../lib/stores';
  import { api } from '../../lib/api';
  import { trafficText } from '../../lib/format';
  import DashCard from '../../components/dash/DashCard.svelte';
  import ConnectedButton from '../../components/dash/ConnectedButton.svelte';
  import RequestDetail from './RequestDetail.svelte';

  type Feed = 'live' | 'recent' | 'dns';
  let feed = $state<Feed>('live');
  let detail = $state<any | null>(null);

  const tabHeight = 36;
  const tabInset = 32 - tabHeight / 2; // 14: pill center concentric with card corner
  const cellStart = 22 - 8;

  // dashboardCount: grouped digits like the original's NumberFormat('en').
  function groupDigits(count: number): string {
    return count.toLocaleString('en-US');
  }

  const rows = $derived.by(() => {
    if (feed === 'recent') return $recentRequests.slice(0, 100);
    if (feed === 'dns') return [];
    return $liveRequests.slice(0, 100);
  });

  async function closeOne(id: string) {
    try {
      await api.closeRequest(id);
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    }
  }

  function destinationOf(row: any): string {
    const host = row.metadata.host || row.metadata.destinationIP;
    return `${host}:${row.metadata.destinationPort}`;
  }

  function exitOf(row: any): string {
    return row.chains?.[0] ?? '';
  }

  function isBuiltinExit(name: string): boolean {
    return ['DIRECT', 'REJECT', 'REJECT-DROP', 'PASS'].includes(name);
  }

  function letterOf(name: string): string {
    return name ? name.slice(0, 1).toUpperCase() : '?';
  }
</script>

<DashCard pad={false}>
  <div class="inner" style="padding: {tabInset}px {cellStart}px 0">
    <div class="tabs" style="height: {tabHeight}px">
      <ConnectedButton first selected={feed === 'live'} outer={tabHeight / 2} inner={8} onclick={() => (feed = 'live')}>
        <span class="tab-content">
          <span class="count tabular">{groupDigits($liveCount)}</span>
          <span class="label">{t('liveConnections')}</span>
        </span>
      </ConnectedButton>
      <ConnectedButton selected={feed === 'recent'} outer={tabHeight / 2} inner={8} onclick={() => (feed = 'recent')}>
        <span class="tab-content">
          <span class="count tabular">{groupDigits($recentRequests.length)}</span>
          <span class="label">{t('recentRequests')}</span>
        </span>
      </ConnectedButton>
      <ConnectedButton last selected={feed === 'dns'} outer={tabHeight / 2} inner={8} onclick={() => (feed = 'dns')}>
        <span class="tab-content">
          <span class="count tabular">{groupDigits($dnsQueries.length)}</span>
          <span class="label">{t('dnsQueries')}</span>
        </span>
      </ConnectedButton>
    </div>
  </div>
  <div class="table-wrap">
    {#if feed === 'dns'}
      <div class="empty muted">{t('noData')}</div>
    {:else}
      <table class="record-table">
        <thead>
          <tr>
            <th class="col-source">{t('source')}</th>
            <th>{t('target')}</th>
            <th class="col-exit">{t('outboundExit')}</th>
            <th class="col-num">{t('download')}</th>
            <th class="col-num">{t('upload')}</th>
            {#if feed === 'live'}<th class="col-close"></th>{/if}
          </tr>
        </thead>
        <tbody>
          {#each rows as row (row.id)}
            <tr onclick={() => (detail = row)}>
              <td class="col-source">
                <span class="source-cell">
                  <span class="avatar">{letterOf(row.metadata.process || row.metadata.host || '')}</span>
                  <span class="source-name ellipsis">{row.metadata.process || '—'}</span>
                </span>
              </td>
              <td class="ellipsis mono">{destinationOf(row)}</td>
              <td class="col-exit">
                <span class="exit-cell" class:muted-text={isBuiltinExit(exitOf(row))}>{exitOf(row) || '—'}</span>
              </td>
              <td class="col-num tabular">{trafficText(row.download)}</td>
              <td class="col-num tabular">{trafficText(row.upload)}</td>
              {#if feed === 'live'}
                <td class="col-close">
                  <button class="close" onclick={(event) => { event.stopPropagation(); closeOne(row.id); }} title={t('closeConnection')}>
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"/></svg>
                  </button>
                </td>
              {/if}
            </tr>
          {:else}
            <tr><td colspan="6" class="empty muted">{t('noData')}</td></tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</DashCard>

{#if detail}
  <RequestDetail view={detail} onclose={() => (detail = null)} />
{/if}

<style>
  .inner {
    flex-shrink: 0;
  }

  /* natural width like the original's FittedBox(CappedRow(share .45)) */
  .tabs {
    display: inline-flex;
    gap: 3px;
    max-width: 45%;
  }

  .tabs :global(.connected) {
    flex: 0 0 auto;
  }

  .tab-content {
    display: flex;
    align-items: center;
    gap: 6px;
    height: 100%;
    padding: 0 14px;
    white-space: nowrap;
    min-width: 0;
    overflow: hidden;
  }

  /* _TabCount: plain text, subheadline w600 tabular, in the segment color */
  .count {
    font-size: 15px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
  }

  .label {
    font-size: 13px;
    font-weight: 600;
  }

  .table-wrap {
    flex: 1;
    overflow-y: auto;
    margin: 12px 8px 0;
    min-height: 0;
    /* DashboardFadeOut: the last 44px of the list fade out */
    -webkit-mask-image: linear-gradient(to bottom, #000 calc(100% - 44px), transparent);
    mask-image: linear-gradient(to bottom, #000 calc(100% - 44px), transparent);
  }

  .table-wrap::-webkit-scrollbar {
    width: 0;
    height: 0;
  }

  .record-table th {
    background: transparent;
    font-size: 12px;
    padding: 6px 14px;
  }

  .record-table td {
    padding: 7px 14px;
    height: 34px;
    box-sizing: border-box;
  }

  tbody tr {
    cursor: pointer;
  }

  tbody tr:hover {
    background: color-mix(in srgb, var(--on-surface) 5%, transparent);
  }

  .col-source {
    width: 150px;
  }

  .col-exit {
    width: 110px;
  }

  .col-num {
    width: 70px;
    text-align: right;
  }

  .col-close {
    width: 36px;
    text-align: center;
  }

  /* _Source: 22px letter avatar + process name */
  .source-cell {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    max-width: 100%;
  }

  .avatar {
    width: 22px;
    height: 22px;
    border-radius: 7px;
    background: var(--surface-container-high);
    color: var(--on-surface-variant);
    font-size: 11px;
    font-weight: 600;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .source-name {
    font-weight: 600;
    font-size: 13px;
  }

  .exit-cell {
    display: inline-block;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .muted-text {
    color: var(--on-surface-variant);
  }

  .empty {
    text-align: center;
    padding: 40px 0;
  }

  .close {
    width: 26px;
    height: 26px;
    border-radius: 50%;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--on-surface-variant);
  }

  .close:hover {
    background: color-mix(in srgb, var(--error) 12%, transparent);
    color: var(--error);
  }
</style>
