<script lang="ts">
  // ConnectionStatsCard (summary): three connected-pill stats per feed.
  import { t } from '../../lib/i18n';
  import { liveCount, recentRequests, dnsQueries } from '../../lib/stores';

  let { mobile = false }: { mobile?: boolean } = $props();

  const pillGlyph = '<circle cx="12" cy="12" r="8.8"/><path d="M12 7.2V12l3.2 2"/>';
</script>

<div class="dash-card">
  <div class="stats">
    <button class="stat selected">
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html pillGlyph}</svg>
      <span class="figure tabular">{$liveCount}</span>
      <span class="label">{t('liveConnections')}</span>
    </button>
    <button class="stat">
      <span class="figure tabular">{$recentRequests.length}</span>
      <span class="label">{t('recentRequests')}</span>
    </button>
    <button class="stat">
      <span class="figure tabular">{$dnsQueries.length}</span>
      <span class="label">{t('dnsQueries')}</span>
    </button>
  </div>
</div>

<style>
  .dash-card {
    background: var(--surface-bright);
    border-radius: var(--dash-radius);
    box-shadow: var(--shadow-e1);
    height: 100%;
    display: flex;
    padding: 12px;
  }

  :root.dark .dash-card {
    background: var(--surface-container);
  }

  .stats {
    display: flex;
    gap: 3px;
    flex: 1;
  }

  .stat {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    background: var(--surface-container);
    color: var(--on-surface-variant);
    border-radius: var(--dash-radius);
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
    min-width: 0;
  }

  .stat:first-child {
    border-top-left-radius: var(--dash-radius);
  }

  .stat.selected {
    background: var(--primary);
    color: var(--on-primary);
  }

  .figure {
    font-size: 17px;
    font-weight: 700;
  }

  .label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
