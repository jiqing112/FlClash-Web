<script lang="ts">
  // IntranetIpCard per intranet_ip.dart: header + copy, title2 figure, LAN
  // switch row (radius 20, icon 38) at the bottom.
  import { t } from '../../lib/i18n';
  import { config, pushConfig } from '../../lib/stores';
  import DashCard from '../../components/dash/DashCard.svelte';
  import DashHeader from '../../components/dash/DashHeader.svelte';
  import DashSwitchRow from '../../components/dash/DashSwitchRow.svelte';

  let { mobile = false }: { mobile?: boolean } = $props();

  const devicesGlyph = '<rect x="2.5" y="5" width="14" height="11" rx="2"/><path d="M7 19.5h11a2.5 2.5 0 0 0 2.5-2.5V9"/>';
  let intranetIp = $state<string>('');
  let copied = $state(false);

  $effect(() => {
    fetch('/api/intranet')
      .then((response) => response.json())
      .then((data) => (intranetIp = data.ip ?? ''))
      .catch(() => (intranetIp = ''));
  });

  async function copyIp() {
    const allowLan = $config?.patchClashConfig['allow-lan'] ?? false;
    const port = $config?.patchClashConfig['mixed-port'] ?? 7890;
    await navigator.clipboard.writeText(allowLan ? `${intranetIp}:${port}` : intranetIp);
    copied = true;
    setTimeout(() => (copied = false), 1500);
  }

  const allowLan = $derived($config?.patchClashConfig['allow-lan'] ?? false);
  const port = $derived($config?.patchClashConfig['mixed-port'] ?? 7890);
</script>

<DashCard>
  <DashHeader glyph={devicesGlyph} title={t('intranetIP')}>
    {#snippet trailing()}
      <button class="copy" onclick={copyIp} title={t('copy')}>
        {#if copied}
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12l5 5L20 7"/></svg>
        {:else}
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="11" height="11" rx="2"/><path d="M5 15V6a2 2 0 0 1 2-2h9"/></svg>
        {/if}
      </button>
    {/snippet}
  </DashHeader>
  <div style="height: {mobile ? 12 : 10}px"></div>
  <div class="ip figure ellipsis">
    {#if intranetIp}
      {allowLan ? `${intranetIp}:${port}` : intranetIp}
    {:else}
      <span class="skeleton" style="width: 120px; height: 22px; display: inline-block"></span>
    {/if}
  </div>
  <div class="spacer"></div>
  <DashSwitchRow
    glyph="<path d='M5 12.5a10 10 0 0 1 14 0'/><path d='M8.5 15.5a6 6 0 0 1 7 0'/><circle cx='12' cy='18.5' r='1.3' fill='currentColor' stroke='none'/>"
    title={t('allowLan')}
    subtitle={allowLan ? t('portValue', { port }) : t('disabled')}
    value={allowLan}
    iconSize={38}
    pad="10px 12px 10px 10px"
    radius={20}
    onchange={(next) => pushConfig({ patchClashConfig: { 'allow-lan': next } })}
    onclick={() => pushConfig({ patchClashConfig: { 'allow-lan': !allowLan } })}
  />
</DashCard>

<style>
  .copy {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--on-surface-variant);
  }

  .copy:hover {
    background: color-mix(in srgb, var(--on-surface) 7%, transparent);
  }

  .ip {
    font-size: 22px;
    line-height: 28px;
    font-weight: 700;
    letter-spacing: 0;
    color: var(--on-surface);
  }

  .spacer {
    flex: 1;
  }
</style>
