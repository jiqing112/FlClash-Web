<script lang="ts">
  // Theme settings: mode, seed colors, color style, pure black, text scale.
  import Scaffold from '../../components/Scaffold.svelte';
  import ListRow from '../../components/ListRow.svelte';
  import { config, pushConfig } from '../../lib/stores';
  import { t } from '../../lib/i18n';

  let { title, onback }: { title: string; onback: () => void } = $props();

  const theme = $derived($config?.themeProps);
  const isDarkNow = $derived(document.documentElement.classList.contains('dark'));

  function setThemeMode(mode: 'system' | 'light' | 'dark') {
    pushConfig({ themeProps: { themeMode: mode } });
  }

  function setColor(color: number) {
    pushConfig({ themeProps: { primaryColor: color } });
  }

  function setStyle(style: string) {
    pushConfig({ themeProps: { colorStyle: style as any } });
  }

  function togglePureBlack() {
    pushConfig({ themeProps: { pureBlack: !theme?.pureBlack } });
  }

  function resetColor() {
    pushConfig({ themeProps: { primaryColor: undefined } });
  }
</script>

<Scaffold title={title}>
  {#snippet children()}
    <div class="settings-page">
      <button class="back" onclick={onback}>‹ {t('back')}</button>

      <h2>{t('themeMode')}</h2>
      <div class="connected-pills">
        <button class:selected={theme?.themeMode === 'system'} onclick={() => setThemeMode('system')}>{t('followSystem')}</button>
        <button class:selected={theme?.themeMode === 'light'} onclick={() => setThemeMode('light')}>{t('light')}</button>
        <button class:selected={theme?.themeMode === 'dark'} onclick={() => setThemeMode('dark')}>{t('dark')}</button>
      </div>

      <h2>{t('themeColor')}</h2>
      <div class="swatches">
        {#each theme?.primaryColors ?? [] as color (color)}
          <button
            class="swatch"
            class:selected={theme?.primaryColor === color}
            style="background: #{(color & 0xffffff).toString(16).padStart(6, '0')}"
            onclick={() => setColor(color)}
            title={'#' + (color & 0xffffff).toString(16)}
          >
            {#if theme?.primaryColor === color}
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12l5 5L20 7"/></svg>
            {/if}
          </button>
        {/each}
      </div>
      <div class="row">
        <button class="tonal-button small" onclick={resetColor}>{t('reset')}</button>
      </div>

      <h2>{t('colorStyle')}</h2>
      <div class="connected-pills">
        <button class:selected={theme?.colorStyle === 'vivid'} onclick={() => setStyle('vivid')}>{t('vividStyle')}</button>
        <button class:selected={theme?.colorStyle === 'tinted'} onclick={() => setStyle('tinted')}>{t('tintedStyle')}</button>
        <button class:selected={theme?.colorStyle === 'soft'} onclick={() => setStyle('soft')}>{t('softStyle')}</button>
        <button class:selected={theme?.colorStyle === 'monochrome'} onclick={() => setStyle('monochrome')}>{t('monochromeScheme')}</button>
      </div>

      <h2>{t('pureBlackMode')}</h2>
      <div class="list-group">
        <ListRow title={t('pureBlack')} clickable onclick={togglePureBlack}>
          {#snippet trailing()}
            <button class="switch" class:on={theme?.pureBlack} role="switch" aria-checked={theme?.pureBlack}></button>
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

  .swatches {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }

  .swatch {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: transform 0.15s ease;
    border: 2px solid transparent;
  }

  .swatch:hover {
    transform: scale(1.1);
  }

  .swatch.selected {
    border-color: var(--on-surface);
  }
</style>
