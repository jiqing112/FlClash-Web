<script lang="ts">
  // Requests: live connections and recent requests tables.
  import { t } from '../lib/i18n';
  import Scaffold from '../components/Scaffold.svelte';
  import { liveRequests, recentRequests, liveCount, showBanner } from '../lib/stores';
  import { api } from '../lib/api';
  import { trafficText } from '../lib/format';
  import RequestDetail from './dashboard/RequestDetail.svelte';
  import type { RequestView } from '../lib/types';

  type Feed = 'live' | 'recent';
  let feed = $state<Feed>('live');
  let search = $state('');
  let detail = $state<RequestView | null>(null);

  const rows = $derived.by(() => {
    const source = feed === 'live' ? $liveRequests : $recentRequests;
    if (!search.trim()) return source;
    const needle = search.trim().toLowerCase();
    return source.filter(
      (row) =>
        row.metadata.host.toLowerCase().includes(needle) ||
        row.metadata.process.toLowerCase().includes(needle) ||
        row.metadata.destinationIP.includes(needle) ||
        row.rule.toLowerCase().includes(needle),
    );
  });

  async function closeOne(id: string) {
    try {
      await api.closeRequest(id);
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    }
  }

  async function closeAll() {
    try {
      await api.closeAllRequests();
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    }
  }

  function destinationOf(row: RequestView): string {
    const host = row.metadata.host || row.metadata.destinationIP;
    return `${host}:${row.metadata.destinationPort}`;
  }
</script>

<Scaffold
  title={t('requests')}
  bind:searchText={search}
  search
  searchPlaceholder={t('search')}
  actions={[
    {
      icon: '<path d="M6 6l12 12M18 6L6 18"/>',
      label: t('closeAllConnections'),
      danger: true,
      show: feed === 'live',
      onClick: closeAll,
    },
  ]}
>
  {#snippet children()}
    <div class="requests">
      <div class="connected-pills">
        <button class:selected={feed === 'live'} onclick={() => (feed = 'live')}>
          <span class="chip count-chip tonal">{$liveCount}</span>
          {t('running')}
        </button>
        <button class:selected={feed === 'recent'} onclick={() => (feed = 'recent')}>
          <span class="chip count-chip">{$recentRequests.length}</span>
          {t('recentRequests')}
        </button>
      </div>
      <div class="table-scroll">
        <table class="record-table">
          <thead>
            <tr>
              {#if feed === 'recent'}<th>{t('time')}</th>{/if}
              <th>{t('process')}</th>
              <th>{t('target')}</th>
              <th>{t('rule')}</th>
              <th>{t('exit')}</th>
              {#if feed === 'live'}
                <th>{t('download')} /s</th>
                <th>{t('upload')} /s</th>
              {/if}
              <th>{t('totalDownload')}</th>
              <th>{t('totalUpload')}</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {#each rows as row (row.id)}
              <tr onclick={() => (detail = row)}>
                {#if feed === 'recent'}
                  <td class="mono muted">{new Date(row.start).toLocaleTimeString()}</td>
                {/if}
                <td class="ellipsis">{row.metadata.process || '--'}</td>
                <td class="ellipsis mono">{destinationOf(row)}</td>
                <td class="ellipsis">{row.rule} {row.rulePayload}</td>
                <td class="ellipsis">{row.chains?.[0] ?? '--'}</td>
                {#if feed === 'live'}
                  <td class="mono">{trafficText(row.downloadSpeed ?? 0)}/s</td>
                  <td class="mono">{trafficText(row.uploadSpeed ?? 0)}/s</td>
                {/if}
                <td class="mono">{trafficText(row.download)}</td>
                <td class="mono">{trafficText(row.upload)}</td>
                <td>
                  {#if feed === 'live'}
                    <button class="icon-button small" onclick={(event) => { event.stopPropagation(); closeOne(row.id); }} title={t('closeConnection')}>
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"/></svg>
                    </button>
                  {/if}
                </td>
              </tr>
            {:else}
              <tr><td colspan="9" class="empty muted">--</td></tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/snippet}
</Scaffold>

{#if detail}
  <RequestDetail view={detail} onclose={() => (detail = null)} />
{/if}

<style>
  .requests {
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 4px 26px 26px;
    box-sizing: border-box;
    min-height: 0;
  }

  .count-chip {
    height: 22px;
    padding: 0 8px;
    font-size: 0.76rem;
  }

  .table-scroll {
    flex: 1;
    overflow: auto;
    background: color-mix(in srgb, var(--on-surface) 3%, transparent);
    border-radius: var(--radius-xl);
  }

  tbody tr {
    cursor: pointer;
  }

  tbody tr:hover {
    background: color-mix(in srgb, var(--on-surface) 5%, transparent);
  }

  .empty {
    text-align: center;
    padding: 40px 0;
  }
</style>
