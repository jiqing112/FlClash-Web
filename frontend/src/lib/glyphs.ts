/**
 * Glyph icon system ported from FlClash (Flutter) to TypeScript/SVG.
 *
 * Sources:
 * - reference/FlClash/lib/ui/icons/app_glyphs.dart          (AppGlyphs definitions)
 * - reference/FlClash/plugins/prism_ui/lib/icons/glyph.dart  (shape semantics)
 * - reference/FlClash/plugins/prism_ui/lib/icons/glyphs.dart (base `Glyphs` definitions)
 * - reference/FlClash/plugins/prism_ui/lib/icons/glyph_painter.dart (paint widths per role)
 *
 * Rendering semantics replicated from the Dart painter at rest (fill = 0):
 * - 24x24 grid, round caps and round joins, stroke centered on the geometry.
 * - Non-solid shapes are stroked; `solid` shapes are filled.
 * - Stroke widths per role at rest: body = 1.8, detail = 1.8, line = 1.8.
 *   (glyph_painter.dart strokes every role with `Glyph.strokeWidth` = 1.8 at
 *   fill = 0. The detail role only becomes thinner — down to 0, then a cutout —
 *   and the line role only becomes heavier — up to 2.2 — during the fill
 *   animation; see GLYPH_FILLED_LINE_WIDTH / GLYPH_DETAIL_SWAP below.)
 * - Angles: 0 deg points along +x, positive sweep goes clockwise on screen
 *   (Flutter's y-down coordinate system, same as SVG).
 *
 * Approximations (the only two constructs SVG cannot express exactly):
 * - `GlyphBox` uses `addRSuperellipse` (squircle). SVG has no superellipse
 *   primitive, so boxes render as plain rounded rectangles with circular arc
 *   corners — visually indistinguishable at icon sizes.
 * - Everything else (arcs, quadratic corner rounding, link insets, path data)
 *   is converted exactly.
 */

/** Visual role of a shape; mirrors Dart `GlyphRole`. */
export type GlyphRole = 'body' | 'detail' | 'line';

/** Polyline vertex with an optional corner rounding radius. */
export interface PolyVertex {
  x: number;
  y: number;
  r?: number;
}

/** Rounded rectangle: left, top, right, bottom, corner radius. */
export interface BoxShape {
  kind: 'box';
  l: number;
  t: number;
  r: number;
  b: number;
  radius: number;
  role?: GlyphRole;
}

/** Circle centered at (x, y); optionally filled when `solid`. */
export interface CircleShape {
  kind: 'circle';
  x: number;
  y: number;
  r: number;
  role?: GlyphRole;
  solid?: boolean;
}

/** Ellipse inscribed in the (l, t, r, b) bounds. */
export interface OvalShape {
  kind: 'oval';
  l: number;
  t: number;
  r: number;
  b: number;
  role?: GlyphRole;
}

/** Circular arc; angles in degrees (0 = +x axis, positive = clockwise). */
export interface ArcShape {
  kind: 'arc';
  x: number;
  y: number;
  r: number;
  start: number;
  sweep: number;
  role?: GlyphRole;
}

/** Line whose round caps are inset so they never poke into hollow nodes. */
export interface LinkShape {
  kind: 'link';
  x1: number;
  y1: number;
  x2: number;
  y2: number;
  radius: number;
  role: 'line';
}

/** Raw SVG path data on the 24 grid, tracing the stroke center. */
export interface PathShape {
  kind: 'path';
  d: string;
  role?: GlyphRole;
  solid?: boolean;
}

/** Polyline (open unless `closed`), with per-vertex rounded corners. */
export interface PolylineShape {
  kind: 'polyline';
  points: PolyVertex[];
  closed?: boolean;
  role?: GlyphRole;
}

export type GlyphShape =
  | BoxShape
  | CircleShape
  | OvalShape
  | ArcShape
  | LinkShape
  | PathShape
  | PolylineShape;

/** A glyph: the list of shapes it is drawn from. */
export type GlyphDef = readonly GlyphShape[];

// ---------------------------------------------------------------------------
// Shape constructors (mirror the Dart Glyph* classes)
// ---------------------------------------------------------------------------

export function box(
  l: number,
  t: number,
  r: number,
  b: number,
  radius: number,
  role?: GlyphRole,
): BoxShape {
  return { kind: 'box', l, t, r, b, radius, role };
}

export interface CircleOptions {
  solid?: boolean;
  role?: GlyphRole;
}

export function circle(x: number, y: number, r: number, opts: CircleOptions = {}): CircleShape {
  return { kind: 'circle', x, y, r, role: opts.role, solid: opts.solid };
}

export function oval(l: number, t: number, r: number, b: number, role?: GlyphRole): OvalShape {
  return { kind: 'oval', l, t, r, b, role };
}

/** Angles in degrees: 0 = +x axis, positive sweep = clockwise on screen. */
export function arc(
  x: number,
  y: number,
  r: number,
  start: number,
  sweep: number,
  role?: GlyphRole,
): ArcShape {
  return { kind: 'arc', x, y, r, start, sweep, role };
}

export function link(x1: number, y1: number, x2: number, y2: number, radius: number): LinkShape {
  return { kind: 'link', x1, y1, x2, y2, radius, role: 'line' };
}

export function path(d: string, role?: GlyphRole, solid?: boolean): PathShape {
  return { kind: 'path', d, role, solid };
}

export function polyline(
  points: PolyVertex[],
  closed?: boolean,
  role?: GlyphRole,
): PolylineShape {
  return { kind: 'polyline', points, closed, role };
}

/** Polyline vertex shorthand. */
export function v(x: number, y: number, r?: number): PolyVertex {
  return r === undefined ? { x, y } : { x, y, r };
}

// ---------------------------------------------------------------------------
// Painter constants (glyph.dart / glyph_painter.dart)
// ---------------------------------------------------------------------------

/** `Glyph.size` — the square grid every glyph is drawn on. */
export const GLYPH_SIZE = 24;
/** `Glyph.strokeWidth` — stroke width of every stroked role at rest. */
export const GLYPH_STROKE_WIDTH = 1.8;
/** `_filledLineWidth` — what the line role grows to at fill = 1. */
export const GLYPH_FILLED_LINE_WIDTH = 2.2;
/** `_detailSwap` — fill level where details become cutouts instead of strokes. */
export const GLYPH_DETAIL_SWAP = 0.5;

// ---------------------------------------------------------------------------
// Shape -> SVG path conversion
// ---------------------------------------------------------------------------

/** Round to 4 decimals and drop trailing zeros for compact `d` strings. */
function f(n: number): string {
  return String(Math.round(n * 10000) / 10000);
}

/** Full ellipse as two arcs (used for circle and oval). */
function ellipseD(cx: number, cy: number, rx: number, ry: number): string {
  return (
    `M ${f(cx - rx)} ${f(cy)}` +
    ` A ${f(rx)} ${f(ry)} 0 1 1 ${f(cx + rx)} ${f(cy)}` +
    ` A ${f(rx)} ${f(ry)} 0 1 1 ${f(cx - rx)} ${f(cy)} Z`
  );
}

function boxD(s: BoxShape): string {
  const { l, t, r, b } = s;
  // Superellipse in Dart; a plain rounded rect is the SVG stand-in.
  const rad = Math.max(0, Math.min(s.radius, (r - l) / 2, (b - t) / 2));
  return (
    `M ${f(l + rad)} ${f(t)}` +
    ` L ${f(r - rad)} ${f(t)} A ${f(rad)} ${f(rad)} 0 0 1 ${f(r)} ${f(t + rad)}` +
    ` L ${f(r)} ${f(b - rad)} A ${f(rad)} ${f(rad)} 0 0 1 ${f(r - rad)} ${f(b)}` +
    ` L ${f(l + rad)} ${f(b)} A ${f(rad)} ${f(rad)} 0 0 1 ${f(l)} ${f(b - rad)}` +
    ` L ${f(l)} ${f(t + rad)} A ${f(rad)} ${f(rad)} 0 0 1 ${f(l + rad)} ${f(t)} Z`
  );
}

function arcD(s: ArcShape): string {
  const rad = (s.start * Math.PI) / 180;
  const a1 = rad;
  const a2 = rad + (s.sweep * Math.PI) / 180;
  const x1 = s.x + s.r * Math.cos(a1);
  const y1 = s.y + s.r * Math.sin(a1);
  const x2 = s.x + s.r * Math.cos(a2);
  const y2 = s.y + s.r * Math.sin(a2);
  const largeArc = Math.abs(s.sweep) > 180 ? 1 : 0;
  // Positive sweep runs clockwise on screen (y-down), which is sweep-flag 1.
  const sweepFlag = s.sweep >= 0 ? 1 : 0;
  return `M ${f(x1)} ${f(y1)} A ${f(s.r)} ${f(s.r)} 0 ${largeArc} ${sweepFlag} ${f(x2)} ${f(y2)}`;
}

function linkD(s: LinkShape): string {
  const dx = s.x2 - s.x1;
  const dy = s.y2 - s.y1;
  const distance = Math.hypot(dx, dy);
  // Inset the round caps by (radius + strokeWidth) along the line so a cap
  // never pokes into a hollow node — same as Dart `GlyphLink`.
  const inset = (s.radius + GLYPH_STROKE_WIDTH) / distance;
  const sx = s.x1 + dx * inset;
  const sy = s.y1 + dy * inset;
  const ex = s.x2 - dx * inset;
  const ey = s.y2 - dy * inset;
  return `M ${f(sx)} ${f(sy)} L ${f(ex)} ${f(ey)}`;
}

function polylineD(s: PolylineShape): string {
  const points = s.points;
  const closed = s.closed ?? false;
  const count = points.length;
  if (count === 0) return '';
  if (count === 1) return closed ? '' : `M ${f(points[0].x)} ${f(points[0].y)}`;

  // Rounded corner exactly like Dart `GlyphPolyline._corner`: inset
  // min(radius, dist/2) along both adjacent edges, then a quadratic bezier
  // through the vertex.
  const corner = (index: number): { enter: [number, number]; exit: [number, number] } => {
    const vertex = points[index];
    const previous = points[(index - 1 + count) % count];
    const next = points[(index + 1) % count];
    const toPreviousX = previous.x - vertex.x;
    const toPreviousY = previous.y - vertex.y;
    const toNextX = next.x - vertex.x;
    const toNextY = next.y - vertex.y;
    const dPrev = Math.hypot(toPreviousX, toPreviousY);
    const dNext = Math.hypot(toNextX, toNextY);
    const radius = vertex.r ?? 0;
    const kPrev = dPrev === 0 ? 0 : Math.min(radius, dPrev / 2) / dPrev;
    const kNext = dNext === 0 ? 0 : Math.min(radius, dNext / 2) / dNext;
    return {
      enter: [vertex.x + toPreviousX * kPrev, vertex.y + toPreviousY * kPrev],
      exit: [vertex.x + toNextX * kNext, vertex.y + toNextY * kNext],
    };
  };

  const parts: string[] = [];
  const first = closed ? corner(0).exit : ([points[0].x, points[0].y] as [number, number]);
  parts.push(`M ${f(first[0])} ${f(first[1])}`);
  const last = closed ? count : count - 1;
  for (let index = 1; index <= last; index++) {
    const vertex = points[index % count];
    if (!closed && index === count - 1) {
      // The end vertex of an open polyline is never rounded.
      parts.push(`L ${f(vertex.x)} ${f(vertex.y)}`);
      break;
    }
    const { enter, exit } = corner(index % count);
    parts.push(
      `L ${f(enter[0])} ${f(enter[1])}` +
        ` Q ${f(vertex.x)} ${f(vertex.y)} ${f(exit[0])} ${f(exit[1])}`,
    );
  }
  if (closed) parts.push('Z');
  return parts.join(' ');
}

function shapeD(shape: GlyphShape): string {
  switch (shape.kind) {
    case 'box':
      return boxD(shape);
    case 'circle':
      return ellipseD(shape.x, shape.y, shape.r, shape.r);
    case 'oval':
      return ellipseD(
        (shape.l + shape.r) / 2,
        (shape.t + shape.b) / 2,
        (shape.r - shape.l) / 2,
        (shape.b - shape.t) / 2,
      );
    case 'arc':
      return arcD(shape);
    case 'link':
      return linkD(shape);
    case 'path':
      return shape.d;
    case 'polyline':
      return polylineD(shape);
  }
}

/** SVG `d` strings grouped the way the Dart painter paints them. */
export interface GlyphSvgPaths {
  /** Non-solid body shapes, stroked at GLYPH_STROKE_WIDTH. */
  body: string;
  /** Non-solid detail shapes; same stroke width as body at rest. */
  detail: string;
  /** Non-solid line shapes; 1.8 at rest, up to 2.2 when a fill animates. */
  line: string;
  /** Solid (filled) shapes, in paint order; render with fill and no stroke. */
  solid: string[];
}

/**
 * Convert a glyph definition to SVG path `d` strings grouped by paint role,
 * so a component can render one `<path>` per group with the right paint.
 */
export function glyphSvgPaths(def: GlyphDef): GlyphSvgPaths {
  const body: string[] = [];
  const detail: string[] = [];
  const line: string[] = [];
  const solid: string[] = [];
  for (const shape of def) {
    const d = shapeD(shape);
    if ('solid' in shape && shape.solid) {
      solid.push(d);
    } else if (shape.role === 'detail') {
      detail.push(d);
    } else if (shape.role === 'line') {
      line.push(d);
    } else {
      body.push(d);
    }
  }
  return {
    body: body.join(' '),
    detail: detail.join(' '),
    line: line.join(' '),
    solid,
  };
}

// ---------------------------------------------------------------------------
// Glyph definitions
// Order follows AppGlyphs in app_glyphs.dart; the glyphs Dart borrows from
// the base `Glyphs` class (arrowBack, close, check, ...) are defined inline
// at their alias position. `clock` aliases `requests` like in Dart.
// ---------------------------------------------------------------------------

const dashboard: GlyphDef = [
  box(3.5, 3.5, 10.5, 11.5, 2),
  box(13.5, 3.5, 20.5, 9.5, 2),
  box(3.5, 14.5, 10.5, 20.5, 2),
  box(13.5, 12.5, 20.5, 20.5, 2),
];

const proxies: GlyphDef = [
  circle(12, 12, 8.8),
  oval(8.2, 3.2, 15.8, 20.8, 'detail'),
  polyline([v(3.2, 12), v(20.8, 12)], undefined, 'detail'),
];

const profiles: GlyphDef = [
  polyline([v(5, 3, 2.5), v(13.5, 3, 1), v(19, 8.5, 1), v(19, 21, 2.5), v(5, 21, 2.5)], true),
  polyline([v(13.5, 3.2), v(13.5, 8.5, 1.5), v(18.8, 8.5)], undefined, 'detail'),
  polyline([v(8.5, 12.5), v(15.5, 12.5)], undefined, 'detail'),
  polyline([v(8.5, 16), v(13, 16)], undefined, 'detail'),
];

const requests: GlyphDef = [
  circle(12, 12, 8.8),
  polyline([v(12, 7.2), v(12, 12, 0.6), v(15.2, 14)], undefined, 'detail'),
];

const connections: GlyphDef = [
  circle(12, 5.2, 2.4),
  circle(5.2, 17.6, 2.4),
  circle(18.8, 17.6, 2.4),
  link(12, 5.2, 5.2, 17.6, 2.4),
  link(12, 5.2, 18.8, 17.6, 2.4),
  link(5.2, 17.6, 18.8, 17.6, 2.4),
];

const dns: GlyphDef = [
  box(3.5, 3.5, 20.5, 10.5, 2.5),
  box(3.5, 13.5, 20.5, 20.5, 2.5),
  circle(7.2, 7, 1.1, { role: 'detail', solid: true }),
  circle(7.2, 17, 1.1, { role: 'detail', solid: true }),
];

const resources: GlyphDef = [
  box(3, 3.5, 21, 8.5, 1.8),
  box(4.5, 10.8, 19.5, 20.5, 2.2),
  polyline([v(10, 14.2), v(14, 14.2)], undefined, 'detail'),
];

const logs: GlyphDef = [
  box(3, 4.5, 21, 19.5, 3.5),
  polyline([v(7.3, 9.3), v(9.9, 11.9, 0.4), v(7.3, 14.5)], undefined, 'detail'),
  polyline([v(12.2, 14.5), v(16.2, 14.5)], undefined, 'detail'),
];

const tools: GlyphDef = [
  box(3, 7, 21, 20, 2.8),
  polyline([v(9, 7), v(9, 4.6, 1.4), v(15, 4.6, 1.4), v(15, 7)], undefined, 'line'),
  polyline([v(3.2, 12.6), v(9.6, 12.6)], undefined, 'detail'),
  polyline([v(14.4, 12.6), v(20.8, 12.6)], undefined, 'detail'),
  box(10.4, 11, 13.6, 14.4, 1, 'detail'),
];

const target: GlyphDef = [
  // Dart: pi * -0.93 / pi * 0.86 radians -> degrees below.
  arc(12, 12, 8.8, 180 * -0.93, 180 * 0.86, 'line'),
  arc(12, 12, 8.8, 180 * 0.07, 180 * 0.86, 'line'),
  circle(12, 12, 2),
];

const arrowBack: GlyphDef = [
  polyline([v(5, 12), v(19, 12)], undefined, 'line'),
  polyline([v(11.5, 5.5), v(5, 12, 0.4), v(11.5, 18.5)], undefined, 'line'),
];

const chevronBack: GlyphDef = [
  polyline([v(15.25, 5.5), v(8.75, 12, 0.4), v(15.25, 18.5)], undefined, 'line'),
];

const chevronForward: GlyphDef = [
  polyline([v(8.75, 5.5), v(15.25, 12, 0.4), v(8.75, 18.5)], undefined, 'line'),
];

const chevronDown: GlyphDef = [
  polyline([v(5.5, 8.75), v(12, 15.25, 0.4), v(18.5, 8.75)], undefined, 'line'),
];

const chevronUp: GlyphDef = [
  polyline([v(5.5, 15.25), v(12, 8.75, 0.4), v(18.5, 15.25)], undefined, 'line'),
];

const picker: GlyphDef = [
  polyline([v(7.5, 9.5), v(12, 5, 0.4), v(16.5, 9.5)], undefined, 'line'),
  polyline([v(7.5, 14.5), v(12, 19, 0.4), v(16.5, 14.5)], undefined, 'line'),
];

const caretDown: GlyphDef = [polyline([v(5.2, 8, 1), v(18.8, 8, 1), v(12, 16, 1)], true)];

const arrowUp: GlyphDef = [
  polyline([v(12, 5), v(12, 19)], undefined, 'line'),
  polyline([v(5.5, 11.5), v(12, 5, 0.4), v(18.5, 11.5)], undefined, 'line'),
];

const arrowDown: GlyphDef = [
  polyline([v(12, 5), v(12, 19)], undefined, 'line'),
  polyline([v(5.5, 12.5), v(12, 19, 0.4), v(18.5, 12.5)], undefined, 'line'),
];

const openExternal: GlyphDef = [
  box(3, 3, 21, 21, 3),
  polyline([v(7.6, 16.4), v(16.4, 7.6)], undefined, 'detail'),
  polyline([v(11, 7.6), v(16.4, 7.6, 0.4), v(16.4, 13)], undefined, 'detail'),
];

const undo: GlyphDef = [
  polyline([v(5, 9), v(19, 9, 4.8), v(19, 18.6, 4.8), v(9, 18.6)], undefined, 'line'),
  polyline([v(8.8, 5.2), v(5, 9, 0.4), v(8.8, 12.8)], undefined, 'line'),
];

const redo: GlyphDef = [
  polyline([v(19, 9), v(5, 9, 4.8), v(5, 18.6, 4.8), v(15, 18.6)], undefined, 'line'),
  polyline([v(15.2, 5.2), v(19, 9, 0.4), v(15.2, 12.8)], undefined, 'line'),
];

const sync: GlyphDef = [
  arc(12, 12, 7.78, 180 * 1.144, 180 * 0.889, 'line'),
  polyline([v(5, 5), v(5, 8.6, 0.4), v(8.6, 8.6)], undefined, 'line'),
  arc(12, 12, 7.78, 180 * 0.144, 180 * 0.889, 'line'),
  polyline([v(19, 19), v(19, 15.4, 0.4), v(15.4, 15.4)], undefined, 'line'),
];

const refresh: GlyphDef = [
  arc(12, 12, 7.78, 180 * 0.19, 180 * 1.666, 'line'),
  polyline([v(19, 5), v(19, 8.6, 0.4), v(15.4, 8.6)], undefined, 'line'),
];

const reset: GlyphDef = [
  arc(12, 13.77, 7.07, 180 * 1.15, 180 * -1.65, 'line'),
  polyline([v(12, 6.7), v(9.9, 6.7)], undefined, 'line'),
  polyline([v(13.7, 10.5), v(9.9, 6.7, 0.4), v(13.7, 2.9)], undefined, 'line'),
];

const exportGlyph: GlyphDef = [
  polyline(
    [v(8.4, 10.5), v(5, 10.5, 1.5), v(5, 21, 2.8), v(19, 21, 2.8), v(19, 10.5, 1.5), v(15.6, 10.5)],
    undefined,
    'line',
  ),
  polyline([v(12, 15.6), v(12, 3)], undefined, 'line'),
  polyline([v(8.2, 6.8), v(12, 3, 0.4), v(15.8, 6.8)], undefined, 'line'),
];

const upload: GlyphDef = [
  polyline([v(4.5, 16.4), v(4.5, 20, 2.2), v(19.5, 20, 2.2), v(19.5, 16.4)], undefined, 'line'),
  polyline([v(12, 15.4), v(12, 3.5)], undefined, 'line'),
  polyline([v(8.2, 7.3), v(12, 3.5, 0.4), v(15.8, 7.3)], undefined, 'line'),
];

const add: GlyphDef = [
  polyline([v(12, 5), v(12, 19)], undefined, 'line'),
  polyline([v(5, 12), v(19, 12)], undefined, 'line'),
];

const remove: GlyphDef = [polyline([v(5, 12), v(19, 12)], undefined, 'line')];

const addCircle: GlyphDef = [
  circle(12, 12, 8.8, { role: 'line' }),
  polyline([v(12, 7.8), v(12, 16.2)], undefined, 'line'),
  polyline([v(7.8, 12), v(16.2, 12)], undefined, 'line'),
];

const removeCircle: GlyphDef = [
  circle(12, 12, 8.8, { role: 'line' }),
  polyline([v(7.8, 12), v(16.2, 12)], undefined, 'line'),
];

const close: GlyphDef = [
  polyline([v(6, 6), v(18, 18)], undefined, 'line'),
  polyline([v(18, 6), v(6, 18)], undefined, 'line'),
];

const clearField: GlyphDef = [
  circle(12, 12, 8.8),
  polyline([v(9.3, 9.3), v(14.7, 14.7)], undefined, 'detail'),
  polyline([v(14.7, 9.3), v(9.3, 14.7)], undefined, 'detail'),
];

const check: GlyphDef = [polyline([v(4.8, 12.5), v(9.6, 17.3), v(19.4, 6.7)], undefined, 'line')];

const checkDouble: GlyphDef = [
  polyline([v(2.9, 12.5), v(6.7, 16.3), v(14.45, 7.9)], undefined, 'line'),
  polyline([v(11.8, 16.3), v(13.2, 17.7), v(20.95, 9.3)], undefined, 'line'),
];

const checkCircle: GlyphDef = [
  circle(12, 12, 8.8),
  polyline([v(7.59, 12.3), v(10.49, 15.2), v(16.41, 8.79)], undefined, 'detail'),
];

const block: GlyphDef = [
  circle(12, 12, 8.8, { role: 'line' }),
  polyline([v(5.78, 5.78), v(18.22, 18.22)], undefined, 'line'),
];

const info: GlyphDef = [
  circle(12, 12, 8.8),
  circle(12, 7.5, 1.4, { role: 'detail', solid: true }),
  polyline([v(12, 11.8), v(12, 16.7)], undefined, 'detail'),
];

const warning: GlyphDef = [
  polyline([v(12, 1.6, 3), v(21.9, 20.3, 2.5), v(2.1, 20.3, 2.5)], true),
  polyline([v(12, 9), v(12, 12.1)], undefined, 'detail'),
  circle(12, 16.2, 1.4, { role: 'detail', solid: true }),
];

const error: GlyphDef = [
  circle(12, 12, 8.8),
  polyline([v(12, 7.3), v(12, 12.2)], undefined, 'detail'),
  circle(12, 16.5, 1.4, { role: 'detail', solid: true }),
];

const more: GlyphDef = [
  circle(12, 4.2, 1.65, { role: 'line', solid: true }),
  circle(12, 12, 1.65, { role: 'line', solid: true }),
  circle(12, 19.8, 1.65, { role: 'line', solid: true }),
];

const moreCircle: GlyphDef = [
  circle(12, 12, 8.8),
  circle(7.4, 12, 1.4, { role: 'detail', solid: true }),
  circle(12, 12, 1.4, { role: 'detail', solid: true }),
  circle(16.6, 12, 1.4, { role: 'detail', solid: true }),
];

const dragHandle: GlyphDef = [
  polyline([v(4, 7), v(20, 7)], undefined, 'line'),
  polyline([v(4, 12), v(20, 12)], undefined, 'line'),
  polyline([v(4, 17), v(20, 17)], undefined, 'line'),
];

const sort: GlyphDef = [
  polyline([v(4, 7), v(20, 7)], undefined, 'line'),
  polyline([v(4, 12), v(15, 12)], undefined, 'line'),
  polyline([v(4, 17), v(10, 17)], undefined, 'line'),
];

const filter: GlyphDef = [
  polyline([v(4, 7), v(20, 7)], undefined, 'line'),
  polyline([v(7, 12), v(17, 12)], undefined, 'line'),
  polyline([v(10, 17), v(14, 17)], undefined, 'line'),
];

const textShort: GlyphDef = [
  polyline([v(4, 9), v(20, 9)], undefined, 'line'),
  polyline([v(4, 15), v(13, 15)], undefined, 'line'),
];

const wrap: GlyphDef = [
  polyline([v(4, 6), v(20, 6)], undefined, 'line'),
  polyline([v(4, 12), v(20, 12, 3), v(20, 18, 3), v(11, 18)], undefined, 'line'),
  polyline([v(13.8, 15.2), v(11, 18, 0.4), v(13.8, 20.8)], undefined, 'line'),
  polyline([v(4, 18), v(6.5, 18)], undefined, 'line'),
];

const listAdd: GlyphDef = [
  polyline([v(4, 4.4), v(20, 4.4)], undefined, 'line'),
  polyline([v(4, 9.2), v(20, 9.2)], undefined, 'line'),
  polyline([v(4, 14), v(10, 14)], undefined, 'line'),
  polyline([v(17, 13.4), v(17, 20.2)], undefined, 'line'),
  polyline([v(13.6, 16.8), v(20.4, 16.8)], undefined, 'line'),
];

const edit: GlyphDef = [
  polyline(
    [v(4, 20, 1), v(8.75, 18.75, 1), v(20.25, 7.25, 1.5), v(16.75, 3.75, 1.5), v(5.25, 15.25, 1)],
    true,
  ),
  polyline([v(13.25, 7.25), v(16.75, 10.75)], undefined, 'detail'),
];

const compose: GlyphDef = [
  polyline([v(9.4, 6.2), v(4, 6.2, 2.5), v(4, 20.2, 2.5), v(18, 20.2, 2.5), v(18, 14.8)], undefined, 'line'),
  polyline(
    [v(9, 15.2, 1), v(13.75, 13.95, 1), v(20.75, 6.95, 1.5), v(17.25, 3.45, 1.5), v(10.25, 10.45, 1)],
    true,
  ),
  polyline([v(13.75, 6.95), v(17.25, 10.45)], undefined, 'detail'),
];

const deleteGlyph: GlyphDef = [
  polyline([v(5.8, 6.8), v(18.2, 6.8), v(17.2, 21, 2.5), v(6.8, 21, 2.5)], true),
  polyline([v(10, 10.5), v(10, 17.5)], undefined, 'detail'),
  polyline([v(14, 10.5), v(14, 17.5)], undefined, 'detail'),
  polyline([v(3.8, 6.8), v(20.2, 6.8)], undefined, 'line'),
  polyline([v(9, 6.8), v(9, 3.2, 1.4), v(15, 3.2, 1.4), v(15, 6.8)], undefined, 'line'),
];

const clearAll: GlyphDef = [
  polyline([v(4.24, 8.6), v(13.16, 8.6), v(12.44, 18.8, 1.8), v(4.96, 18.8, 1.8)], true),
  polyline([v(8.7, 12), v(8.7, 15.3)], undefined, 'detail'),
  polyline([v(3, 8.6), v(14.4, 8.6)], undefined, 'line'),
  polyline([v(6.2, 8.6), v(6.2, 5, 1.4), v(11.2, 5, 1.4), v(11.2, 8.6)], undefined, 'line'),
  polyline([v(17, 11.2), v(21, 11.2)], undefined, 'line'),
  polyline([v(17, 14.9), v(20.4, 14.9)], undefined, 'line'),
  polyline([v(17, 18.6), v(19.8, 18.6)], undefined, 'line'),
];

const copy: GlyphDef = [
  box(8.5, 7.5, 20.5, 21, 2.5),
  polyline([v(4, 16.5), v(4, 3, 2.5), v(16, 3)], undefined, 'line'),
];

const cut: GlyphDef = [
  circle(7, 17.5, 3),
  circle(17, 17.5, 3),
  polyline([v(9.64, 13.73), v(16.8, 3.5)], undefined, 'line'),
  polyline([v(14.36, 13.73), v(7.2, 3.5)], undefined, 'line'),
];

const paste: GlyphDef = [
  polyline(
    [
      v(4.5, 5.2, 2.5),
      v(8.5, 5.2),
      v(8.5, 2.9, 1),
      v(15.5, 2.9, 1),
      v(15.5, 5.2),
      v(19.5, 5.2, 2.5),
      v(19.5, 21, 2.5),
      v(4.5, 21, 2.5),
    ],
    true,
  ),
  polyline([v(8.5, 5.2), v(8.5, 7.3, 1), v(15.5, 7.3, 1), v(15.5, 5.2)], undefined, 'detail'),
];

const selectAll: GlyphDef = [
  polyline([v(3.5, 7), v(3.5, 3.5, 2.5), v(7, 3.5)], undefined, 'line'),
  polyline([v(20.5, 7), v(20.5, 3.5, 2.5), v(17, 3.5)], undefined, 'line'),
  polyline([v(20.5, 17), v(20.5, 20.5, 2.5), v(17, 20.5)], undefined, 'line'),
  polyline([v(3.5, 17), v(3.5, 20.5, 2.5), v(7, 20.5)], undefined, 'line'),
  polyline([v(10.6, 3.5), v(13.4, 3.5)], undefined, 'line'),
  polyline([v(20.5, 10.6), v(20.5, 13.4)], undefined, 'line'),
  polyline([v(13.4, 20.5), v(10.6, 20.5)], undefined, 'line'),
  polyline([v(3.5, 13.4), v(3.5, 10.6)], undefined, 'line'),
  polyline([v(8, 12.3), v(10.8, 15.1, 0.4), v(16, 9.2)], undefined, 'line'),
];

const deselect: GlyphDef = [
  polyline([v(3.5, 7), v(3.5, 3.5, 2.5), v(7, 3.5)], undefined, 'line'),
  polyline([v(20.5, 7), v(20.5, 3.5, 2.5), v(17, 3.5)], undefined, 'line'),
  polyline([v(20.5, 17), v(20.5, 20.5, 2.5), v(17, 20.5)], undefined, 'line'),
  polyline([v(3.5, 17), v(3.5, 20.5, 2.5), v(7, 20.5)], undefined, 'line'),
  polyline([v(10.6, 3.5), v(13.4, 3.5)], undefined, 'line'),
  polyline([v(20.5, 10.6), v(20.5, 13.4)], undefined, 'line'),
  polyline([v(13.4, 20.5), v(10.6, 20.5)], undefined, 'line'),
  polyline([v(3.5, 13.4), v(3.5, 10.6)], undefined, 'line'),
  polyline([v(8, 8), v(16, 16)], undefined, 'line'),
];

const search: GlyphDef = [
  circle(10.3, 10.3, 6.9, { role: 'line' }),
  polyline([v(15.18, 15.18), v(20.6, 20.6)], undefined, 'line'),
];

const matchCase: GlyphDef = [
  polyline([v(2.4, 18.5), v(7.4, 5.5, 0.6), v(12.4, 18.5)], undefined, 'line'),
  polyline([v(4.13, 14), v(10.67, 14)], undefined, 'line'),
  circle(18.1, 15.1, 3.4, { role: 'line' }),
  polyline([v(21.5, 11.7), v(21.5, 18.5)], undefined, 'line'),
];

const regex: GlyphDef = [
  circle(5.5, 17.5, 0.9, { role: 'line' }),
  polyline([v(15.5, 4), v(15.5, 13)], undefined, 'line'),
  polyline([v(11.6, 6.25), v(19.4, 10.75)], undefined, 'line'),
  polyline([v(11.6, 10.75), v(19.4, 6.25)], undefined, 'line'),
];

const importFile: GlyphDef = [
  polyline([v(5, 3, 2.5), v(13.5, 3, 1), v(19, 8.5, 1), v(19, 21, 2.5), v(5, 21, 2.5)], true),
  polyline([v(13.5, 3.2), v(13.5, 8.5, 1.5), v(18.8, 8.5)], undefined, 'detail'),
  polyline([v(12, 17.6), v(12, 11.4)], undefined, 'detail'),
  polyline([v(9, 14.4), v(12, 11.4, 0.4), v(15, 14.4)], undefined, 'detail'),
];

const cloudDownload: GlyphDef = [
  path('M7 19.4A4.2 4.2 0 0 1 5.06 11.47A6 6 0 1 1 16.95 9.83A4.8 4.8 0 0 1 16.4 19.4Z'),
  polyline([v(12, 8.4), v(12, 16)], undefined, 'detail'),
  polyline([v(9, 13), v(12, 16, 0.4), v(15, 13)], undefined, 'detail'),
];

const cloudSync: GlyphDef = [
  path('M7 19.4A4.2 4.2 0 0 1 5.06 11.47A6 6 0 1 1 16.95 9.83A4.8 4.8 0 0 1 16.4 19.4Z'),
  arc(11, 13.2, 3, 180 * 0.1944, 180 * 1.5, 'detail'),
  polyline([v(12.72, 10.74), v(13.46, 11.26)], undefined, 'detail'),
  polyline([v(13.01, 8.7), v(13.46, 11.26, 0.4), v(10.9, 11.71)], undefined, 'detail'),
];

const linkGlyph: GlyphDef = [
  path('M11.01 8.46L14.55 4.93A3.2 3.2 0 0 1 19.07 9.45L15.54 12.99', 'line'),
  path('M12.99 15.54L9.45 19.07A3.2 3.2 0 0 1 4.93 14.55L8.46 11.01', 'line'),
  polyline([v(9.17, 14.83), v(14.83, 9.17)], undefined, 'line'),
];

const send: GlyphDef = [
  polyline([v(3.2, 10.8, 1), v(20.6, 3.4, 1), v(13.2, 20.8, 1), v(10.6, 13.4, 0.6)], true),
  polyline([v(10.6, 13.4), v(15.2, 8.8)], undefined, 'detail'),
];

const rules: GlyphDef = [
  box(3, 4, 21, 20, 3),
  circle(7, 8, 1.1, { role: 'detail', solid: true }),
  polyline([v(10.8, 8), v(17.2, 8)], undefined, 'detail'),
  circle(7, 12, 1.1, { role: 'detail', solid: true }),
  polyline([v(10.8, 12), v(17.2, 12)], undefined, 'detail'),
  circle(7, 16, 1.1, { role: 'detail', solid: true }),
  polyline([v(10.8, 16), v(17.2, 16)], undefined, 'detail'),
];

const rulePresets: GlyphDef = [
  box(3, 9.5, 21, 21, 3),
  polyline([v(5.5, 6.2), v(18.5, 6.2)], undefined, 'line'),
  polyline([v(8, 3), v(16, 3)], undefined, 'line'),
  circle(7, 13.4, 1.1, { role: 'detail', solid: true }),
  polyline([v(10.8, 13.4), v(17.2, 13.4)], undefined, 'detail'),
  circle(7, 17.1, 1.1, { role: 'detail', solid: true }),
  polyline([v(10.8, 17.1), v(17.2, 17.1)], undefined, 'detail'),
];

const code: GlyphDef = [
  polyline([v(7.5, 7.5), v(3, 12, 0.5), v(7.5, 16.5)], undefined, 'line'),
  polyline([v(16.5, 7.5), v(21, 12, 0.5), v(16.5, 16.5)], undefined, 'line'),
  polyline([v(13.8, 5), v(10.2, 19)], undefined, 'line'),
];

const settings: GlyphDef = [
  polyline(
    [
      v(9.87, 5.44, 0.6),
      v(10.09, 3, 0.9),
      v(13.91, 3, 0.9),
      v(14.13, 5.44, 0.6),
      v(16.62, 6.87, 0.6),
      v(18.84, 5.84, 0.9),
      v(20.75, 9.16, 0.9),
      v(18.75, 10.57, 0.6),
      v(18.75, 13.43, 0.6),
      v(20.75, 14.84, 0.9),
      v(18.84, 18.16, 0.9),
      v(16.62, 17.13, 0.6),
      v(14.13, 18.56, 0.6),
      v(13.91, 21, 0.9),
      v(10.09, 21, 0.9),
      v(9.87, 18.56, 0.6),
      v(7.38, 17.13, 0.6),
      v(5.16, 18.16, 0.9),
      v(3.25, 14.84, 0.9),
      v(5.25, 13.43, 0.6),
      v(5.25, 10.57, 0.6),
      v(3.25, 9.16, 0.9),
      v(5.16, 5.84, 0.9),
      v(7.38, 6.87, 0.6),
    ],
    true,
  ),
  circle(12, 12, 3, { role: 'detail' }),
];

const sliders: GlyphDef = [
  polyline([v(3.2, 5.5), v(10.4, 5.5)], undefined, 'line'),
  polyline([v(18.6, 5.5), v(20.8, 5.5)], undefined, 'line'),
  circle(14.5, 5.5, 2.5),
  polyline([v(3.2, 12), v(4.4, 12)], undefined, 'line'),
  polyline([v(12.6, 12), v(20.8, 12)], undefined, 'line'),
  circle(8.5, 12, 2.5),
  polyline([v(3.2, 18.5), v(11.4, 18.5)], undefined, 'line'),
  polyline([v(19.6, 18.5), v(20.8, 18.5)], undefined, 'line'),
  circle(15.5, 18.5, 2.5),
];

const wrench: GlyphDef = [
  path(
    'M3.58 17.59L10.1 11.07A1.2 1.2 0 0 0 10.41 9.89' +
      'A5.4 5.4 0 0 1 16.36 3.05A0.9 0.9 0 0 1 16.87 4.58L14.04 7.41' +
      'A1.8 1.8 0 0 0 16.59 9.96L19.42 7.13A0.9 0.9 0 0 1 20.95 7.64' +
      'A5.4 5.4 0 0 1 14.11 13.59A1.2 1.2 0 0 0 12.93 13.9L6.41 20.42' +
      'A2 2 0 0 1 3.58 17.59Z',
  ),
];

const toggle: GlyphDef = [
  box(2.9, 6.5, 21.1, 17.5, 5.5),
  circle(8.4, 12, 2.8, { role: 'detail', solid: true }),
];

const toggleOn: GlyphDef = [
  box(2.9, 6.5, 21.1, 17.5, 5.5),
  circle(15.6, 12, 2.8, { role: 'detail', solid: true }),
];

const pin: GlyphDef = [
  polyline(
    [
      v(15.42, 3.5, 1.5),
      v(20.5, 8.58, 1.5),
      v(15.73, 13.35, 1.4),
      v(14.94, 19.55, 1),
      v(4.45, 9.06, 1),
      v(10.65, 8.27, 1.4),
    ],
    true,
  ),
  polyline([v(9.7, 14.3), v(3.5, 20.5)], undefined, 'line'),
];

const eye: GlyphDef = [
  path(
    'M3.01 11.32A10.08 10.08 0 0 1 20.99 11.32A1.5 1.5 0 0 1 20.99 12.68' +
      'A10.08 10.08 0 0 1 3.01 12.68A1.5 1.5 0 0 1 3.01 11.32Z',
  ),
  circle(12, 12, 2.8, { role: 'detail' }),
];

const eyeOff: GlyphDef = [
  path(
    'M10.95 5.86A10.08 10.08 0 0 1 20.99 11.32A1.5 1.5 0 0 1 20.99 12.68' +
      'A10.08 10.08 0 0 1 19.71 14.62',
    'line',
  ),
  path(
    'M16.92 16.92A10.08 10.08 0 0 1 3.01 12.68A1.5 1.5 0 0 1 3.01 11.32' +
      'A10.08 10.08 0 0 1 7.08 7.08',
    'line',
  ),
  arc(12, 12, 2.8, 180 * 0.25, 180, 'line'),
  polyline([v(4, 4), v(20, 20)], undefined, 'line'),
];

const lock: GlyphDef = [
  box(5, 10.5, 19, 20.8, 2.5),
  path('M8.5 10.5V7A3.5 3.5 0 0 1 15.5 7V10.5', 'line'),
  polyline([v(12, 14.4), v(12, 17)], undefined, 'detail'),
];

const lockOpen: GlyphDef = [
  box(5, 10.5, 19, 20.8, 2.5),
  path('M8.5 10.5V7A3.5 3.5 0 0 1 15.5 7', 'line'),
  polyline([v(12, 14.4), v(12, 17)], undefined, 'detail'),
];

const key: GlyphDef = [
  circle(9.27, 7.99, 5),
  circle(9.27, 7.99, 1.4, { role: 'detail', solid: true }),
  polyline([v(12.8, 11.53), v(19.73, 18.46, 0.8), v(17.19, 21.01)], undefined, 'line'),
  polyline([v(17.19, 15.91), v(15.07, 18.04)], undefined, 'line'),
];

const password: GlyphDef = [
  box(2.9, 6, 21.1, 18, 3),
  circle(7.4, 12, 1.4, { role: 'detail', solid: true }),
  circle(12, 12, 1.4, { role: 'detail', solid: true }),
  circle(16.6, 12, 1.4, { role: 'detail', solid: true }),
];

const account: GlyphDef = [
  circle(12, 12, 8.8),
  circle(12, 9.4, 2.8, { role: 'detail' }),
  arc(12, 22.6, 7, 180 * -0.8087, 180 * 0.6173, 'detail'),
];

const sparkle: GlyphDef = [
  polyline(
    [
      v(10, 6.4, 0.9),
      v(11.48, 12.32, 3),
      v(17.4, 13.8, 0.9),
      v(11.48, 15.28, 3),
      v(10, 21.2, 0.9),
      v(8.52, 15.28, 3),
      v(2.6, 13.8, 0.9),
      v(8.52, 12.32, 3),
    ],
    true,
  ),
  polyline(
    [
      v(17.8, 2.9, 0.6),
      v(18.58, 5.42, 1.2),
      v(21.1, 6.2, 0.6),
      v(18.58, 6.98, 1.2),
      v(17.8, 9.5, 0.6),
      v(17.02, 6.98, 1.2),
      v(14.5, 6.2, 0.6),
      v(17.02, 5.42, 1.2),
    ],
    true,
  ),
];

const bolt: GlyphDef = [
  polyline(
    [v(14.5, 2.8, 1), v(5, 13.7, 1), v(11.2, 13.7, 0.6), v(9.5, 21.2, 1), v(19, 10.3, 1), v(12.8, 10.3, 0.6)],
    true,
  ),
];

const locate: GlyphDef = [
  circle(12, 12, 6.2),
  circle(12, 12, 1.8, { role: 'detail', solid: true }),
  polyline([v(12, 2.9), v(12, 5.8)], undefined, 'line'),
  polyline([v(12, 18.2), v(12, 21.1)], undefined, 'line'),
  polyline([v(2.9, 12), v(5.8, 12)], undefined, 'line'),
  polyline([v(18.2, 12), v(21.1, 12)], undefined, 'line'),
];

const star: GlyphDef = [
  polyline(
    [
      v(12, 3.2, 1.2),
      v(14.41, 9.48, 0.6),
      v(21.13, 9.83, 1.2),
      v(15.9, 14.07, 0.6),
      v(17.64, 20.57, 1.2),
      v(12, 16.9, 0.6),
      v(6.36, 20.57, 1.2),
      v(8.1, 14.07, 0.6),
      v(2.87, 9.83, 1.2),
      v(9.59, 9.48, 0.6),
    ],
    true,
  ),
];

const puzzle: GlyphDef = [
  path(
    'M3 10A2.2 2.2 0 0 1 5.2 7.8H7.94A2.6 2.6 0 1 1 11.26 7.8H14' +
      'A2.2 2.2 0 0 1 16.2 10V12.74A2.6 2.6 0 1 1 16.2 16.06V18.8' +
      'A2.2 2.2 0 0 1 14 21H5.2A2.2 2.2 0 0 1 3 18.8V17.58' +
      'A0.8 0.8 0 0 1 4.08 16.83A2.6 2.6 0 1 0 4.08 11.97' +
      'A0.8 0.8 0 0 1 3 11.22Z',
  ),
];

const customize: GlyphDef = [
  box(3.5, 3.5, 10.5, 10.5, 2),
  box(3.5, 13.5, 10.5, 20.5, 2),
  box(13.5, 13.5, 20.5, 20.5, 2),
  polyline([v(17, 4), v(17, 10)], undefined, 'line'),
  polyline([v(14, 7), v(20, 7)], undefined, 'line'),
];

const layers: GlyphDef = [
  polyline([v(12, 3, 1.5), v(20.8, 7.6, 1.2), v(12, 12.2, 1.5), v(3.2, 7.6, 1.2)], true),
  polyline([v(3.2, 12), v(12, 16.6, 1.5), v(20.8, 12)], undefined, 'line'),
  polyline([v(3.2, 16.4), v(12, 21, 1.5), v(20.8, 16.4)], undefined, 'line'),
];

const sun: GlyphDef = [
  circle(12, 12, 4),
  polyline([v(12, 2.9), v(12, 4.4)], undefined, 'line'),
  polyline([v(12, 19.6), v(12, 21.1)], undefined, 'line'),
  polyline([v(2.9, 12), v(4.4, 12)], undefined, 'line'),
  polyline([v(19.6, 12), v(21.1, 12)], undefined, 'line'),
  polyline([v(5.57, 5.57), v(6.63, 6.63)], undefined, 'line'),
  polyline([v(18.43, 5.57), v(17.37, 6.63)], undefined, 'line'),
  polyline([v(5.57, 18.43), v(6.63, 17.37)], undefined, 'line'),
  polyline([v(18.43, 18.43), v(17.37, 17.37)], undefined, 'line'),
];

const pureBlack: GlyphDef = [
  box(3, 3.5, 21, 16.5, 2.5),
  path(
    'M7 6H17A1.5 1.5 0 0 1 18.5 7.5V12.5A1.5 1.5 0 0 1 17 14H7A1.5 1.5 0 0 1 5.5 12.5V7.5A1.5 1.5 0 0 1 7 6Z',
    'detail',
    true,
  ),
  polyline([v(12, 16.5), v(12, 20.4)], undefined, 'line'),
  polyline([v(8, 20.4), v(16, 20.4)], undefined, 'line'),
];

const palette: GlyphDef = [
  path(
    'M20.59 13.9A8.8 8.8 0 1 0 13.9 20.59A1.2 1.2 0 0 0 14.83 19.28A4 4 0 0 1 19.28 14.83A1.2 1.2 0 0 0 20.59 13.9Z',
  ),
  circle(7.54, 13.89, 1.4, { role: 'detail', solid: true }),
  circle(7.96, 9.32, 1.4, { role: 'detail', solid: true }),
  circle(12, 7.15, 1.4, { role: 'detail', solid: true }),
  circle(16.04, 9.32, 1.4, { role: 'detail', solid: true }),
];

const paintbrush: GlyphDef = [
  polyline(
    [v(9.1, 10.51, 1.2), v(13.49, 14.9, 1.2), v(9.95, 18.43, 3.6), v(4.08, 20.2, 1), v(5.57, 14.05, 3.6)],
    true,
  ),
  polyline([v(12.64, 8.96, 0.6), v(15.04, 11.36, 0.6), v(21.4, 5, 1.7), v(19, 2.6, 1.7)], true),
];

const eyedropper: GlyphDef = [
  polyline(
    [
      v(4.79, 19.21, 0.9),
      v(8.46, 18.08, 1.2),
      v(14.4, 12.14, 0.5),
      v(16.24, 13.98, 1),
      v(18.79, 11.43, 1),
      v(17.73, 10.37, 0.5),
      v(21.4, 6.7, 2.9),
      v(17.3, 2.6, 2.9),
      v(13.63, 6.27, 0.5),
      v(12.57, 5.21, 1),
      v(10.02, 7.76, 1),
      v(11.86, 9.6, 0.5),
      v(5.92, 15.54, 1.2),
    ],
    true,
  ),
];

const motion: GlyphDef = [
  circle(15.6, 12, 5.3),
  polyline([v(4.4, 8), v(7.6, 8)], undefined, 'line'),
  polyline([v(2.9, 12), v(6.5, 12)], undefined, 'line'),
  polyline([v(4.4, 16), v(7.6, 16)], undefined, 'line'),
];

const blur: GlyphDef = [
  circle(9.4, 9.4, 1.6, { role: 'line', solid: true }),
  circle(14.6, 9.4, 1.6, { role: 'line', solid: true }),
  circle(9.4, 14.6, 1.6, { role: 'line', solid: true }),
  circle(14.6, 14.6, 1.6, { role: 'line', solid: true }),
  circle(9.4, 4.2, 1.1, { role: 'line', solid: true }),
  circle(14.6, 4.2, 1.1, { role: 'line', solid: true }),
  circle(4.2, 9.4, 1.1, { role: 'line', solid: true }),
  circle(19.8, 9.4, 1.1, { role: 'line', solid: true }),
  circle(4.2, 14.6, 1.1, { role: 'line', solid: true }),
  circle(19.8, 14.6, 1.1, { role: 'line', solid: true }),
  circle(9.4, 19.8, 1.1, { role: 'line', solid: true }),
  circle(14.6, 19.8, 1.1, { role: 'line', solid: true }),
];

const liquidGlass: GlyphDef = [
  path(
    'M12 3.4C14.6 6.2 18.4 10.2 18.4 14.4A6.4 6.4 0 0 1 5.6 14.4' +
      'C5.6 10.2 9.4 6.2 12 3.4Z',
  ),
  arc(12, 14.4, 3.4, 180 * 0.6, 180 * 0.36, 'detail'),
];

const font: GlyphDef = [
  polyline([v(5.2, 19.6), v(12, 4.4, 0.6), v(18.8, 19.6)], undefined, 'line'),
  polyline([v(7.7, 14), v(16.3, 14)], undefined, 'line'),
];

const textSize: GlyphDef = [
  polyline([v(3.2, 4.4), v(14, 4.4)], undefined, 'line'),
  polyline([v(8.6, 4.4), v(8.6, 19.6)], undefined, 'line'),
  polyline([v(14.8, 10.8), v(21, 10.8)], undefined, 'line'),
  polyline([v(17.9, 10.8), v(17.9, 19.6)], undefined, 'line'),
];

const language: GlyphDef = [
  polyline(
    [v(3.2, 3.9, 3), v(20.8, 3.9, 3), v(20.8, 17.5, 3), v(10, 17.5, 0.8), v(6.2, 21, 0.6), v(6.2, 17.5, 0.6), v(3.2, 17.5, 3)],
    true,
  ),
  polyline([v(8.6, 14.1), v(12, 7.3, 0.4), v(15.4, 14.1)], undefined, 'detail'),
  polyline([v(9.85, 11.6), v(14.15, 11.6)], undefined, 'detail'),
];

const keyboard: GlyphDef = [
  box(3, 4.6, 21, 19.4, 2.8),
  circle(6.75, 8.3, 1.1, { role: 'detail', solid: true }),
  circle(10.25, 8.3, 1.1, { role: 'detail', solid: true }),
  circle(13.75, 8.3, 1.1, { role: 'detail', solid: true }),
  circle(17.25, 8.3, 1.1, { role: 'detail', solid: true }),
  circle(6.75, 11.9, 1.1, { role: 'detail', solid: true }),
  circle(10.25, 11.9, 1.1, { role: 'detail', solid: true }),
  circle(13.75, 11.9, 1.1, { role: 'detail', solid: true }),
  circle(17.25, 11.9, 1.1, { role: 'detail', solid: true }),
  polyline([v(8.6, 15.6), v(15.4, 15.6)], undefined, 'detail'),
];

const photos: GlyphDef = [
  box(3, 6.4, 17.6, 21, 2.6),
  polyline([v(6.4, 3), v(21, 3, 2.6), v(21, 17.6)], undefined, 'line'),
  polyline(
    [v(3, 17.2), v(7.8, 12.4, 0.8), v(12.2, 16.8, 0.8), v(14, 15, 0.8), v(17.6, 18.6)],
    undefined,
    'detail',
  ),
  circle(13.4, 10.6, 1.4, { role: 'detail', solid: true }),
];

const camera: GlyphDef = [
  polyline(
    [
      v(3, 7.5, 2.6),
      v(8.2, 7.5, 0.6),
      v(9.6, 5, 0.8),
      v(14.4, 5, 0.8),
      v(15.8, 7.5, 0.6),
      v(21, 7.5, 2.6),
      v(21, 19.5, 2.6),
      v(3, 19.5, 2.6),
    ],
    true,
  ),
  circle(12, 13.4, 3.4, { role: 'detail' }),
];

const qrCode: GlyphDef = [
  box(3, 3, 10.2, 10.2, 2.2),
  box(13.8, 3, 21, 10.2, 2.2),
  box(3, 13.8, 10.2, 21, 2.2),
  path(
    'M6.1 5.5H7.1A0.6 0.6 0 0 1 7.7 6.1V7.1A0.6 0.6 0 0 1 7.1 7.7H6.1A0.6 0.6 0 0 1 5.5 7.1V6.1A0.6 0.6 0 0 1 6.1 5.5Z',
    'detail',
    true,
  ),
  path(
    'M16.9 5.5H17.9A0.6 0.6 0 0 1 18.5 6.1V7.1A0.6 0.6 0 0 1 17.9 7.7H16.9A0.6 0.6 0 0 1 16.3 7.1V6.1A0.6 0.6 0 0 1 16.9 5.5Z',
    'detail',
    true,
  ),
  path(
    'M6.1 16.3H7.1A0.6 0.6 0 0 1 7.7 16.9V17.9A0.6 0.6 0 0 1 7.1 18.5H6.1A0.6 0.6 0 0 1 5.5 17.9V16.9A0.6 0.6 0 0 1 6.1 16.3Z',
    'detail',
    true,
  ),
  path(
    'M14.5 13.8H15.5A0.7 0.7 0 0 1 16.2 14.5V15.5A0.7 0.7 0 0 1 15.5 16.2H14.5A0.7 0.7 0 0 1 13.8 15.5V14.5A0.7 0.7 0 0 1 14.5 13.8Z',
    'line',
    true,
  ),
  path(
    'M19.3 13.8H20.3A0.7 0.7 0 0 1 21 14.5V15.5A0.7 0.7 0 0 1 20.3 16.2H19.3A0.7 0.7 0 0 1 18.6 15.5V14.5A0.7 0.7 0 0 1 19.3 13.8Z',
    'line',
    true,
  ),
  path(
    'M19.3 18.6H20.3A0.7 0.7 0 0 1 21 19.3V20.3A0.7 0.7 0 0 1 20.3 21H19.3A0.7 0.7 0 0 1 18.6 20.3V19.3A0.7 0.7 0 0 1 19.3 18.6Z',
    'line',
    true,
  ),
];

const torch: GlyphDef = [
  polyline(
    [
      v(2.24, 17.66, 1.8),
      v(6.34, 21.76, 1.8),
      v(12.92, 15.18, 1.2),
      v(15.61, 14.76, 0.8),
      v(16.74, 13.63, 1.2),
      v(10.37, 7.26, 1.2),
      v(9.24, 8.39, 0.8),
      v(8.82, 11.08, 1.2),
    ],
    true,
  ),
  circle(8.68, 15.32, 1.1, { role: 'detail', solid: true }),
  polyline([v(13.27, 5.64), v(15.25, 3.66)], undefined, 'line'),
  polyline([v(15.82, 8.18), v(17.8, 6.2)], undefined, 'line'),
  polyline([v(18.36, 10.73), v(20.34, 8.75)], undefined, 'line'),
];

const torchOff: GlyphDef = [
  polyline(
    [
      v(2.24, 17.66, 1.8),
      v(6.34, 21.76, 1.8),
      v(12.92, 15.18, 1.2),
      v(15.61, 14.76, 0.8),
      v(16.74, 13.63, 1.2),
      v(10.37, 7.26, 1.2),
      v(9.24, 8.39, 0.8),
      v(8.82, 11.08, 1.2),
    ],
    true,
  ),
  circle(8.68, 15.32, 1.1, { role: 'detail', solid: true }),
];

const torchAuto: GlyphDef = [
  polyline(
    [
      v(2.24, 17.66, 1.8),
      v(6.34, 21.76, 1.8),
      v(12.92, 15.18, 1.2),
      v(15.61, 14.76, 0.8),
      v(16.74, 13.63, 1.2),
      v(10.37, 7.26, 1.2),
      v(9.24, 8.39, 0.8),
      v(8.82, 11.08, 1.2),
    ],
    true,
  ),
  circle(8.68, 15.32, 1.1, { role: 'detail', solid: true }),
  polyline([v(16.3, 8), v(18.7, 2.9, 0.4), v(21.1, 8)], undefined, 'line'),
  polyline([v(17.1, 6.3), v(20.3, 6.3)], undefined, 'line'),
];

const vpn: GlyphDef = [
  polyline(
    [v(12, 2.9, 1.5), v(19.6, 5.5, 2), v(19.6, 12.5, 5), v(12, 21.1, 2.5), v(4.4, 12.5, 5), v(4.4, 5.5, 2)],
    true,
  ),
  circle(12, 10.2, 1.6, { role: 'detail', solid: true }),
  polyline([v(12, 11), v(12, 14.6)], undefined, 'detail'),
];

const status: GlyphDef = [
  box(3, 4.2, 21, 19.8, 4.5),
  polyline(
    [
      v(6.4, 12.3),
      v(8.8, 12.3, 0.5),
      v(10.4, 8.2, 0.5),
      v(13.4, 16.4, 0.5),
      v(15, 12.3, 0.5),
      v(17.6, 12.3),
    ],
    undefined,
    'detail',
  ),
];

const verified: GlyphDef = [
  polyline(
    [v(12, 2.9, 1.5), v(19.6, 5.5, 2), v(19.6, 12.5, 5), v(12, 21.1, 2.5), v(4.4, 12.5, 5), v(4.4, 5.5, 2)],
    true,
  ),
  polyline([v(8.7, 11.8), v(11, 14.1), v(15.3, 9.6)], undefined, 'detail'),
];

const swap: GlyphDef = [
  polyline([v(3.5, 8), v(20.5, 8)], undefined, 'line'),
  polyline([v(17.5, 5), v(20.5, 8, 0.4), v(17.5, 11)], undefined, 'line'),
  polyline([v(20.5, 16), v(3.5, 16)], undefined, 'line'),
  polyline([v(6.5, 13), v(3.5, 16, 0.4), v(6.5, 19)], undefined, 'line'),
];

const shuffle: GlyphDef = [
  polyline([v(3, 7), v(7, 7, 2), v(15, 17, 2), v(20.5, 17)], undefined, 'line'),
  path('M3 17H5Q7 17 8.25 15.44L8.69 14.88', 'line'),
  path('M13.31 9.12L13.75 8.56Q15 7 17 7H20.5', 'line'),
  polyline([v(17.5, 4), v(20.5, 7, 0.4), v(17.5, 10)], undefined, 'line'),
  polyline([v(17.5, 14), v(20.5, 17, 0.4), v(17.5, 20)], undefined, 'line'),
];

const split: GlyphDef = [
  polyline([v(4.5, 4), v(12, 11.5, 2.5), v(12, 20.5)], undefined, 'line'),
  polyline([v(19.5, 4), v(12, 11.5, 2.5), v(12, 20.5)], undefined, 'line'),
  polyline([v(4.5, 9.5), v(4.5, 4, 0.5), v(10, 4)], undefined, 'line'),
  polyline([v(14, 4), v(19.5, 4, 0.5), v(19.5, 9.5)], undefined, 'line'),
];

const speed: GlyphDef = [
  path(
    'M6.759 19.057 Q5.259 19.057 4.396 17.83 A8.8 8.8 0 1 1 19.604 17.83 ' +
      'Q18.741 19.057 17.241 19.057 Z',
  ),
  circle(12, 13.4, 1.4, { role: 'detail', solid: true }),
  polyline([v(12, 13.4), v(15.47, 9.26)], undefined, 'detail'),
];

const dataUsage: GlyphDef = [
  path(
    'M10 14 H15.8 Q17 14 16.897 15.194 ' +
      'A7 7 0 1 1 8.806 7.103 Q10 7 10 8.2 Z',
  ),
  path(
    'M13.6 9.4 V4.6 Q13.6 3.4 14.794 3.503 ' +
      'A7 7 0 0 1 20.497 9.206 Q20.6 10.4 19.4 10.4 ' +
      'H14.6 Q13.6 10.4 13.6 9.4 Z',
  ),
];

const memory: GlyphDef = [
  box(3.5, 7.2, 20.5, 16.8, 2.5),
  polyline([v(8, 7.2), v(8, 3.8)], undefined, 'line'),
  polyline([v(12, 7.2), v(12, 3.8)], undefined, 'line'),
  polyline([v(16, 7.2), v(16, 3.8)], undefined, 'line'),
  polyline([v(8, 16.8), v(8, 20.2)], undefined, 'line'),
  polyline([v(12, 16.8), v(12, 20.2)], undefined, 'line'),
  polyline([v(16, 16.8), v(16, 20.2)], undefined, 'line'),
];

const cpu: GlyphDef = [
  box(5.5, 5.5, 18.5, 18.5, 2.2),
  box(9.2, 9.2, 14.8, 14.8, 1.2, 'detail'),
  polyline([v(8.4, 5.5), v(8.4, 2.9)], undefined, 'line'),
  polyline([v(12, 5.5), v(12, 2.9)], undefined, 'line'),
  polyline([v(15.6, 5.5), v(15.6, 2.9)], undefined, 'line'),
  polyline([v(8.4, 18.5), v(8.4, 21.1)], undefined, 'line'),
  polyline([v(12, 18.5), v(12, 21.1)], undefined, 'line'),
  polyline([v(15.6, 18.5), v(15.6, 21.1)], undefined, 'line'),
  polyline([v(5.5, 8.4), v(2.9, 8.4)], undefined, 'line'),
  polyline([v(5.5, 12), v(2.9, 12)], undefined, 'line'),
  polyline([v(5.5, 15.6), v(2.9, 15.6)], undefined, 'line'),
  polyline([v(18.5, 8.4), v(21.1, 8.4)], undefined, 'line'),
  polyline([v(18.5, 12), v(21.1, 12)], undefined, 'line'),
  polyline([v(18.5, 15.6), v(21.1, 15.6)], undefined, 'line'),
];

const devices: GlyphDef = [
  polyline(
    [v(11.4, 14.5), v(4, 14.5, 2), v(4, 4, 2), v(18.8, 4, 2), v(18.8, 6.4)],
    undefined,
    'line',
  ),
  polyline([v(2.9, 18.2), v(11.4, 18.2)], undefined, 'line'),
  box(15, 10, 20.8, 20.5, 2),
];

const broom: GlyphDef = [
  polyline([v(20.1, 3.9), v(13.6, 10.4)], undefined, 'line'),
  polyline([v(15.44, 12.24, 1), v(11.76, 8.56, 1), v(9.78, 10.54), v(13.46, 14.22)], true),
  polyline([v(14.02, 14.78, 1), v(9.22, 9.98, 1), v(3.28, 13.09, 1.5), v(10.91, 20.72, 1.5)], true),
];

const wifi: GlyphDef = [
  path('M3.32 9.56 A13.5 13.5 0 0 1 20.68 9.56', 'line'),
  path('M6.54 13.39 A8.5 8.5 0 0 1 17.46 13.39', 'line'),
  circle(12, 18.6, 1.7, { role: 'line', solid: true }),
];

const wifiOff: GlyphDef = [
  path('M3.32 9.56 A13.5 13.5 0 0 1 6.15 7.73', 'line'),
  path('M10.75 6.46 A13.5 13.5 0 0 1 20.68 9.56', 'line'),
  path('M6.54 13.39 A8.5 8.5 0 0 1 10.56 11.52', 'line'),
  path('M15.27 12.05 A8.5 8.5 0 0 1 17.46 13.39', 'line'),
  circle(12, 18.6, 1.7, { role: 'line', solid: true }),
  polyline([v(4.5, 3), v(20.5, 19)], undefined, 'line'),
];

const gavel: GlyphDef = [
  polyline([v(14.93, 7.15, 1.5), v(11.25, 3.47, 1.5), v(3.47, 11.25, 1.5), v(7.15, 14.93, 1.5)], true),
  polyline([v(11.04, 11.04), v(19.8, 19.8)], undefined, 'line'),
  polyline([v(3.2, 20.5), v(11.5, 20.5)], undefined, 'line'),
];

const appsList: GlyphDef = [
  box(3.5, 3.3, 6.9, 6.7, 1),
  box(3.5, 10.3, 6.9, 13.7, 1),
  box(3.5, 17.3, 6.9, 20.7, 1),
  polyline([v(10.5, 5), v(20.5, 5)], undefined, 'line'),
  polyline([v(10.5, 12), v(20.5, 12)], undefined, 'line'),
  polyline([v(10.5, 19), v(20.5, 19)], undefined, 'line'),
];

const iconTile: GlyphDef = [
  box(3, 3, 21, 21, 5),
  polyline(
    [v(5.6, 16.8), v(9.6, 12.8, 0.8), v(12.8, 16, 0.8), v(14.6, 14.2, 0.8), v(18.4, 18)],
    undefined,
    'detail',
  ),
  circle(15.2, 8.8, 1.5, { role: 'detail', solid: true }),
];

// ---------------------------------------------------------------------------
// The glyph record
// ---------------------------------------------------------------------------

/**
 * Every AppGlyphs glyph by name (identical to the Dart identifiers).
 * `clock` shares the `requests` definition exactly like `AppGlyphs.clock`.
 */
export const glyphs = {
  dashboard,
  proxies,
  profiles,
  requests,
  clock: requests,
  connections,
  dns,
  resources,
  logs,
  tools,
  target,
  arrowBack,
  chevronBack,
  chevronForward,
  chevronDown,
  chevronUp,
  picker,
  caretDown,
  arrowUp,
  arrowDown,
  openExternal,
  undo,
  redo,
  sync,
  refresh,
  reset,
  export: exportGlyph,
  upload,
  add,
  remove,
  addCircle,
  removeCircle,
  close,
  clearField,
  check,
  checkDouble,
  checkCircle,
  block,
  info,
  warning,
  error,
  more,
  moreCircle,
  dragHandle,
  sort,
  filter,
  textShort,
  wrap,
  listAdd,
  edit,
  compose,
  delete: deleteGlyph,
  clearAll,
  copy,
  cut,
  paste,
  selectAll,
  deselect,
  search,
  matchCase,
  regex,
  importFile,
  cloudDownload,
  cloudSync,
  link: linkGlyph,
  send,
  rules,
  rulePresets,
  code,
  settings,
  sliders,
  wrench,
  toggle,
  toggleOn,
  pin,
  eye,
  eyeOff,
  lock,
  lockOpen,
  key,
  password,
  account,
  sparkle,
  bolt,
  locate,
  star,
  puzzle,
  customize,
  layers,
  sun,
  pureBlack,
  palette,
  paintbrush,
  eyedropper,
  motion,
  blur,
  liquidGlass,
  font,
  textSize,
  language,
  keyboard,
  photos,
  camera,
  qrCode,
  torch,
  torchOff,
  torchAuto,
  vpn,
  status,
  verified,
  swap,
  shuffle,
  split,
  speed,
  dataUsage,
  memory,
  cpu,
  devices,
  broom,
  wifi,
  wifiOff,
  gavel,
  appsList,
  iconTile,
} satisfies Record<string, GlyphDef>;

/** Union of all glyph names in the `glyphs` record. */
export type GlyphName = keyof typeof glyphs;

// ---------------------------------------------------------------------------
// Parametric glyphs
// ---------------------------------------------------------------------------

/** Sidebar drawer glyph; `progress` 0..1 moves the divider across the panel. */
export function sidebar(progress: number): GlyphDef {
  const divider = 8.5 + (11.5 - 8.5) * progress;
  return [
    box(3, 4.5, 21, 19.5, 3.5),
    polyline([v(divider, 4.5), v(divider, 19.5)], undefined, 'detail'),
    ...[8.5, 11.0, 13.5].map((y) =>
      polyline([v(5.5, y), v(divider - 2.5, y)], undefined, 'detail'),
    ),
  ];
}

type Point2 = [number, number];

const playCorners: Point2[] = [
  [8.3, 5.6],
  [18.8, 12],
  [8.3, 18.4],
];

const playHalves: { half: Point2[]; bar: Point2[] }[] = [
  {
    half: [
      [8.3, 5.6],
      [18.8, 12],
      [18.8, 12],
      [8.3, 12],
    ],
    bar: [
      [5.9, 6.9],
      [18.1, 6.9],
      [18.1, 9.1],
      [5.9, 9.1],
    ],
  },
  {
    half: [
      [8.3, 12],
      [18.8, 12],
      [18.8, 12],
      [8.3, 18.4],
    ],
    bar: [
      [5.9, 14.9],
      [18.1, 14.9],
      [18.1, 17.1],
      [5.9, 17.1],
    ],
  },
];

/** Rotate a point around the grid center (12, 12) — Dart `_turnedVertex`. */
function turnedVertex(corner: Point2, angle: number): PolyVertex {
  const dx = corner[0] - 12;
  const dy = corner[1] - 12;
  const cos = Math.cos(angle);
  const sin = Math.sin(angle);
  return v(12 + dx * cos - dy * sin, 12 + dx * sin + dy * cos, 1.2);
}

function lerpPoint(a: Point2, b: Point2, t: number): Point2 {
  return [a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t];
}

/**
 * Play at 0, pause at 1; past either end it keeps turning, so it bounces.
 * Port of Dart `AppGlyphs.playPause(progress)`.
 */
export function playPause(progress: number): GlyphDef {
  const fold = Math.min(1, Math.max(0, progress));
  const turn = (Math.PI / 2) * progress;
  if (fold === 0) {
    return [polyline(playCorners.map((corner) => turnedVertex(corner, turn)), true)];
  }
  return playHalves.map(({ half, bar }) =>
    polyline(
      half.map((point, i) => turnedVertex(lerpPoint(point, bar[i], fold), turn)),
      true,
    ),
  );
}

/**
 * Port of Dart `Glyphs.backFor(platform)`: the web replica is always
 * "web"/LTR, which uses `arrowBack`.
 */
export function backFor(): GlyphDef {
  return glyphs.arrowBack;
}

/**
 * Glyphs declared with `matchTextDirection: true` in Dart; flip them
 * horizontally in RTL locales (`sidebar` is the parametric one).
 */
export const matchTextDirection: ReadonlySet<string> = new Set([
  'arrowBack',
  'chevronBack',
  'chevronForward',
  'undo',
  'redo',
  'textShort',
  'wrap',
  'listAdd',
  'rules',
  'rulePresets',
  'sort',
  'sidebar',
]);
