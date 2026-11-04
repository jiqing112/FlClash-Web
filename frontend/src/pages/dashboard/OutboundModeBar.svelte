<script lang="ts">
  // OutboundModeBar per outbound_mode.dart: glyph + title left, connected
  // buttons (outer 24 = card radius - inset 8, inner 10) filling the rest.
  // Selection reads the patch config's mode so a click flips instantly.
  import { t } from '../../lib/i18n';
  import { tunnelMode, config, pushConfig } from '../../lib/stores';
  import DashCard from '../../components/dash/DashCard.svelte';
  import ConnectedButton from '../../components/dash/ConnectedButton.svelte';

  const glyph = '<path d="M7 3v6a5 5 0 0 0 10 0V3"/><path d="M12 14v7"/><path d="M8 21h8"/>';

  const modes = $derived([
    { value: 'rule', label: t('rule'), icon: '<rect x="5" y="3" width="14" height="18" rx="2.5"/><path d="M9 8h6M9 12h6M9 16h4"/>' },
    { value: 'global', label: t('global'), icon: '<circle cx="12" cy="12" r="8.8"/><ellipse cx="12" cy="12" rx="3.8" ry="8.8"/><path d="M3.2 12h17.6"/>' },
    { value: 'direct', label: t('direct'), icon: '<circle cx="12" cy="12" r="8.8"/><circle cx="12" cy="12" r="3.4"/><path d="M12 2v4M12 18v4M2 12h4M18 12h4"/>' },
  ]);

  const currentMode = $derived($tunnelMode);

  async function setMode(next: string) {
    if (currentMode === next) return;
    // optimistic flip: the PUT response and the config event confirm it
    if ($config) {
      config.set({
        ...$config,
        patchClashConfig: { ...$config.patchClashConfig, mode: next as 'rule' | 'global' | 'direct' },
      });
    }
    pushConfig({ patchClashConfig: { mode: next as 'rule' | 'global' | 'direct' } });
  }
</script>

<DashCard pad={false}>
  <div class="bar">
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html glyph}</svg>
    <span class="title ellipsis">{t('outboundMode')}</span>
    <div class="buttons">
      {#each modes as item, index (item.value)}
        <ConnectedButton
          first={index === 0}
          last={index === modes.length - 1}
          selected={currentMode === item.value}
          outer={24}
          inner={10}
          onclick={() => setMode(item.value)}
        >
          <span class="mode-content">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html item.icon}</svg>
            {item.label}
          </span>
        </ConnectedButton>
      {/each}
    </div>
  </div>
</DashCard>

<style>
  .bar {
    display: flex;
    align-items: center;
    gap: 16px;
    height: 100%;
    min-width: 0;
    padding: 8px 8px 8px 22px;
    box-sizing: border-box;
  }

  .bar > svg {
    color: var(--on-surface-variant);
    flex-shrink: 0;
  }

  .title {
    font-size: 13px;
    font-weight: 600;
    color: var(--on-surface-variant);
    max-width: 25%;
    flex-shrink: 0;
  }

  .buttons {
    flex: 1;
    display: flex;
    gap: 3px;
    height: 100%;
    min-width: 0;
  }

  .mode-content {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    width: 100%;
    height: 100%;
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
  }
</style>
