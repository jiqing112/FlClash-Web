<script lang="ts">
  // Network settings: ports, LAN, system proxy, TUN, keep-alive, overrides.
  import Scaffold from '../../components/Scaffold.svelte';
  import ListRow from '../../components/ListRow.svelte';
  import { config, pushConfig, showBanner } from '../../lib/stores';
  import { api } from '../../lib/api';
  import { t } from '../../lib/i18n';

  let { title, onback }: { title: string; onback: () => void } = $props();

  const patch = $derived($config?.patchClashConfig);
  const network = $derived($config?.networkProps);

  function num(value: string, fallback: number): number {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : fallback;
  }

  async function reopenPorts() {
    // ports only take effect on a full setup
    try {
      await api.applyConfig();
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    }
  }
</script>

<Scaffold title={title}>
  {#snippet children()}
    <div class="settings-page">
      <button class="back" onclick={onback}>‹ {t('back')}</button>

      <h2>{t('inbound')}</h2>
      <div class="list-group">
        <ListRow glyph="🔌" title={t('mixedPort')}>
          {#snippet trailing()}
            <div class="inline-input">
              <input
                type="number"
                min="0"
                max="65535"
                value={patch?.['mixed-port'] ?? 7890}
                onchange={(e) => pushConfig({ patchClashConfig: { 'mixed-port': num((e.target as HTMLInputElement).value, 7890) } })}
              />
            </div>
          {/snippet}
        </ListRow>
        <div class="divider"></div>
        <ListRow glyph="🏰" title={t('allowLan')} clickable onclick={() => pushConfig({ patchClashConfig: { 'allow-lan': !(patch?.['allow-lan'] ?? false) } })}>
          {#snippet trailing()}
            <button class="switch" class:on={patch?.['allow-lan']} role="switch" aria-checked={patch?.['allow-lan']}></button>
          {/snippet}
        </ListRow>
        <div class="divider"></div>
        <ListRow glyph="🖥" title={t('systemProxy')} clickable onclick={() => pushConfig({ networkProps: { systemProxy: !(network?.systemProxy ?? false) } })}>
          {#snippet trailing()}
            <button class="switch" class:on={network?.systemProxy} role="switch" aria-checked={network?.systemProxy}></button>
          {/snippet}
        </ListRow>
        <div class="divider"></div>
        <ListRow glyph="🧩" title="TUN" clickable onclick={() => pushConfig({ patchClashConfig: { tun: { ...(patch?.tun as any), enable: !(patch?.tun?.enable ?? false) } } })}>
          {#snippet trailing()}
            <button class="switch" class:on={patch?.tun?.enable} role="switch" aria-checked={patch?.tun?.enable}></button>
          {/snippet}
        </ListRow>
      </div>
      <div class="row">
        <button class="tonal-button small" onclick={reopenPorts}>{t('applyNow')}</button>
        <span class="muted small-text">{t('portsApplyHint')}</span>
      </div>

      <h2>{t('connection')}</h2>
      <div class="list-group">
        <ListRow glyph="🔎" title={t('findProcessMode')}>
          {#snippet trailing()}
            <div class="connected-pills">
              <button class:selected={patch?.['find-process-mode'] === 'always'} onclick={() => pushConfig({ patchClashConfig: { 'find-process-mode': 'always' } })}>{t('always')}</button>
              <button class:selected={patch?.['find-process-mode'] === 'off'} onclick={() => pushConfig({ patchClashConfig: { 'find-process-mode': 'off' } })}>{t('off')}</button>
            </div>
          {/snippet}
        </ListRow>
        <div class="divider"></div>
        <ListRow glyph="⏳" title={t('keepAliveInterval')}>
          {#snippet trailing()}
            <div class="inline-input">
              <input
                type="number"
                value={patch?.['keep-alive-interval'] ?? 30}
                onchange={(e) => pushConfig({ patchClashConfig: { 'keep-alive-interval': num((e.target as HTMLInputElement).value, 30) } })}
              />
            </div>
          {/snippet}
        </ListRow>
        <div class="divider"></div>
        <ListRow glyph="⚡" title={t('tcpConcurrent')} clickable onclick={() => pushConfig({ patchClashConfig: { 'tcp-concurrent': !(patch?.['tcp-concurrent'] ?? false) } })}>
          {#snippet trailing()}
            <button class="switch" class:on={patch?.['tcp-concurrent']} role="switch" aria-checked={patch?.['tcp-concurrent']}></button>
          {/snippet}
        </ListRow>
        <div class="divider"></div>
        <ListRow glyph="🌍" title="IPv6" clickable onclick={() => pushConfig({ patchClashConfig: { ipv6: !(patch?.ipv6 ?? false) } })}>
          {#snippet trailing()}
            <button class="switch" class:on={patch?.ipv6} role="switch" aria-checked={patch?.ipv6}></button>
          {/snippet}
        </ListRow>
      </div>

      <h2>{t('testUrl')}</h2>
      <div class="text-field">
        <input
          value={patch?.['test-url'] ?? ''}
          onchange={(e) => pushConfig({ patchClashConfig: { 'test-url': (e.target as HTMLInputElement).value } })}
        />
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

  .inline-input input {
    width: 100px;
    background: color-mix(in srgb, var(--on-surface) 8%, transparent);
    border: none;
    border-radius: var(--radius-sm);
    padding: 7px 10px;
    text-align: right;
  }
</style>
