<script lang="ts">
  // DashboardChip: h24 pill; idle = tone bg, ok/warn = containers; dot option.
  let {
    label,
    tone = 'idle',
    dot = false,
    large = false,
  }: {
    label: string;
    tone?: 'idle' | 'ok' | 'warn';
    dot?: boolean;
    large?: boolean;
  } = $props();
</script>

<span class="chip-item" class:large data-tone={tone}>
  {#if dot}<span class="dot"></span>{/if}
  {label}
</span>

<style>
  .chip-item {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 24px;
    padding: 0 10px;
    border-radius: var(--radius-full);
    font-size: 12px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
    background: var(--surface-container-high);
    color: var(--on-surface-variant);
  }

  .chip-item.large {
    height: 28px;
    padding: 0 12px;
  }

  .chip-item[data-tone='ok'] {
    background: var(--success-container);
    color: var(--on-success-container);
  }

  .chip-item[data-tone='warn'] {
    background: var(--warning-container);
    color: var(--on-warning-container);
  }

  .dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
  }
</style>
