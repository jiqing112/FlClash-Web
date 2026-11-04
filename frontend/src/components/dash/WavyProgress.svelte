<script lang="ts">
  // The M3 Expressive wavy linear indicator: wave up to value, a gap, then a
  // flat track ending in a stop dot (port of _WavyProgressPainter).
  let {
    value,
    width = 240,
    warn = false,
  }: {
    value: number;
    width?: number;
    warn?: boolean;
  } = $props();

  const stroke = 4;
  const amplitude = 3;
  const wavelength = 12;
  const gap = 6;
  const stop = 4;

  function wavePath(w: number, v: number): string {
    const mid = 6;
    const start = stroke / 2;
    const end = w - stroke / 2;
    const split = start + (end - start) * v;
    if (v <= 0) return '';
    let d = `M${start},${mid}`;
    for (let x = start; x <= split; x += 1) {
      const y = mid + amplitude * Math.sin(((x - start) / wavelength) * Math.PI * 2);
      d += ` L${x},${y.toFixed(2)}`;
    }
    return d;
  }

  const split = $derived.by(() => {
    const start = stroke / 2;
    const end = width - stroke / 2;
    return start + (end - start) * value;
  });
</script>

<svg viewBox="0 0 {width} 12" style="width:100%;height:12px;display:block">
  {#if value > 0}
    <path d={wavePath(width, value)} fill="none" stroke={warn ? 'var(--warning)' : 'var(--primary)'} stroke-width={stroke} stroke-linecap="round" />
  {/if}
  {#if split + gap + stroke < width - stroke / 2 - stop}
    <line x1={split + gap + stroke} y1="6" x2={width - stroke / 2 - stop - gap} y2="6" stroke="var(--surface-container-high)" stroke-width={stroke} stroke-linecap="round" />
  {/if}
  {#if value < 1}
    <circle cx={width - stroke / 2 - stop / 2} cy="6" r={stop / 2} fill={warn ? 'var(--warning)' : 'var(--primary)'} />
  {/if}
</svg>
