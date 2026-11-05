<script lang="ts">
  // ServiceStatusCard, stacked variant (columns == 1) as the real dashboard
  // shows: header with chevron, then two full-width cells — brand logo tile
  // with a status dot, and the delay right-aligned. No verdict chip and no
  // check-all button in this layout; the chevron opens the full sheet.
  import { t } from '../../lib/i18n';
  import { api } from '../../lib/api';
  import { showBanner } from '../../lib/stores';
  import DashCard from '../../components/dash/DashCard.svelte';
  import DashHeader from '../../components/dash/DashHeader.svelte';

  let { mobile = false }: { mobile?: boolean } = $props();

  const verifiedGlyph = '<path d="M12 3l7 3v5c0 5-3.2 8.4-7 10-3.8-1.6-7-5-7-10V6z"/><path d="M9.2 11.8l2 2 3.6-3.8"/>';
  const chevronGlyph = '<path d="M9 6l6 6-6 6"/>';

  type Service = { id: string; label: string; svg: string };

  // First two of ServiceTarget.values: what the stacked card shows.
  const services: Service[] = [
    { id: 'google', label: 'Google', svg: `<svg viewBox="0 0 24 24"><path fill="#EA4335" d="M12 5.3c1.7 0 3.2.6 4.4 1.7l3.2-3.2C17.6 1.9 15 .9 12 .9 7.4.9 3.4 3.6 1.5 7.5l3.8 2.9C6.2 7.5 8.9 5.3 12 5.3z"/><path fill="#4285F4" d="M23.1 12.3c0-.9-.1-1.6-.2-2.3H12v4.5h6.3c-.3 1.5-1.1 2.7-2.4 3.6l3.7 2.9c2.2-2 3.5-5 3.5-8.7z"/><path fill="#FBBC05" d="M5.3 14.4c-.3-.8-.4-1.6-.4-2.4s.2-1.6.4-2.4L1.5 6.7C.7 8.3.3 10.1.3 12s.4 3.7 1.2 5.3l3.8-2.9z"/><path fill="#34A853" d="M12 23.1c3 0 5.6-1 7.5-2.8l-3.7-2.9c-1 .7-2.3 1.1-3.8 1.1-3.1 0-5.8-2.2-6.7-5.1l-3.8 2.9c1.9 3.9 5.9 6.8 10.5 6.8z"/></svg>` },
    { id: 'github', label: 'GitHub', svg: `<svg viewBox="0 0 24 24"><path fill="currentColor" d="M12 .8C5.7.8.6 5.9.6 12.2c0 5 3.3 9.3 7.8 10.8.6.1.8-.2.8-.5v-2c-3.2.7-3.8-1.3-3.8-1.3-.5-1.3-1.3-1.7-1.3-1.7-1-.7.1-.7.1-.7 1.2.1 1.8 1.2 1.8 1.2 1 1.8 2.8 1.3 3.4 1 .1-.8.4-1.3.7-1.6-2.6-.3-5.3-1.3-5.3-5.7 0-1.3.4-2.3 1.2-3.1-.1-.3-.5-1.5.1-3.1 0 0 1-.3 3.2 1.2a11 11 0 0 1 5.8 0c2.2-1.5 3.2-1.2 3.2-1.2.6 1.6.2 2.8.1 3.1.7.8 1.2 1.8 1.2 3.1 0 4.4-2.7 5.4-5.3 5.7.4.4.8 1.1.8 2.2v3.2c0 .3.2.6.8.5 4.6-1.5 7.8-5.8 7.8-10.8C23.4 5.9 18.3.8 12 .8z"/></svg>` },
  ];

  let results = $state<Record<string, { delay: number; error?: string; status?: string }>>({});
  let checking = $state(false);

  async function probe() {
    checking = true;
    try {
      const raw: any = await api.probeServices(services.map((service) => service.id));
      const next: Record<string, { delay: number; error?: string; status?: string }> = {};
      if (Array.isArray(raw)) {
        for (const item of raw) {
          next[item.name] = { delay: item.delay ?? 0, error: item.error, status: item.status };
        }
      }
      results = next;
    } catch {
      // ignore
    } finally {
      checking = false;
    }
  }

  $effect(() => {
    let retries = 0;
    const run = async () => {
      await probe();
      if (Object.keys(results).length === 0 && retries < 4) {
        retries++;
        setTimeout(run, 3000);
      }
    };
    run();
    const timer = setInterval(probe, 60_000);
    return () => clearInterval(timer);
  });

  function statusTextOf(id: string): { text: string; tone: 'idle' | 'ok' | 'warn' | 'bad' } {
    const result = results[id];
    if (!result) return { text: '', tone: 'idle' };
    if (result.status === 'available' && result.delay > 0) return { text: `${result.delay} ms`, tone: 'idle' };
    if (result.status === 'available') return { text: t('serviceAvailable'), tone: 'ok' };
    if (result.status === 'restricted' || result.status === 'blocked') return { text: t('serviceRestricted'), tone: 'warn' };
    return { text: t('serviceFailed'), tone: 'bad' };
  }

  function dotOf(id: string): string | null {
    const result = results[id];
    if (!result) return null;
    if (result.status === 'available') return 'var(--success)';
    if (result.error || (result.status && result.status !== 'available')) return 'var(--error)';
    return null;
  }
</script>

<DashCard>
  <DashHeader glyph={verifiedGlyph} title={t('serviceStatus')}>
    {#snippet trailing()}
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html chevronGlyph}</svg>
    {/snippet}
  </DashHeader>
  <div class="spacer"></div>
  <div class="cells" style="gap: 8px">
    {#each services as service (service.id)}
      <div class="cell" title={`${service.label} · ${statusTextOf(service.id).text}`}>
        <span class="badge">
          <span class="logo" style="color: var(--on-surface-variant)">{@html service.svg}</span>
          {#if dotOf(service.id)}
            <span class="dot" style="background: {dotOf(service.id)}"></span>
          {/if}
        </span>
        <span class="status" class:ok={statusTextOf(service.id).tone === 'ok'} class:warn={statusTextOf(service.id).tone === 'warn'} class:bad={statusTextOf(service.id).tone === 'bad'}>
          {#if results[service.id]}
            {statusTextOf(service.id).text}
          {:else}
            <span class="skeleton" style="width: 44px; height: 12px; display: inline-block"></span>
          {/if}
        </span>
      </div>
    {/each}
  </div>
</DashCard>

<style>
  .spacer {
    flex: 1;
    min-height: 10px;
  }

  .cells {
    display: flex;
    flex-direction: column;
  }

  /* _ServiceCell stacked: block radius 16, padding 7/12, height 46 */
  .cell {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 46px;
    background: var(--surface-container);
    border-radius: var(--dash-tab-radius);
    padding: 0 12px 0 7px;
    min-width: 0;
  }

  :root.dark .cell {
    background: color-mix(in srgb, var(--surface-container-high) 55%, transparent);
  }

  /* _ServiceBadge: 32px tile, radius 9, card fill, dot at bottom start */
  .badge {
    position: relative;
    width: 32px;
    height: 32px;
    border-radius: 9px;
    background: var(--surface-bright);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  :root.dark .badge {
    background: var(--surface-container);
  }

  .logo {
    width: 18px;
    height: 18px;
    display: inline-flex;
  }

  .dot {
    position: absolute;
    left: 1px;
    bottom: 1px;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    border: 1.5px solid var(--surface-bright);
  }

  :root.dark .dot {
    border-color: var(--surface-container);
  }

  .status {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 12px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    color: var(--on-surface-variant);
    white-space: nowrap;
  }

  .status.ok {
    color: var(--success);
  }

  .status.warn {
    color: var(--warning);
  }

  .status.bad {
    color: var(--error);
  }
</style>
