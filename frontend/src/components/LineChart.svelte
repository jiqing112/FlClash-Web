<script lang="ts">
  // Smooth traffic sparkline, matching the dashboard speed tiles.
  let {
    values,
    color = 'var(--primary)',
    height = 56,
    fill = true,
    strokeWidth = 2,
  }: {
    values: number[];
    color?: string;
    height?: number;
    fill?: boolean;
    strokeWidth?: number;
  } = $props();

  const width = 200;
  const id = $props.id();

  function buildPath(): { line: string; area: string } {
    const data = values;
    const points = 60;
    const max = Math.max(...data, 1);
    const step = width / (points - 1);
    const sample = (index: number): number => {
      if (data.length === 0) return 0;
      const scaled = (index / (points - 1)) * (data.length - 1);
      const low = Math.floor(scaled);
      const high = Math.ceil(scaled);
      const frac = scaled - low;
      return data[low] * (1 - frac) + data[high] * frac;
    };
    // smooth polyline with quadratic midpoints
    const coords: [number, number][] = [];
    for (let i = 0; i < points; i++) {
      const value = sample(i);
      const y = height - 3 - (value / max) * (height - 8);
      coords.push([i * step, Math.max(2, y)]);
    }
    let line = `M ${coords[0][0]},${coords[0][1]}`;
    for (let i = 1; i < coords.length - 1; i++) {
      const [x, y] = coords[i];
      const [nx, ny] = coords[i + 1];
      line += ` Q ${x},${y} ${(x + nx) / 2},${(y + ny) / 2}`;
    }
    const area = `${line} L ${width},${height} L 0,${height} Z`;
    return { line, area };
  }

  const paths = $derived(buildPath());
</script>

<svg viewBox="0 0 {width} {height}" preserveAspectRatio="none" style="width:100%;height:{height}px;display:block">
  {#if fill}
    <defs>
      <linearGradient id="grad-{id}" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color={color} stop-opacity="0.28" />
        <stop offset="100%" stop-color={color} stop-opacity="0.02" />
      </linearGradient>
    </defs>
    <path d={paths.area} fill="url(#grad-{id})" />
  {/if}
  <path d={paths.line} fill="none" stroke={color} stroke-width={strokeWidth} stroke-linecap="round" />
</svg>
