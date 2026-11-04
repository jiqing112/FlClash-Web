<script lang="ts">
  // Profiles: card grid, add via URL/file/clipboard, edit sheet, select, sync.
  import { t } from '../lib/i18n';
  import Scaffold from '../components/Scaffold.svelte';
  import Sheet from '../components/Sheet.svelte';
  import Dialog from '../components/Dialog.svelte';
  import PopupMenu from '../components/PopupMenu.svelte';
  import { profiles, config, showBanner, pushConfig } from '../lib/stores';
  import { api } from '../lib/api';
  import { trafficText, dateText } from '../lib/format';

  let addOpen = $state(false);
  let addUrl = $state('');
  let adding = $state(false);
  let editOpen = $state(false);
  let editId = $state<number | null>(null);
  let editLabel = $state('');
  let editUrl = $state('');
  let editAutoUpdate = $state(true);
  let editInterval = $state(120);
  let deletingProfile = $state<any | null>(null);
  let syncing = $state(false);
  let fileInput = $state<HTMLInputElement | null>(null);

  const currentId = $derived($config?.currentProfileId ?? null);
  const hasProfiles = $derived($profiles.length > 0);

  function usageOf(profile: any) {
    const info = profile.subscriptionInfo;
    if (!info || info.total <= 0) return null;
    const used = info.upload + info.download;
    return {
      percent: Math.min(1, used / info.total),
      used: trafficText(used),
      total: trafficText(info.total),
      expireDays: info.expire > 0 ? Math.ceil((info.expire * 1000 - Date.now()) / 86_400_000) : null,
    };
  }

  async function addFromUrl() {
    if (!addUrl.trim()) return;
    adding = true;
    try {
      await api.addProfile({ url: addUrl.trim() });
      addUrl = '';
      addOpen = false;
      const updated = await api.profiles();
      pushConfig({});
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    } finally {
      adding = false;
    }
  }

  async function onFilePicked(event: Event) {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    adding = true;
    try {
      const content = await file.text();
      await api.addProfile({ type: 'file', label: file.name, fileContent: content });
      addOpen = false;
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    } finally {
      adding = false;
      input.value = '';
    }
  }

  function openEdit(profile: any) {
    editId = profile.id;
    editLabel = profile.label;
    editUrl = profile.url;
    editAutoUpdate = profile.autoUpdate;
    editInterval = profile.autoUpdateDuration || 120;
    editOpen = true;
  }

  async function saveEdit() {
    if (editId == null) return;
    try {
      await api.editProfile(editId, {
        label: editLabel,
        url: editUrl,
        autoUpdate: editAutoUpdate,
        autoUpdateDuration: Number(editInterval) || 120,
      });
      editOpen = false;
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    }
  }

  async function doDelete() {
    if (!deletingProfile) return;
    try {
      await api.deleteProfile(deletingProfile.id);
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    } finally {
      deletingProfile = null;
    }
  }

  async function selectProfile(id: number) {
    try {
      await api.selectProfile(id);
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    }
  }

  async function syncAll() {
    syncing = true;
    try {
      for (const profile of $profiles) {
        if (profile.type === 'url') {
          try {
            await api.updateProfile(profile.id);
          } catch (error) {
            showBanner(`${profile.label}: ${String((error as Error).message ?? error)}`, 'error');
          }
        }
      }
    } finally {
      syncing = false;
    }
  }

  async function pasteAdd() {
    try {
      const text = await navigator.clipboard.readText();
      if (text) {
        addUrl = text.trim();
        addOpen = true;
      }
    } catch {
      addOpen = true;
    }
  }
</script>

<Scaffold
  title={t('profiles')}
  loading={syncing}
  actions={[
    { icon: '<path d="M20 12a8 8 0 1 1-2.3-5.7"/><path d="M20 3v4h-4"/>', label: t('sync'), onClick: syncAll, show: hasProfiles },
  ]}
>
  {#snippet children()}
    <div class="profiles-page">
      {#if !hasProfiles}
        <div class="empty">
          <div class="empty-title muted">{t('profiles')}</div>
          <button class="tonal-button" onclick={() => (addOpen = true)}>+ {t('addProfile')}</button>
        </div>
      {:else}
        <div class="grid">
          {#each $profiles as profile (profile.id)}
            <div
              class="profile-card app-card clickable"
              class:selected={currentId === profile.id}
              role="button"
              tabindex="0"
              onclick={() => selectProfile(profile.id)}
              onkeydown={(e) => { if (e.key === 'Enter') selectProfile(profile.id); }}
            >
              <div class="row-1">
                <div class="name ellipsis">{profile.label || profile.id}</div>
                {#if currentId === profile.id}
                  <span class="selection-badge">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12l5 5L20 7"/></svg>
                  </span>
                {/if}
                <PopupMenu align="end">
                  {#snippet children()}
                    <button class="icon-button small" title={t('edit')}>
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M4 20h4L20.5 7.5a2.1 2.1 0 0 0-3-3L5 17l-1 4z"/></svg>
                    </button>
                  {/snippet}
                  {#snippet menu()}
                    <button class="popup-item" onclick={() => openEdit(profile)}>{t('edit')}</button>
                    <button class="popup-item" onclick={() => pasteAdd()}>{t('addProfile')}</button>
                    {#if profile.type === 'url'}
                      <button class="popup-item" onclick={() => api.updateProfile(profile.id).catch((e) => showBanner(String(e), 'error'))}>{t('sync')}</button>
                    {/if}
                    <button class="popup-item danger" onclick={() => (deletingProfile = profile)}>{t('delete')}</button>
                  {/snippet}
                </PopupMenu>
              </div>
              <div class="row-2 muted">
                {#if usageOf(profile)}
                  <span>{usageOf(profile)!.used} / {usageOf(profile)!.total}</span>
                  {#if usageOf(profile)!.expireDays != null}
                    <span class="chip small-chip">{t('profileDaysLeft', { days: usageOf(profile)!.expireDays! })}</span>
                  {/if}
                {:else}
                  <span>{profile.type === 'url' ? t('remote') : profile.type === 'custom' ? t('custom') : t('local')}</span>
                {/if}
              </div>
              <div class="row-3 muted">
                {#if profile.updating}
                  <span class="skeleton" style="width:110px;height:12px"></span>
                {:else if profile.lastUpdateDate}
                  {dateText(profile.lastUpdateDate)}
                {/if}
              </div>
              {#if usageOf(profile)}
                <div class="linear-progress slim"><div style="width: {usageOf(profile)!.percent * 100}%"></div></div>
              {/if}
            </div>
          {/each}
          <button class="add-card app-card clickable" onclick={() => (addOpen = true)}>+</button>
        </div>
      {/if}
    </div>
  {/snippet}
</Scaffold>

<!-- add sheet -->
<Sheet bind:open={addOpen} title={t('addProfile')} width={440}>
  <div class="text-field">
    <label>{t('url')}</label>
    <input bind:value={addUrl} placeholder="https://..." onkeydown={(e) => { if (e.key === 'Enter') addFromUrl(); }} />
  </div>
  <div class="row">
    <button class="tonal-button" disabled={adding || !addUrl.trim()} onclick={addFromUrl}>{t('add')}</button>
    <div class="spacer"></div>
    <button class="tonal-button" onclick={() => fileInput?.click()}>📁 {t('file')}</button>
    <input type="file" accept=".yaml,.yml,.txt" style="display:none" bind:this={fileInput} onchange={onFilePicked} />
  </div>
</Sheet>

<!-- edit sheet -->
<Sheet bind:open={editOpen} title={t('edit')} width={440}>
  {#snippet footer()}
    <button class="tonal-button" onclick={() => (editOpen = false)}>{t('cancel')}</button>
    <button class="tonal-button filled" onclick={saveEdit}>{t('save')}</button>
  {/snippet}
  <div class="text-field">
    <label>{t('name')}</label>
    <input bind:value={editLabel} />
  </div>
  <div class="text-field">
    <label>{t('url')}</label>
    <input bind:value={editUrl} />
  </div>
  <div class="row list-row-like">
    <span>{t('autoUpdate')}</span>
    <div class="spacer"></div>
    <button class="switch" class:on={editAutoUpdate} onclick={() => (editAutoUpdate = !editAutoUpdate)}></button>
  </div>
  {#if editAutoUpdate}
    <div class="text-field">
      <label>{t('autoUpdateInterval')} (min)</label>
      <input type="number" bind:value={editInterval} min="15" />
    </div>
  {/if}
</Sheet>

<Dialog
  open={deletingProfile != null}
  title={t('delete')}
  actions={[
    { label: t('cancel'), onClick: () => (deletingProfile = null) },
    { label: t('delete'), onClick: doDelete, destructive: true, primary: true },
  ]}
>
  <p>{t('deleteTip', { label: deletingProfile?.label ?? '' })}</p>
</Dialog>

<style>
  .profiles-page {
    height: 100%;
    overflow-y: auto;
    padding: 4px 26px 26px;
    box-sizing: border-box;
  }

  .empty {
    height: 60vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
  }

  .empty-title {
    font-size: 1.2rem;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 14px;
  }

  .profile-card {
    padding: 16px 18px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    border-radius: var(--radius-xl);
    position: relative;
  }

  .row-1 {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .name {
    flex: 1;
    font-weight: 650;
    font-size: 1.05rem;
  }

  .selection-badge {
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--primary);
    color: var(--on-primary);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .row-2 {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.86rem;
  }

  .row-3 {
    font-size: 0.78rem;
  }

  .linear-progress.slim {
    height: 4px;
  }

  .add-card {
    min-height: 110px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 2rem;
    color: var(--on-surface-variant);
    border-radius: var(--radius-xl);
    border: 1.5px dashed color-mix(in srgb, var(--on-surface) 20%, transparent);
    background: transparent;
    box-shadow: none;
  }

  .list-row-like {
    padding: 10px 0;
  }

  .small-chip {
    height: 22px;
    font-size: 0.72rem;
    padding: 0 8px;
  }
</style>
