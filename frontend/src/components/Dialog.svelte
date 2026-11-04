<script lang="ts">
  // Modal dialog matching the app's alert style.
  import type { Snippet } from 'svelte';

  let {
    open,
    title,
    children,
    actions = [] as { label: string; onClick: () => void; primary?: boolean; destructive?: boolean }[],
    width = 400,
  }: {
    open: boolean;
    title: string;
    children?: Snippet;
    actions?: { label: string; onClick: () => void; primary?: boolean; destructive?: boolean }[];
    width?: number;
  } = $props();
</script>

{#if open}
  <div class="scrim" role="presentation"></div>
  <div class="dialog-layer">
    <div class="dialog rise-in" style="width: {width}px; max-width: calc(100vw - 48px);">
      <h2>{title}</h2>
      <div class="content">{@render children?.()}</div>
      {#if actions.length}
        <div class="dialog-actions">
          {#each actions as action}
            <button
              class="dialog-button"
              class:primary={action.primary}
              class:destructive={action.destructive}
              onclick={action.onClick}
            >
              {action.label}
            </button>
          {/each}
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .dialog-layer {
    position: fixed;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .dialog {
    background: var(--surface-bright);
    border-radius: var(--radius-alert);
    box-shadow: var(--shadow-e2);
    padding: 24px;
  }

  :root.dark .dialog {
    background: var(--surface-container-high);
  }

  .dialog h2 {
    margin: 0 0 12px;
    font-size: 1.25rem;
    font-weight: 700;
  }

  .content {
    color: var(--on-surface-variant);
    font-size: 0.95rem;
    line-height: 1.5;
  }

  .dialog-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 22px;
  }

  .dialog-button {
    height: 38px;
    padding: 0 18px;
    border-radius: var(--radius-full);
    font-weight: 590;
    color: var(--primary);
    transition: background 0.15s ease;
  }

  .dialog-button:hover {
    background: color-mix(in srgb, var(--primary) 10%, transparent);
  }

  .dialog-button.destructive {
    color: var(--error);
  }
</style>
