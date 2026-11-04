<script lang="ts">
  // SpeedCard per traffic.dart: chart bleeding off the card bottom (88px),
  // display-size figure, peak right, total under it.
  import { t } from '../../lib/i18n';
  import { trafficHistory, totalTraffic } from '../../lib/stores';
  import { trafficText } from '../../lib/format';
  import LineChart from '../../components/LineChart.svelte';
  import DashCard from '../../components/dash/DashCard.svelte';

  let {
    direction,
    mobile = false,
  }: {
    direction: 'upload' | 'download';
    mobile?: boolean;
  } = $props();

  const isUpload = $derived(direction === 'upload');
  const values = $derived(($trafficHistory.map((point) => (isUpload ? point.up : point.down))));
  const current = $derived(values.length ? values[values.length - 1] : 0);
  const peak = $derived(values.reduce((acc, value) => Math.max(acc, value), 0));
  const total = $derived(isUpload ? $totalTraffic.up : $totalTraffic.down);
  const arrowGlyph = $derived(
    isUpload
      ? '<path d="M12 19V5M6 11l6-6 6 6"/>'
      : '<path d="M12 5v14M6 13l6 6 6-6"/>',
  );
</script>

{#if mobile}
  <div class="tile">
    <div class="tile-head">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke={isUpload ? 'var(--on-surface-variant)' : 'var(--primary)'} stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html arrowGlyph}</svg>
      <span class="tile-label">{isUpload ? t('upload') : t('download')}</span>
    </div>
    <div class="figure-row">
      <span class="figure tabular">{trafficText(current)}</span>
      <span class="unit">{trafficText(current).split(' ')[1]}/s</span>
    </div>
    <div class="total">{t('trafficTotal', { value: trafficText(total) })}</div>
  </div>
{:else}
  <DashCard pad={false} clip>
    <div class="chart">
      <LineChart
        values={values}
        color={isUpload ? 'color-mix(in srgb, var(--primary) 45%, var(--surface-bright))' : 'var(--primary)'}
        height={84}
        strokeWidth={2.5}
      />
    </div>
    <div class="content">
      <div class="head">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={isUpload ? 'var(--on-surface-variant)' : 'var(--primary)'} stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html arrowGlyph}</svg>
        <span class="title">{isUpload ? t('upload') : t('download')}</span>
        <span class="peak ellipsis">{t('trafficPeak', { value: trafficText(peak) + '/s' })}</span>
      </div>
      <div class="figure-row">
        <span class="figure tabular">{trafficText(current).split(' ')[0]}</span>
        <span class="unit">{trafficText(current).split(' ')[1] ?? 'B'}/s</span>
      </div>
      <div class="total">{t('trafficTotal', { value: trafficText(total) })}</div>
    </div>
  </DashCard>
{/if}

<style>
  .chart {
    position: absolute;
    left: -3px;
    right: -3px;
    bottom: -3px;
    height: 84px;
  }

  .content {
    position: relative;
    z-index: 1;
    padding: 20px 22px;
    display: flex;
    flex-direction: column;
    height: 100%;
    box-sizing: border-box;
  }

  .head {
    height: 24px;
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .title {
    font-size: 13px;
    font-weight: 600;
    color: var(--on-surface-variant);
  }

  .peak {
    flex: 1;
    text-align: right;
    font-size: 12px;
    color: var(--on-surface-variant);
    font-variant-numeric: tabular-nums;
  }

  .figure-row {
    margin-top: 10px;
    display: flex;
    align-items: baseline;
    gap: 4px;
  }

  .figure {
    font-size: 40px;
    line-height: 44px;
    font-weight: 700;
    letter-spacing: -0.02em;
    white-space: nowrap;
  }

  .unit {
    font-size: 15px;
    font-weight: 600;
    color: var(--on-surface-variant);
    white-space: nowrap;
  }

  .total {
    margin-top: 2px;
    font-size: 12px;
    color: var(--on-surface-variant);
    font-variant-numeric: tabular-nums;
  }

  /* mobile tile */
  .tile {
    background: var(--surface-container);
    border-radius: 22px;
    padding: 14px 16px 12px;
    min-height: 150px;
    display: flex;
    flex-direction: column;
    position: relative;
    overflow: hidden;
    width: 100%;
  }

  :root.dark .tile {
    background: color-mix(in srgb, var(--surface-container-high) 55%, transparent);
  }

  .tile-head {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 13px;
    font-weight: 600;
    color: var(--on-surface-variant);
  }

  .tile-label {
    font-size: 13px;
    font-weight: 600;
    color: var(--on-surface-variant);
  }

  .figure-row .figure {
    font-size: 28px;
    font-weight: 700;
  }
</style>
