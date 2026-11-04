<script lang="ts">
  // About page.
  import Scaffold from '../../components/Scaffold.svelte';
  import ListRow from '../../components/ListRow.svelte';
  import { t } from '../../lib/i18n';

  let { title, onback }: { title: string; onback: () => void } = $props();

  const version = 'v0.9.3-web';
</script>

<Scaffold title={title}>
  {#snippet children()}
    <div class="settings-page">
      <button class="back" onclick={onback}>‹ {t('back')}</button>
      <div class="hero">
        <div class="logo">Fl</div>
        <div class="name">FlClash Web</div>
        <span class="chip">{version}</span>
      </div>
      <p class="desc muted">A multi-platform proxy client based on ClashMeta (mihomo), simple and easy to use, open-source and ad-free. Web edition.</p>
      <div class="list-group">
        <ListRow glyph="🐙" title="GitHub" subtitle="github.com/chen08209/FlClash" clickable onclick={() => window.open('https://github.com/chen08209/FlClash', '_blank')} />
        <div class="divider"></div>
        <ListRow glyph="⚙️" title="Core" subtitle="github.com/chen08209/Clash.Meta" clickable onclick={() => window.open('https://github.com/chen08209/Clash.Meta', '_blank')} />
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
    gap: 16px;
    max-width: 900px;
  }

  .back {
    align-self: flex-start;
    color: var(--primary);
    font-weight: 590;
    font-size: 0.95rem;
  }

  .hero {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .logo {
    width: 56px;
    height: 56px;
    border-radius: 18px;
    background: var(--primary);
    color: var(--on-primary);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.4rem;
    font-weight: 800;
  }

  .name {
    font-size: 1.3rem;
    font-weight: 700;
  }

  .desc {
    margin: 0;
    font-size: 0.92rem;
    line-height: 1.6;
  }
</style>
