<script lang="ts">
  // ProfilesCard per profiles.dart: name + updated/expiry row + usage meter
  // with the wavy indicator, or config counts.
  import { t } from '../../lib/i18n';
  import { currentProfile, groups, showBanner } from '../../lib/stores';
  import { api } from '../../lib/api';
  import { trafficText, relativeTimeText } from '../../lib/format';
  import DashCard from '../../components/dash/DashCard.svelte';
  import DashHeader from '../../components/dash/DashHeader.svelte';
  import WavyProgress from '../../components/dash/WavyProgress.svelte';

  let { mobile = false }: { mobile?: boolean } = $props();

  const profilesGlyph = '<rect x="5" y="3" width="14" height="18" rx="2.5"/><path d="M9 8.5h6M9 12.5h6M9 16h4"/>';
  const chevronGlyph = '<path d="M9 6l6 6-6 6"/>';
  const syncGlyph = '<path d="M20 12a8 8 0 1 1-2.3-5.7"/><path d="M20 3v4h-4"/>';

  const profile = $derived($currentProfile);
  const info = $derived(profile?.subscriptionInfo);
  const usage = $derived.by(() => {
    if (!info || info.total <= 0) return null;
    return {
      used: info.upload + info.download,
      total: info.total,
      share: Math.min(1, (info.upload + info.download) / info.total),
    };
  });
  const expiryDays = $derived.by(() => {
    if (!info?.expire) return null;
    return Math.max(0, Math.ceil((info.expire * 1000 - Date.now()) / 86_400_000));
  });
  const sourceLabel = $derived(
    profile == null ? '' : profile.type === 'url' ? t('remote') : profile.type === 'custom' ? t('custom') : t('local'),
  );

  let syncing = $state(false);

  async function sync() {
    if (!profile) return;
    syncing = true;
    try {
      await api.updateProfile(profile.id);
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    } finally {
      syncing = false;
    }
  }
</script>

<DashCard>
  <DashHeader glyph={profilesGlyph} title={t('profile')}>
    {#snippet trailing()}
      {#if profile?.type === 'url'}
        <button class="card-icon" onclick={sync} title={t('update')}>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" class:spinning={syncing || profile.updating}>{@html syncGlyph}</svg>
        </button>
      {/if}
      <button class="card-icon" title={t('details', { label: t('profile') })}>
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">{@html chevronGlyph}</svg>
      </button>
    {/snippet}
  </DashHeader>
  <div style="height: 10px"></div>
  {#if profile}
    <div class="name ellipsis">{profile.label || profile.id}</div>
    <div class="meta-row">
      <span class="meta ellipsis">
        {#if profile.lastUpdateDate}
          {t('profileUpdatedAt', { time: relativeTimeText(profile.lastUpdateDate, t) })}
        {/if}
      </span>
      <span class="meta ellipsis" class:soon={expiryDays != null && expiryDays <= 7}>
        {#if expiryDays != null}
          {t('profileDaysLeft', { days: expiryDays })}
        {:else}
          {sourceLabel}
        {/if}
      </span>
    </div>
    <div class="spacer"></div>
    {#if usage}
      <div class="usage-row">
        <span class="usage-figure tabular">{Math.round(usage.share * 100)}%<span class="usage-unit"> {t('usedShort')}</span></span>
        <span class="traffic ellipsis">{trafficText(usage.used)} / {trafficText(usage.total)}</span>
      </div>
      <div class="wave">
        <WavyProgress value={usage.share} warn={usage.share >= 0.9} />
      </div>
    {:else}
      <div class="stats-row">
        <div class="stat"><span class="stat-figure tabular">{$groups.length}</span><span class="stat-label">{t('proxyGroup')}</span></div>
        <div class="stat"><span class="stat-figure tabular">—</span><span class="stat-label">{t('proxies')}</span></div>
        <div class="stat"><span class="stat-figure tabular">—</span><span class="stat-label">{t('rules')}</span></div>
      </div>
    {/if}
  {:else}
    <div class="name muted">{t('addProfile')}</div>
  {/if}
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

  .spinning {
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .name {
    font-size: 17px;
    line-height: 22px;
    font-weight: 700;
    color: var(--on-surface);
  }

  .meta-row {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    margin-top: 2px;
  }

  .meta {
    font-size: 12px;
    color: var(--on-surface-variant);
    font-variant-numeric: tabular-nums;
  }

  .meta.soon {
    color: var(--warning);
  }

  .spacer {
    flex: 1;
  }

  .usage-row {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 12px;
  }

  .usage-figure {
    font-size: 17px;
    line-height: 22px;
    font-weight: 700;
    letter-spacing: 0;
  }

  .usage-unit {
    font-size: 13px;
    font-weight: 600;
    color: var(--on-surface-variant);
  }

  .traffic {
    font-size: 13px;
    color: var(--on-surface-variant);
    font-variant-numeric: tabular-nums;
  }

  .wave {
    margin-top: 6px;
  }

  .stats-row {
    display: flex;
    justify-content: space-between;
    gap: 16px;
  }

  .stat {
    display: flex;
    flex-direction: column;
  }

  .stat-figure {
    font-size: 17px;
    line-height: 22px;
    font-weight: 700;
  }

  .stat-label {
    font-size: 12px;
    color: var(--on-surface-variant);
  }
</style>
