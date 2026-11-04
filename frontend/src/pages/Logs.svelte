<script lang="ts">
  // Logs list with level filter.
  import { t } from '../lib/i18n';
  import Scaffold from '../components/Scaffold.svelte';
  import { logs } from '../lib/stores';

  let search = $state('');
  let level = $state('');

  const levelOrder: Record<string, number> = { DEBUG: 0, INFO: 1, WARNING: 2, ERROR: 3 };

  const rows = $derived.by(() => {
    let list = [...$logs].reverse();
    if (level) {
      list = list.filter((row) => (levelOrder[row.LogLevel] ?? 0) >= (levelOrder[level] ?? 0));
    }
    if (search.trim()) {
      const needle = search.trim().toLowerCase();
      list = list.filter((row) => row.Payload.toLowerCase().includes(needle));
    }
    return list.slice(0, 600);
  });

  function tone(levelName: string): string {
    switch (levelName) {
      case 'ERROR':
        return 'var(--error)';
      case 'WARNING':
        return 'var(--warning)';
      default:
        return 'var(--on-surface-variant)';
    }
  }
</script>

<Scaffold title={t('logs')} bind:searchText={search} search searchPlaceholder={t('search')}>
  {#snippet children()}
    <div class="logs-page">
      <div class="level-pills connected-pills">
        <button class:selected={level === ''} onclick={() => (level = '')}>All</button>
        <button class:selected={level === 'INFO'} onclick={() => (level = 'INFO')}>Info+</button>
        <button class:selected={level === 'WARNING'} onclick={() => (level = 'WARNING')}>Warn+</button>
        <button class:selected={level === 'ERROR'} onclick={() => (level = 'ERROR')}>Error</button>
      </div>
      <div class="log-list">
        {#each rows as row, index (index)}
          <div class="log-row">
            <span class="dot" style="background: {tone(row.LogLevel)}"></span>
            <span class="payload ellipsis">{row.Payload}</span>
          </div>
        {:else}
          <div class="empty muted">--</div>
        {/each}
      </div>
    </div>
  {/snippet}
</Scaffold>

<style>
  .logs-page {
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 4px 26px 26px;
    box-sizing: border-box;
    min-height: 0;
  }

  .log-list {
    flex: 1;
    overflow-y: auto;
    background: color-mix(in srgb, var(--on-surface) 3%, transparent);
    border-radius: var(--radius-xl);
    padding: 10px 14px;
    font-family: var(--font-mono);
    font-size: 0.8rem;
    line-height: 1.7;
  }

  .log-row {
    display: flex;
    align-items: baseline;
    gap: 8px;
  }

  .dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    flex-shrink: 0;
    transform: translateY(-1px);
  }

  .payload {
    word-break: break-all;
    white-space: normal;
  }

  .empty {
    text-align: center;
    padding: 40px 0;
  }
</style>
