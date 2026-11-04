<script lang="ts">
  // General settings: language, request behaviour, dashboard, core options.
  import Scaffold from '../../components/Scaffold.svelte';
  import ListRow from '../../components/ListRow.svelte';
  import { config, pushConfig } from '../../lib/stores';
  import { t, locales } from '../../lib/i18n';

  let { title, onback }: { title: string; onback: () => void } = $props();

  const app = $derived($config?.appSettingProps);

  function setLocale(locale: string) {
    pushConfig({ appSettingProps: { locale } });
  }

  function setNumber(key: 'delayTestTimeout', value: number) {
    pushConfig({ appSettingProps: { [key]: value } as any });
  }
</script>

<Scaffold title={title}>
  {#snippet children()}
    <div class="settings-page">
      <button class="back" onclick={onback}>‹ {t('back')}</button>

      <h2>{t('language')}</h2>
      <div class="list-group">
        {#each locales as item (item.value)}
          <ListRow title={item.label} clickable selected={($config?.appSettingProps?.locale ?? '') === item.value} onclick={() => setLocale(item.value)}>
            {#snippet trailing()}
              {#if ($config?.appSettingProps?.locale ?? '') === item.value}
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--primary)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12l5 5L20 7"/></svg>
              {/if}
            {/snippet}
          </ListRow>
        {/each}
      </div>

      <h2>{t('requestsAndUpdates')}</h2>
      <div class="list-group">
        <ListRow glyph="⏱" title={t('delayTestTimeout')}>
          {#snippet trailing()}
            <div class="inline-input">
              <input
                type="number"
                min="1"
                max="60"
                value={app?.delayTestTimeout ?? 10}
                onchange={(e) => setNumber('delayTestTimeout', Number((e.target as HTMLInputElement).value) || 10)}
              />
              <span class="muted small-text">s</span>
            </div>
          {/snippet}
        </ListRow>
        <div class="divider"></div>
        <ListRow glyph="🔗" title={t('onlyStatisticsProxy')} clickable onclick={() => pushConfig({ appSettingProps: { onlyStatisticsProxy: !app?.onlyStatisticsProxy } })}>
          {#snippet trailing()}
            <button class="switch" class:on={app?.onlyStatisticsProxy} role="switch" aria-checked={app?.onlyStatisticsProxy}></button>
          {/snippet}
        </ListRow>
        <div class="divider"></div>
        <ListRow glyph="📄" title={t('logs')} clickable onclick={() => pushConfig({ appSettingProps: { openLogs: !app?.openLogs } })}>
          {#snippet trailing()}
            <button class="switch" class:on={app?.openLogs} role="switch" aria-checked={app?.openLogs}></button>
          {/snippet}
        </ListRow>
      </div>

      <h2>{t('core')}</h2>
      <div class="list-group">
        <ListRow glyph="🚀" title={t('autoRun')} clickable onclick={() => pushConfig({ appSettingProps: { autoRun: !app?.autoRun } })}>
          {#snippet trailing()}
            <button class="switch" class:on={app?.autoRun} role="switch" aria-checked={app?.autoRun}></button>
          {/snippet}
        </ListRow>
        <div class="divider"></div>
        <ListRow glyph="🔄" title={t('autoCheckUpdate')} clickable onclick={() => pushConfig({ appSettingProps: { autoCheckUpdate: !app?.autoCheckUpdate } })}>
          {#snippet trailing()}
            <button class="switch" class:on={app?.autoCheckUpdate} role="switch" aria-checked={app?.autoCheckUpdate}></button>
          {/snippet}
        </ListRow>
      </div>
    </div>
  {/snippet}
</Scaffold>

<style>
  .settings-page {
    height: 100%;
    overflow-y: auto;
    padding: 4px 26px 26px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 14px;
    max-width: 900px;
  }

  .back {
    align-self: flex-start;
    color: var(--primary);
    font-weight: 590;
    font-size: 0.95rem;
    margin-bottom: 4px;
  }

  h2 {
    font-size: 1.05rem;
    margin: 8px 0 0;
  }

  .inline-input {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .inline-input input {
    width: 70px;
    background: color-mix(in srgb, var(--on-surface) 8%, transparent);
    border: none;
    border-radius: var(--radius-sm);
    padding: 7px 10px;
    text-align: right;
  }
</style>
