<script lang="ts">
  // MemoryInfoCard: big figure, segmented bar, 应用/内核 legend.
  import { t } from '../../lib/i18n';
  import { api } from '../../lib/api';
  import { trafficText } from '../../lib/format';
  import DashCard from '../../components/dash/DashCard.svelte';
  import DashHeader from '../../components/dash/DashHeader.svelte';

  let { mobile = false }: { mobile?: boolean } = $props();

  const chipGlyph = '<rect x="6" y="6" width="12" height="12" rx="2"/><path d="M9 2.5V6M15 2.5V6M9 18v3.5M15 18v3.5M2.5 9H6M2.5 15H6M18 9h3.5M18 15h3.5"/>';
  const chevronGlyph = '<path d="M9 6l6 6-6 6"/>';

  let stats = $state<{ rss: number; heapInuse: number; heapIdle: number; stackInuse: number; runtimeOther: number } | null>(null);

  $effect(() => {
    let alive = true;
    const load = async () => {
      try {
        const data = await api.coreMemory();
        if (alive) stats = data;
      } catch {
        // core busy
      }
    };
    load();
    const timer = setInterval(load, 4000);
    return () => {
      alive = false;
      clearInterval(timer);
    };
  });

  const core = $derived(stats ? stats.rss : 0);
  const app = $derived(stats ? Math.max(0, stats.heapInuse + stats.stackInuse + stats.runtimeOther - 0) : 0);
  const total = $derived(Math.max(1, core + app));
  const coreShare = $derived(core / total);
</script>

<DashCard>
  <DashHeader glyph={chipGlyph} title={t('memoryInfo')} chevron={!mobile} />
  <div class="spacer"></div>
  <div class="figure-row">
    <span class="figure tabular">{trafficText(core).split(' ')[0]}</span>
    <span class="unit">{trafficText(core).split(' ')[1] ?? 'MB'}</span>
  </div>
  <div class="bar">
    <div class="seg" style="flex: {coreShare * 1000}; background: var(--primary)"></div>
    <div class="seg" style="flex: {(1 - coreShare) * 1000}; background: color-mix(in srgb, var(--primary) 40%, var(--surface-container-high))"></div>
  </div>
  <div class="legend">
    <div class="legend-row"><span class="dot" style="background: var(--primary)"></span>{t('app')} <b class="tabular">{trafficText(app)}</b></div>
    <div class="legend-row"><span class="dot" style="background: color-mix(in srgb, var(--primary) 40%, var(--surface-container-high))"></span>{t('core')} <b class="tabular">{trafficText(core)}</b></div>
  </div>
</DashCard>

<style>
  .spacer {
    flex: 1;
    min-height: 8px;
  }

  .figure-row {
    display: flex;
    align-items: baseline;
    gap: 4px;
  }

  .figure {
    font-size: 34px;
    line-height: 41px;
    font-weight: 700;
    letter-spacing: -0.02em;
  }

  .unit {
    font-size: 15px;
    font-weight: 600;
    color: var(--on-surface-variant);
  }

  .bar {
    display: flex;
    gap: 2px;
    height: 8px;
    border-radius: 4px;
    overflow: hidden;
    margin-top: 8px;
  }

  .seg {
    height: 100%;
  }

  .legend {
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 13px;
    color: var(--on-surface-variant);
    margin-top: 10px;
  }

  .legend-row {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .legend-row b {
    color: var(--on-surface);
    font-weight: 600;
  }

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }
</style>
