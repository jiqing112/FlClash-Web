<script lang="ts">
  // Tools: entries for hidden pages plus settings subpages.
  import { t, locales } from '../lib/i18n';
  import { config } from '../lib/stores';
  import Scaffold from '../components/Scaffold.svelte';
  import ListRow from '../components/ListRow.svelte';
  import ThemeSettings from './settings/ThemeSettings.svelte';
  import GeneralSettings from './settings/GeneralSettings.svelte';
  import NetworkSettings from './settings/NetworkSettings.svelte';
  import About from './settings/About.svelte';

  let { navigate }: { navigate: (page: string) => void } = $props();

  type SubPage = 'root' | 'theme' | 'general' | 'network' | 'about';
  let sub = $state<SubPage>('root');

  const titles: Record<SubPage, () => string> = {
    root: () => t('tools'),
    theme: () => t('theme'),
    general: () => t('general'),
    network: () => t('network'),
    about: () => t('about'),
  };

  function back() {
    sub = 'root';
  }
</script>

{#if sub === 'root'}
  <Scaffold title={t('tools')}>
    {#snippet children()}
      <div class="tools-page">
        <div class="list-group">
          <ListRow glyph="🕐" title={t('requests')} clickable onclick={() => navigate('requests')}>
            {#snippet trailing()}<span class="chev">›</span>{/snippet}
          </ListRow>
          <div class="divider"></div>
          <ListRow glyph="🌐" title={t('dnsQueries')} clickable onclick={() => navigate('dns')}>
            {#snippet trailing()}<span class="chev">›</span>{/snippet}
          </ListRow>
          <div class="divider"></div>
          <ListRow glyph="🗂" title={t('resources')} clickable onclick={() => navigate('resources')}>
            {#snippet trailing()}<span class="chev">›</span>{/snippet}
          </ListRow>
          <div class="divider"></div>
          <ListRow glyph="📄" title={t('logs')} clickable onclick={() => navigate('logs')}>
            {#snippet trailing()}<span class="chev">›</span>{/snippet}
          </ListRow>
        </div>

        <h2>{t('settings')}</h2>
        <div class="list-group">
          <ListRow glyph="🌐" title={t('language')} value={$config?.appSettingProps?.locale || 'Default'} clickable onclick={() => (sub = 'general')}>
            {#snippet trailing()}<span class="chev">›</span>{/snippet}
          </ListRow>
          <div class="divider"></div>
          <ListRow glyph="🎨" title={t('theme')} clickable onclick={() => (sub = 'theme')}>
            {#snippet trailing()}<span class="chev">›</span>{/snippet}
          </ListRow>
          <div class="divider"></div>
          <ListRow glyph="⚙️" title={t('general')} clickable onclick={() => (sub = 'general')}>
            {#snippet trailing()}<span class="chev">›</span>{/snippet}
          </ListRow>
          <div class="divider"></div>
          <ListRow glyph="📡" title={t('network')} clickable onclick={() => (sub = 'network')}>
            {#snippet trailing()}<span class="chev">›</span>{/snippet}
          </ListRow>
        </div>

        <h2>{t('other')}</h2>
        <div class="list-group">
          <ListRow glyph="ℹ️" title={t('about')} clickable onclick={() => (sub = 'about')}>
            {#snippet trailing()}<span class="chev">›</span>{/snippet}
          </ListRow>
        </div>
      </div>
    {/snippet}
  </Scaffold>
{:else if sub === 'theme'}
  <ThemeSettings title={titles.theme()} onback={back} />
{:else if sub === 'general'}
  <GeneralSettings title={titles.general()} onback={back} />
{:else if sub === 'network'}
  <NetworkSettings title={titles.network()} onback={back} />
{:else}
  <About title={titles.about()} onback={back} />
{/if}

<style>
  .tools-page {
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

  .chev {
    color: var(--on-surface-variant);
    font-size: 1.2rem;
    line-height: 1;
  }
</style>
