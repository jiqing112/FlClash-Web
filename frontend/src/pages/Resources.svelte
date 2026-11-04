<script lang="ts">
  // Resources: geo resources + external proxy providers management.
  import { t } from '../lib/i18n';
  import Scaffold from '../components/Scaffold.svelte';
  import { config, providers, showBanner } from '../lib/stores';
  import { api } from '../lib/api';
  import { dateText, trafficText } from '../lib/format';

  const geoResources = [
    { type: 'geoip', label: 'GeoIP' },
    { type: 'geosite', label: 'GeoSite' },
    { type: 'mmdb', label: 'MMDB' },
    { type: 'asn', label: 'ASN' },
  ];

  let updating = $state<Set<string>>(new Set());

  async function updateGeo(type: string) {
    updating.add(type);
    updating = new Set(updating);
    try {
      await api.geoUpdate(type);
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    } finally {
      updating.delete(type);
      updating = new Set(updating);
    }
  }

  async function updateProvider(name: string) {
    updating.add(name);
    updating = new Set(updating);
    try {
      await api.updateProvider(name);
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    } finally {
      updating.delete(name);
      updating = new Set(updating);
    }
  }

  const providerList = $derived(Object.values($providers));
</script>

<Scaffold title={t('resources')}>
  {#snippet children()}
    <div class="resources-page">
      <section>
        <h2>{t('geoResources')}</h2>
        <div class="list-group">
          {#each geoResources as resource (resource.type)}
            <div class="list-row">
              <div class="titles">
                <div class="title">{resource.label}</div>
                <div class="subtitle">{$config?.patchClashConfig['geox-url']?.[resource.type] ?? ''}</div>
              </div>
              <button class="icon-button small" class:spinning={updating.has(resource.type)} onclick={() => updateGeo(resource.type)} title={t('sync')}>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 12a8 8 0 1 1-2.3-5.7"/><path d="M20 3v4h-4"/></svg>
              </button>
            </div>
            {#if resource !== geoResources[geoResources.length - 1]}<div class="divider"></div>{/if}
          {/each}
        </div>
      </section>

      <section>
        <h2>{t('proxyProviders')}</h2>
        {#if providerList.length === 0}
          <div class="empty muted">{t('noProxyProviders')}</div>
        {:else}
          <div class="list-group">
            {#each providerList as provider (provider.name)}
              <div class="list-row">
                <div class="titles">
                  <div class="title">{provider.name}</div>
                  <div class="subtitle">
                    {provider.count} · {provider['vehicle-type']}
                    {#if provider['update-at']}
                      · {dateText(Date.parse(provider['update-at']))}
                    {/if}
                  </div>
                </div>
                <button class="icon-button small" class:spinning={updating.has(provider.name)} onclick={() => updateProvider(provider.name)} title={t('sync')}>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 12a8 8 0 1 1-2.3-5.7"/><path d="M20 3v4h-4"/></svg>
                </button>
              </div>
            {/each}
          </div>
        {/if}
      </section>
    </div>
  {/snippet}
</Scaffold>

<style>
  .resources-page {
    height: 100%;
    overflow-y: auto;
    padding: 4px 26px 26px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 22px;
    max-width: 900px;
  }

  h2 {
    font-size: 1.1rem;
    margin: 0 0 10px;
  }

  .spinning {
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .empty {
    padding: 24px;
    text-align: center;
    background: color-mix(in srgb, var(--on-surface) 3%, transparent);
    border-radius: var(--radius-lg);
  }
</style>
