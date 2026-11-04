<script lang="ts">
  // Status card per current_outbound.dart: cookie-12 blob behind a 104px
  // morphing start button, clover country badge, IP figure, quality chips.
  import { t } from '../../lib/i18n';
  import { api } from '../../lib/api';
  import { running, setRunningState, showBanner, tunnelMode } from '../../lib/stores';
  import { durationText } from '../../lib/format';
  import { materialShapes } from '../../lib/shapes';
  import DashCard from '../../components/dash/DashCard.svelte';
  import DashHeader from '../../components/dash/DashHeader.svelte';
  import DashChip from '../../components/dash/DashChip.svelte';
  import ShapeBadge from '../../components/dash/ShapeBadge.svelte';

  let { mobile = false }: { mobile?: boolean } = $props();

  const statusGlyph = '<path d="M3.5 12h3l2-4 3 8 2-4h7"/><rect x="2" y="4" width="20" height="16" rx="3"/>';
  let ipInfo = $state<{ ip: string; countryCode?: string; organization?: string; type?: string } | null>(null);
  let probing = $state(false);
  let ipDelay = $state<number | null>(null);
  let runTimeValue = $state(0);

  $effect(() => {
    const timer = setInterval(() => {
      if ($running.running && $running.startTime) {
        runTimeValue = Date.now() - $running.startTime;
      } else {
        runTimeValue = 0;
      }
    }, 1000);
    return () => clearInterval(timer);
  });

  let quality = $state<{ type: string; organization?: string; level: string } | null>(null);

  async function probe() {
    probing = true;
    try {
      const raw: any = await api.probeIP();
      if (raw && !raw.error && raw.body) {
        ipDelay = raw.delay ?? null;
        ipInfo = parseBody(raw.url, raw.body);
        quality = null;
        if (ipInfo?.ip) {
          api.ipQuality(ipInfo.ip)
            .then((answer) => (quality = answer))
            .catch(() => (quality = null));
        }
      }
    } catch {
      // offline: fine
    } finally {
      probing = false;
    }
  }

  function parseBody(url: string, body: string): { ip: string; countryCode?: string; organization?: string; type?: string } | null {
    try {
      if (url.includes('cloudflare.com/cdn-cgi/trace')) {
        const fields: Record<string, string> = {};
        for (const line of body.split('\n')) {
          const [key, value] = line.split('=');
          if (key && value) fields[key.trim()] = value.trim();
        }
        return { ip: fields.ip ?? '', countryCode: fields.loc };
      }
      const json = JSON.parse(body);
      if (url.includes('ip.sb')) {
        return {
          ip: json.ip ?? '',
          countryCode: json.country_code ?? json.country,
          organization: json.organization ?? json.isp ?? json.asn_organization,
          type: json.type,
        };
      }
      if (url.includes('ipinfo.io')) {
        return { ip: json.ip ?? '', countryCode: json.country, organization: json.org };
      }
      if (typeof json.ip === 'string') {
        return { ip: json.ip, countryCode: json.country ?? json.country_code, organization: json.organization ?? json.isp };
      }
    } catch {
      // not json
    }
    return null;
  }

  $effect(() => {
    probe();
    const timer = setInterval(probe, 60_000);
    return () => clearInterval(timer);
  });

  async function toggleRun() {
    try {
      if ($running.running) {
        setRunningState(await api.coreStop());
      } else {
        setRunningState(await api.coreStart());
      }
    } catch (error) {
      showBanner(String((error as Error).message ?? error), 'error');
    }
  }

  function levelLabel(level: string): string {
    switch (level) {
      case 'good': return '优';
      case 'risky': return '风险';
      default: return '普通';
    }
  }

  function typeLabel(type: string): string {
    switch (type.toLowerCase()) {
      case 'hosting': return t('ipTypeHosting');
      case 'business': return t('ipTypeBusiness');
      case 'isp': return t('ipTypeIsp');
      case 'residential': return t('ipTypeResidential');
      case 'education': return t('ipTypeEducation');
      case 'government': return t('ipTypeGovernment');
      case 'military': return t('ipTypeMilitary');
      default: return type;
    }
  }

  function qualityLabel(delay: number): string {
    if (delay <= 200) return '优';
    if (delay <= 400) return '良';
    return '普通';
  }

  const buttonSize = $derived(mobile ? 88 : 104);
  const isRunning = $derived($running.running);
</script>

<DashCard pad={false} clip>
  <div class="blob" style="width:{mobile ? 240 : 290}px; height:{mobile ? 240 : 290}px; right:{mobile ? -60 : -64}px; top:{mobile ? -66 : -87}px">
    <svg viewBox="0 0 1 1" style="width:100%;height:100%;display:block">
      <path
        d={materialShapes.cookie12Sided}
        fill={isRunning ? 'color-mix(in srgb, var(--primary) 10%, var(--surface-bright))' : 'var(--surface-container)'}
        style="transition: fill 0.45s ease"
      />
    </svg>
  </div>
  <div class="content" style="padding: {mobile ? '18px 16px' : '20px 22px'}">
    <div class="main">
      <div class="left">
        <DashHeader glyph={statusGlyph} title={t('status')}>
          {#snippet leading()}
            {#if isRunning}
              <DashChip label={t('running')} tone="ok" dot />
            {:else}
              <DashChip label={t('notRunning')} />
            {/if}
          {/snippet}
        </DashHeader>
        <div style="height: {mobile ? 12 : 10}px"></div>
        <div class="exit-row">
          <button class="country" class:spinning={probing} onclick={probe} title={t('retry')}>
            <ShapeBadge shape="clover4Leaf" size={mobile ? 44 : 48} color="var(--surface-container-high)">
              <span class="country-code">{ipInfo?.countryCode?.toUpperCase() ?? '--'}</span>
            </ShapeBadge>
          </button>
          <div class="ip-col">
            <div class="ip figure" style="font-size: {mobile ? 22 : 28}px; line-height: {mobile ? 28 : 34}px">
              {#if ipInfo?.ip}
                {ipInfo.ip}
              {:else}
                <span class="skeleton" style="width: 140px; height: 26px; display:inline-block"></span>
              {/if}
            </div>
            <div class="chips">
              {#if quality}
                <DashChip label={typeLabel(quality.type)} />
                <DashChip label={t('ipQualityShort') + ' ' + levelLabel(quality.level)} tone={quality.level === 'good' ? 'ok' : quality.level === 'risky' ? 'warn' : 'idle'} />
              {:else if ipDelay != null && ipDelay > 0}
                <DashChip label={t('ipQualityShort') + ' ' + qualityLabel(ipDelay)} tone={ipDelay <= 400 ? 'ok' : 'idle'} />
              {:else if ipInfo}
                <DashChip label={t('ipTypeUnknown')} />
              {:else}
                <span class="skeleton" style="width: 48px; height: 24px; border-radius: 12px; display:inline-block"></span>
                <span class="skeleton" style="width: 72px; height: 24px; border-radius: 12px; display:inline-block"></span>
              {/if}
            </div>
          </div>
        </div>
        <div class="spacer"></div>
        <div class="cells">
          <div class="cell-item">
            <div class="cell-label">{t('runTime')}</div>
            <div class="cell-value tabular">{runTimeValue > 0 ? durationText(runTimeValue) : '—'}</div>
          </div>
          <div class="cell-item">
            <div class="cell-label">{t('ipOrganization')}</div>
            <div class="cell-value ellipsis">{quality?.organization || ipInfo?.organization || '—'}</div>
          </div>
        </div>
      </div>
      <div class="button-slot" style="width: {mobile ? 88 : 120}px">
        <button
          class="start-toggle"
          class:running={isRunning}
          style="width:{buttonSize}px;height:{buttonSize}px"
          onclick={toggleRun}
          title={isRunning ? t('stop') : t('start')}
        >
          <svg viewBox="0 0 1 1" style="width:100%;height:100%;display:block;position:absolute;inset:0">
            <path d={materialShapes.cookie9Sided} fill={isRunning ? 'var(--primary)' : 'var(--surface-container-high)'} style="transition: fill 0.3s ease" />
          </svg>
          <span class="glyph">
            {#if isRunning}
              <svg width={buttonSize * 0.34} height={buttonSize * 0.34} viewBox="0 0 24 24" fill="currentColor"><rect x="7" y="6" width="3.4" height="12" rx="1.4" /><rect x="13.6" y="6" width="3.4" height="12" rx="1.4" /></svg>
            {:else}
              <svg width={buttonSize * 0.34} height={buttonSize * 0.34} viewBox="0 0 24 24" fill="currentColor"><path d="M9 6.5c0-.8.9-1.3 1.6-.9l7 4.6c.6.4.6 1.3 0 1.7l-7 4.6c-.7.4-1.6-.1-1.6-.9z" /></svg>
            {/if}
          </span>
        </button>
      </div>
    </div>
  </div>
</DashCard>

<style>
  .blob {
    position: absolute;
    pointer-events: none;
  }

  .content {
    position: relative;
    z-index: 1;
    height: 100%;
    box-sizing: border-box;
  }

  .main {
    display: flex;
    gap: 16px;
    height: 100%;
    min-height: 0;
  }

  .left {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .exit-row {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .country {
    border-radius: 50%;
    padding: 0;
    flex-shrink: 0;
  }

  .country.spinning :global(svg) {
    animation: spin 1.5s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .country-code {
    font-size: 13px;
    font-weight: 600;
    letter-spacing: 0.04em;
    color: var(--on-surface-variant);
  }

  .ip-col {
    min-width: 0;
  }

  .ip {
    font-size: 28px;
    line-height: 34px;
    font-weight: 700;
    letter-spacing: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--on-surface);
  }

  .chips {
    margin-top: 4px;
    display: flex;
    gap: 6px;
    height: 24px;
    overflow: hidden;
  }

  .spacer {
    flex: 1;
  }

  .cells {
    display: flex;
    gap: 28px;
  }

  .cell-label {
    font-size: 12px;
    line-height: 16px;
    color: var(--on-surface-variant);
    margin-bottom: 2px;
  }

  .cell-value {
    font-size: 15px;
    line-height: 20px;
    font-weight: 600;
  }

  .button-slot {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .start-toggle {
    position: relative;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: transform 0.2s cubic-bezier(0.2, 0.8, 0.3, 1);
    cursor: pointer;
  }

  .start-toggle:hover {
    transform: scale(1.05);
  }

  .start-toggle:active {
    transform: scale(0.96);
  }

  .start-toggle.running {
    box-shadow: 0 6px 14px color-mix(in srgb, var(--primary) 30%, transparent);
  }

  .glyph {
    position: relative;
    z-index: 1;
    display: inline-flex;
    color: var(--on-surface-variant);
  }

  .start-toggle.running .glyph {
    color: var(--on-primary);
  }
</style>
