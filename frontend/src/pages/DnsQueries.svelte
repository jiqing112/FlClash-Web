<script lang="ts">
  // DNS queries table.
  import { t } from '../lib/i18n';
  import Scaffold from '../components/Scaffold.svelte';
  import { dnsQueries, showBanner } from '../lib/stores';
  import { api } from '../lib/api';
  import { timeText } from '../lib/format';

  let search = $state('');

  const rows = $derived.by(() => {
    const list = [...$dnsQueries].reverse();
    if (!search.trim()) return list;
    const needle = search.trim().toLowerCase();
    return list.filter((row) => row.domain.toLowerCase().includes(needle));
  });

  async function clearCache() {
    try {
      await api.clearDnsCache();
      dnsQueries.set([]);
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    }
  }
</script>

<Scaffold
  title={t('dnsQueries')}
  bind:searchText={search}
  search
  searchPlaceholder={t('search')}
  actions={[
    { icon: '<path d="M4 12a8 8 0 1 1 8 8"/><path d="M4 20v-5h5"/>', label: t('clearDnsCache'), onClick: clearCache },
  ]}
>
  {#snippet children()}
    <div class="dns-page">
      <div class="table-scroll">
        <table class="record-table">
          <thead>
            <tr>
              <th>{t('time')}</th>
              <th>{t('domain')}</th>
              <th>{t('type')}</th>
              <th>{t('initiator')}</th>
              <th>{t('result')}</th>
              <th>{t('source')}</th>
              <th>{t('ms')}</th>
            </tr>
          </thead>
          <tbody>
            {#each rows as row, index (index)}
              <tr>
                <td class="mono muted">{timeText(row.time)}</td>
                <td class="mono">{row.domain}</td>
                <td>{row.type}</td>
                <td class="muted">{row.initiator || '--'}</td>
                <td class="ellipsis mono">{row.answers?.slice(0, 2).join(', ') || row.error || row.rcode || ''}</td>
                <td class="muted ellipsis">{row.upstream || (row.cached ? 'cache' : '')}</td>
                <td class="mono muted">{row.delay}</td>
              </tr>
            {:else}
              <tr><td colspan="7" class="empty muted">{t('noData')}</td></tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/snippet}
</Scaffold>

<style>
  .dns-page {
    height: 100%;
    display: flex;
    flex-direction: column;
    padding: 4px 26px 26px;
    box-sizing: border-box;
    min-height: 0;
  }

  .table-scroll {
    flex: 1;
    overflow: auto;
    background: color-mix(in srgb, var(--on-surface) 3%, transparent);
    border-radius: var(--radius-xl);
  }

  .empty {
    text-align: center;
    padding: 40px 0;
  }
</style>
