<script lang="ts">
  // NetworkTakeoverTile: one switch filling its own card (radius = card).
  import { t } from '../../lib/i18n';
  import { running, config, pushConfig } from '../../lib/stores';
  import DashCard from '../../components/dash/DashCard.svelte';
  import DashSwitchRow from '../../components/dash/DashSwitchRow.svelte';

  let { kind }: { kind: 'systemProxy' | 'tun' } = $props();

  const isSystemProxy = $derived(kind === 'systemProxy');
  const title = $derived(isSystemProxy ? t('systemProxy') : t('tun'));
  const glyph = $derived(
    isSystemProxy
      ? '<path d="M16 3h5v5"/><path d="M8 21H3v-5"/><path d="M21 3l-7 7"/><path d="M3 21l7-7"/>'
      : '<path d="M12 3l8 4.5v9L12 21l-8-4.5v-9z"/><path d="M12 12l8-4.5M12 12v9M12 12L4 7.5"/>',
  );

  const value = $derived(
    isSystemProxy ? ($config?.networkProps.systemProxy ?? false) : ($config?.patchClashConfig.tun.enable ?? false),
  );

  const subtitle = $derived(
    value
      ? isSystemProxy
        ? t('portValue', { port: $config?.patchClashConfig['mixed-port'] ?? 7890 })
        : t('enabled')
      : t('disabled'),
  );

  function change(next: boolean) {
    if (isSystemProxy) {
      pushConfig({ networkProps: { systemProxy: next } });
    } else {
      pushConfig({ patchClashConfig: { tun: { ...(($config?.patchClashConfig.tun ?? {}) as any), enable: next } } });
    }
  }
</script>

<DashCard pad={false}>
  <div class="fill">
    <DashSwitchRow
      {glyph}
      {title}
      {subtitle}
      {value}
      iconSize={40}
      pad="0 16px 0 12px"
      radius={32}
      color="transparent"
      onchange={change}
      onclick={() => change(!value)}
    />
  </div>
</DashCard>

<style>
  .fill {
    height: 100%;
    width: 100%;
    display: flex;
    flex-direction: column;
  }

  .fill :global(.switch-row) {
    flex: 1;
  }
</style>
