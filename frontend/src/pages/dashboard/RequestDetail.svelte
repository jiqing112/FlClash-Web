<script lang="ts">
  // Request detail popup with metadata rows.
  import type { RequestView } from '../../lib/types';
  import { trafficText } from '../../lib/format';
  import { t } from '../../lib/i18n';

  let { view, onclose }: { view: RequestView; onclose: () => void } = $props();

  const rows = $derived([
    [t('network'), view.metadata.network],
    [t('target'), `${view.metadata.host || view.metadata.destinationIP}:${view.metadata.destinationPort}`],
    [t('rule'), `${view.rule} ${view.rulePayload}`],
    [t('proxies'), view.chains?.join(' → ') ?? ''],
    [t('sourceAddress'), `${view.metadata.sourceIP}:${view.metadata.sourcePort}`],
    [t('process'), view.metadata.process],
    [t('download'), trafficText(view.download)],
    [t('upload'), trafficText(view.upload)],
    [t('time'), view.start],
  ]);
</script>

<div class="scrim" role="presentation"></div>
<div class="layer">
  <div class="detail popup-surface rise-in">
    <header>
      <h3>{t('request')}</h3>
      <button class="icon-button small" onclick={onclose}>
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"/></svg>
      </button>
    </header>
    <div class="rows">
      {#each rows as row}
        <div class="row">
          <span class="key muted">{row[0]}</span>
          <span class="val ellipsis">{row[1] || '--'}</span>
        </div>
      {/each}
    </div>
  </div>
</div>

<style>
  .layer {
    position: fixed;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .detail {
    width: 480px;
    max-width: calc(100vw - 48px);
    max-height: 70vh;
    overflow-y: auto;
    padding: 20px;
  }

  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  header h3 {
    margin: 0;
  }

  .rows {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .row {
    display: flex;
    gap: 16px;
    font-size: 0.9rem;
  }

  .key {
    width: 110px;
    flex-shrink: 0;
  }

  .val {
    word-break: break-all;
  }
</style>
