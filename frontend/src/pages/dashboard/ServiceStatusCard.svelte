<script lang="ts">
  // ServiceStatusCard: the probe service grid with brand logos, delay + dot.
  import { t } from '../../lib/i18n';
  import { api } from '../../lib/api';
  import { showBanner } from '../../lib/stores';
  import DashCard from '../../components/dash/DashCard.svelte';
  import DashHeader from '../../components/dash/DashHeader.svelte';

  let { mobile = false }: { mobile?: boolean } = $props();

  const shieldGlyph = '<path d="M12 3l7 3v5c0 5-3.2 8.4-7 10-3.8-1.6-7-5-7-10V6z"/>';
  const chevronGlyph = '<path d="M9 6l6 6-6 6"/>';
  const syncGlyph = '<path d="M20 12a8 8 0 1 1-2.3-5.7"/><path d="M20 3v4h-4"/>';

  type Service = { id: string; label: string; svg: string };

  const services: Service[] = [
    { id: 'google', label: 'Google', svg: `<svg viewBox="0 0 24 24"><path fill="#EA4335" d="M12 5.3c1.7 0 3.2.6 4.4 1.7l3.2-3.2C17.6 1.9 15 .9 12 .9 7.4.9 3.4 3.6 1.5 7.5l3.8 2.9C6.2 7.5 8.9 5.3 12 5.3z"/><path fill="#4285F4" d="M23.1 12.3c0-.9-.1-1.6-.2-2.3H12v4.5h6.3c-.3 1.5-1.1 2.7-2.4 3.6l3.7 2.9c2.2-2 3.5-5 3.5-8.7z"/><path fill="#FBBC05" d="M5.3 14.4c-.3-.8-.4-1.6-.4-2.4s.2-1.6.4-2.4L1.5 6.7C.7 8.3.3 10.1.3 12s.4 3.7 1.2 5.3l3.8-2.9z"/><path fill="#34A853" d="M12 23.1c3 0 5.6-1 7.5-2.8l-3.7-2.9c-1 .7-2.3 1.1-3.8 1.1-3.1 0-5.8-2.2-6.7-5.1l-3.8 2.9c1.9 3.9 5.9 6.8 10.5 6.8z"/></svg>` },
    { id: 'github', label: 'GitHub', svg: `<svg viewBox="0 0 24 24"><path fill="currentColor" d="M12 .8C5.7.8.6 5.9.6 12.2c0 5 3.3 9.3 7.8 10.8.6.1.8-.2.8-.5v-2c-3.2.7-3.8-1.3-3.8-1.3-.5-1.3-1.3-1.7-1.3-1.7-1-.7.1-.7.1-.7 1.2.1 1.8 1.2 1.8 1.2 1 1.8 2.8 1.3 3.4 1 .1-.8.4-1.3.7-1.6-2.6-.3-5.3-1.3-5.3-5.7 0-1.3.4-2.3 1.2-3.1-.1-.3-.5-1.5.1-3.1 0 0 1-.3 3.2 1.2a11 11 0 0 1 5.8 0c2.2-1.5 3.2-1.2 3.2-1.2.6 1.6.2 2.8.1 3.1.7.8 1.2 1.8 1.2 3.1 0 4.4-2.7 5.4-5.3 5.7.4.4.8 1.1.8 2.2v3.2c0 .3.2.6.8.5 4.6-1.5 7.8-5.8 7.8-10.8C23.4 5.9 18.3.8 12 .8z"/></svg>` },
    { id: 'youtube', label: 'YouTube', svg: `<svg viewBox="0 0 24 24"><path fill="#FF0000" d="M23.2 6.6c-.3-1-1-1.8-2-2C19.3 4.1 12 4.1 12 4.1s-7.3 0-9.2.5c-1 .3-1.8 1-2 2C.3 8.5.3 12 .3 12s0 3.5.5 5.4c.3 1 1 1.8 2 2 1.9.5 9.2.5 9.2.5s7.3 0 9.2-.5c1-.3 1.8-1 2-2 .5-1.9.5-5.4.5-5.4s0-3.5-.5-5.4zM9.7 15.3V8.7L15.8 12l-6.1 3.3z"/></svg>` },
    { id: 'telegram', label: 'Telegram', svg: `<svg viewBox="0 0 24 24"><path fill="currentColor" d="M22.6 2.7L1.9 10.7c-1.4.5-1.4 1.4-.2 1.7l5.3 1.7 2 6.2c.2.7.4.9 1 .9.5 0 .7-.2 1.2-.6l2.7-2.6 5.4 4c1 .5 1.7.3 2-.9l3.7-17.3c.4-1.5-.5-2.2-1.4-1.7zM7.9 13.7l11.7-7.4c.6-.3 1.1-.2.7.3l-10 9-.4 3.8-2-5.7z"/></svg>` },
    { id: 'apple', label: 'Apple', svg: `<svg viewBox="0 0 24 24"><path fill="currentColor" d="M17.5 12.6c0-2.5 2-3.7 2.1-3.7-1.2-1.7-3-1.9-3.6-1.9-1.5-.2-3 .9-3.8.9-.8 0-2-.9-3.3-.9-1.7 0-3.2 1-4.1 2.5-1.7 3-.4 7.5 1.3 10 .8 1.2 1.8 2.6 3.1 2.5 1.2-.1 1.7-.8 3.2-.8s1.9.8 3.3.8c1.4 0 2.2-1.2 3-2.5.9-1.4 1.3-2.8 1.3-2.8s-2.5-1-2.5-4.1zM15 5.4c.7-.8 1.1-2 1-3.1-1 0-2.2.7-2.9 1.5-.6.7-1.2 1.9-1 3 1.1.1 2.2-.6 2.9-1.4z"/></svg>` },
    { id: 'cloudflare', label: 'Cloudflare', svg: `<svg viewBox="0 0 24 24"><path fill="currentColor" d="M16.5 15.8H7.2c-.2 0-.4-.2-.3-.5l.5-1.6c.1-.3.4-.5.7-.5h8.1c.4 0 .8-.3.9-.7l.9-3.2c0-.2-.1-.4-.3-.4H4.8c-.2 0-.4-.2-.3-.5l.3-1c.1-.2.3-.4.5-.4h13.9c.4 0 .8-.3.9-.7l.6-2c0-.2-.1-.4-.3-.4H6.2C5.6 4 5.1 3.6 5 3c-.1-.9.5-1.6 1.3-1.7.8-.1 1.5.4 1.7 1.2.1.4 0 .8-.2 1.1h4.5c-.1-.3-.1-.7 0-1.1.2-.7.8-1.2 1.6-1.2.9 0 1.6.7 1.7 1.5 0 .3 0 .6-.1.8h4.6c.3 0 .5.2.4.5l-.5 1.9c-.1.3-.4.5-.7.5h1.1c.3 0 .5.2.4.5l-1.6 6c-.1.3-.4.5-.7.5h-.9l-1 3.4z"/></svg>` },
    { id: 'discord', label: 'Discord', svg: `<svg viewBox="0 0 24 24"><path fill="currentColor" d="M20.3 4.4A19.8 19.8 0 0 0 15.4 3c-.2.4-.5 1-.6 1.4a18.3 18.3 0 0 0-5.5 0C9.1 4 8.8 3.4 8.6 3c-1.7.3-3.4.8-4.9 1.5A20.3 20.3 0 0 0 .3 18.1a20 20 0 0 0 6 3c.5-.6.9-1.3 1.3-2-.7-.3-1.4-.6-2-1l.5-.4a14.2 14.2 0 0 0 12 0l.5.4c-.7.4-1.3.7-2 1 .4.7.8 1.4 1.3 2a20 20 0 0 0 6-3c.5-4.6-.7-8.6-3.1-13.7zM8 15.3c-1.2 0-2.1-1-2.1-2.3s.9-2.3 2.1-2.3 2.2 1 2.1 2.3c0 1.3-.9 2.3-2.1 2.3zm8 0c-1.2 0-2.1-1-2.1-2.3s.9-2.3 2.1-2.3 2.2 1 2.1 2.3c0 1.3-.9 2.3-2.1 2.3z"/></svg>` },
    { id: 'openai', label: 'ChatGPT', svg: `<svg viewBox="0 0 24 24"><path fill="currentColor" d="M22.3 9.8a5.9 5.9 0 0 0-.5-4.9 6 6 0 0 0-6.5-2.9A5.9 5.9 0 0 0 10.9.4 6 6 0 0 0 4.5 5.3a5.9 5.9 0 0 0-4 2.9 6 6 0 0 0 .7 7.1 5.9 5.9 0 0 0 .5 4.9 6 6 0 0 0 6.5 2.9 5.9 5.9 0 0 0 4.4 1.6 6 6 0 0 0 5.7-4.2 5.9 5.9 0 0 0 4-2.9 6 6 0 0 0-.7-7.1l-.3-.7zM12 21.1a4.4 4.4 0 0 1-2.9-1l.1-.1 4.8-2.8a.8.8 0 0 0 .4-.7V9.9l2 1.2v5.4a4.5 4.5 0 0 1-4.4 4.6zm-9.5-4.2a4.4 4.4 0 0 1-.5-3l.1.1 4.8 2.8a.8.8 0 0 0 .8 0l5.5-3.2v2.4l-4.7 2.7a4.5 4.5 0 0 1-6-.8zm-1.2-9.6a4.4 4.4 0 0 1 2.3-2l-.1 5.4a.8.8 0 0 0 .4.7l5.5 3.2-2 1.2-4.7-2.8a4.5 4.5 0 0 1-1.7-5.7zm16.6 4.6-5.5-3.2 2-1.2 4.7 2.7a4.5 4.5 0 0 1-.7 8.1v-5.5a.8.8 0 0 0-.5-.9zm2-3.2-.1-.1-4.8-2.9a.8.8 0 0 0-.8 0L9 8.8V6.4l4.7-2.7a4.5 4.5 0 0 1 6.7 4.7zM8.2 13.1l-2-1.2V6.6a4.5 4.5 0 0 1 7.4-3.5l-.1.1-4.8 2.8a.8.8 0 0 0-.4.7v6.4zm1.1-2.4 2.3-1.3 2.3 1.3v2.6l-2.3 1.3-2.3-1.3z"/></svg>` },
  ];

  let results = $state<Record<string, { delay: number; error?: string; status?: string }>>({});
  let checking = $state(false);

  async function checkAll() {
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
    checkAll();
    const timer = setInterval(checkAll, 60_000);
    return () => clearInterval(timer);
  });

  const abnormal = $derived(
    services.filter((service) => results[service.id] && results[service.id].status !== 'available').length,
  );

  function delayTextOf(id: string): string {
    const result = results[id];
    if (!result) return '';
    if (result.status === 'available' && result.delay > 0) return `${result.delay} ms`;
    if (result.status === 'available') return t('serviceAvailable');
    if (result.error) return t('serviceFailed');
    return t('servicePending');
  }
</script>

<DashCard>
  <DashHeader glyph={shieldGlyph} title={t('serviceStatus')} chevron={!mobile}>
    {#snippet leading()}
      {#if abnormal > 0}
        <span class="chip warn">{t('serviceAbnormalCount', { count: abnormal })}</span>
      {:else if Object.keys(results).length}
        <span class="chip ok">{t('serviceAllAvailable')}</span>
      {/if}
    {/snippet}
    {#snippet trailing()}
      <button class="card-icon" class:spinning={checking} onclick={checkAll} title={t('retry')}>
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html syncGlyph}</svg>
      </button>
    {/snippet}
  </DashHeader>
  <div class="spacer"></div>
  <div class="grid" style="grid-template-columns: repeat({mobile ? 1 : 2}, minmax(0, 1fr))">
    {#each services as service (service.id)}
      <div class="cell">
        <span class="badge">
          <span class="logo">{@html service.svg}</span>
          <span
            class="dot"
            class:ok={results[service.id]?.status === 'available'}
            class:bad={results[service.id] && results[service.id].status !== 'available'}
          ></span>
        </span>
        <span class="meta">
          <span class="name ellipsis">{service.label}</span>
          <span class="delay ellipsis tabular">
            {#if results[service.id]}
              {delayTextOf(service.id)}
            {:else}
              <span class="skeleton" style="width: 36px; height: 10px; display: inline-block"></span>
            {/if}
          </span>
        </span>
      </div>
    {/each}
  </div>
</DashCard>

<style>
  .card-icon {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--on-surface-variant);
    flex-shrink: 0;
  }

  .card-icon:hover {
    background: color-mix(in srgb, var(--on-surface) 7%, transparent);
  }

  .card-icon.spinning svg {
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .chip {
    height: 24px;
    padding: 0 10px;
    border-radius: var(--radius-full);
    font-size: 12px;
    font-weight: 600;
    display: inline-flex;
    align-items: center;
    white-space: nowrap;
  }

  .chip.ok {
    background: var(--success-container);
    color: var(--on-success-container);
  }

  .chip.warn {
    background: var(--warning-container);
    color: var(--on-warning-container);
  }

  .spacer {
    flex: 1;
    min-height: 10px;
  }

  .grid {
    display: grid;
    gap: 6px;
  }

  .cell {
    display: flex;
    align-items: center;
    gap: 9px;
    background: var(--surface-container);
    border-radius: var(--dash-tab-radius);
    padding: 4px 7px;
    min-width: 0;
    min-height: 40px;
  }

  :root.dark .cell {
    background: color-mix(in srgb, var(--surface-container-high) 55%, transparent);
  }

  .badge {
    position: relative;
    width: 30px;
    height: 30px;
    border-radius: 8px;
    background: var(--surface-bright);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  :root.dark .badge {
    background: var(--surface-container-high);
  }

  .logo {
    width: 17px;
    height: 17px;
    display: inline-flex;
  }

  .dot {
    position: absolute;
    left: 2px;
    bottom: 2px;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    border: 1.5px solid var(--surface-bright);
    background: var(--on-surface-variant);
  }

  :root.dark .dot {
    border-color: var(--surface-container-high);
  }

  .dot.ok {
    background: var(--success);
  }

  .dot.bad {
    background: var(--error);
  }

  .meta {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 1px;
  }

  .name {
    font-size: 13.5px;
    line-height: 16px;
    font-weight: 600;
    min-width: 0;
  }

  .delay {
    font-size: 11px;
    line-height: 13px;
    color: var(--on-surface-variant);
    font-variant-numeric: tabular-nums;
  }
</style>
