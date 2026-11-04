// M3 Expressive shapes (MaterialShapes in prism_ui), as SVG paths on a
// unit-square: the cookie N-sided gears, the 4-leaf clover and the circle.

function polar(cx: number, cy: number, r: number, angle: number): [number, number] {
  return [cx + r * Math.cos(angle), cy + r * Math.sin(angle)];
}

// A rounded gear: n teeth whose radius wobbles between rOuter and rInner with
// smooth cosine interpolation, drawn as a dense polygon (visually identical to
// the material RoundedPolygon at UI sizes).
function cookie(points: number, depth: number, phase = 0): string {
  const steps = points * 24;
  const coords: [number, number][] = [];
  for (let i = 0; i <= steps; i++) {
    const t = (i / steps) * Math.PI * 2;
    const wobble = Math.cos(t * points + phase);
    const r = 0.5 - depth * 0.5 * (0.5 - 0.5 * wobble) * 2 * 0.5;
    // radius between 0.5*(1-depth) and 0.5, smooth
    const radius = 0.5 * (1 - depth * (0.5 - 0.5 * wobble));
    coords.push(polar(0.5, 0.5, radius, t - Math.PI / 2));
  }
  return (
    coords
      .map(([x, y], index) => `${index === 0 ? 'M' : 'L'}${x.toFixed(4)},${y.toFixed(4)}`)
      .join(' ') + ' Z'
  );
}

function circle(): string {
  // two arcs
  return 'M0.5,0 A0.5,0.5 0 1,1 0.4999,0 Z M0.5,1 A0.5,0.5 0 1,1 0.5001,0 Z';
}

function clover4(): string {
  // four lobes pinched at the diagonals
  const steps = 200;
  const coords: [number, number][] = [];
  for (let i = 0; i <= steps; i++) {
    const t = (i / steps) * Math.PI * 2;
    const wobble = Math.abs(Math.cos(t * 2));
    const radius = 0.5 * (0.72 + 0.28 * wobble);
    coords.push(polar(0.5, 0.5, radius, t - Math.PI / 2));
  }
  return (
    coords
      .map(([x, y], index) => `${index === 0 ? 'M' : 'L'}${x.toFixed(4)},${y.toFixed(4)}`)
      .join(' ') + ' Z'
  );
}

export const materialShapes = {
  circle: circle(),
  cookie6Sided: cookie(6, 0.16),
  cookie9Sided: cookie(9, 0.13),
  cookie12Sided: cookie(12, 0.1),
  clover4Leaf: clover4(),
};

export type MaterialShapeName = keyof typeof materialShapes;
