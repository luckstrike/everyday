import type { Candidate } from './priority.ts';
import type {
  Layouts,
  Placement,
  ScreenId,
  Size,
  WidgetDefinition,
  WidgetId,
} from './types.ts';
export const screenIds: ScreenId[] = ['home', 'weather', 'sports', 'finance'];
export const widgetIds: WidgetId[] = [
  'clock',
  'weather',
  'sports',
  'finance',
  'calendar',
];
export const STORAGE_KEY = 'apartment-display.layouts.v1';
export function defaultLayouts(): Layouts {
  return {
    home: [
      { id: 'clock', x: 0, y: 0, w: 7, h: 3 },
      { id: 'weather', x: 7, y: 0, w: 5, h: 3 },
      { id: 'calendar', x: 0, y: 3, w: 4, h: 3 },
      { id: 'finance', x: 4, y: 3, w: 4, h: 3 },
      { id: 'sports', x: 8, y: 3, w: 4, h: 3 },
    ],
    weather: [
      { id: 'weather', x: 0, y: 0, w: 9, h: 6 },
      { id: 'clock', x: 9, y: 0, w: 3, h: 3 },
      { id: 'calendar', x: 9, y: 3, w: 3, h: 3 },
    ],
    sports: [
      { id: 'sports', x: 0, y: 0, w: 9, h: 6 },
      { id: 'clock', x: 9, y: 0, w: 3, h: 3 },
      { id: 'calendar', x: 9, y: 3, w: 3, h: 3 },
    ],
    finance: [
      { id: 'finance', x: 0, y: 0, w: 9, h: 6 },
      { id: 'clock', x: 9, y: 0, w: 3, h: 3 },
      { id: 'calendar', x: 9, y: 3, w: 3, h: 3 },
    ],
  };
}
export function overlaps(a: Placement, b: Placement): boolean {
  return (
    a.x < b.x + b.w && a.x + a.w > b.x && a.y < b.y + b.h && a.y + a.h > b.y
  );
}
export function canPlace(
  item: Placement,
  items: Placement[],
  definitions: WidgetDefinition[],
): boolean {
  const def = definitions.find((w) => w.id === item.id);
  return (
    !!def &&
    [item.x, item.y, item.w, item.h].every(Number.isInteger) &&
    item.x >= 0 &&
    item.y >= 0 &&
    item.w >= def.minWidth &&
    item.h >= def.minHeight &&
    item.x + item.w <= 12 &&
    item.y + item.h <= 12 &&
    !items.some((other) => other.id !== item.id && overlaps(item, other))
  );
}
export function parseLayouts(
  raw: string,
  definitions: WidgetDefinition[],
): Layouts {
  const parsed = JSON.parse(raw);
  if (!parsed || typeof parsed !== 'object')
    throw new Error('Invalid saved layout');
  for (const id of screenIds) {
    const list = parsed[id];
    if (
      !Array.isArray(list) ||
      list.length > 5 ||
      list.some((item) => !item || !widgetIds.includes(item.id)) ||
      new Set(list.map((item) => item.id)).size !== list.length ||
      list.some((item) => !canPlace(item, list, definitions))
    )
      throw new Error('Invalid saved layout');
  }
  return parsed;
}
export function addWidget(
  items: Placement[],
  definition: WidgetDefinition,
): Placement[] | null {
  if (items.some((item) => item.id === definition.id)) return items;
  for (let y = 0; y <= 12 - definition.defaultHeight; y++)
    for (let x = 0; x <= 12 - definition.defaultWidth; x++) {
      const next = {
        id: definition.id,
        x,
        y,
        w: definition.defaultWidth,
        h: definition.defaultHeight,
      };
      if (canPlace(next, items, [definition])) return [...items, next];
    }
  return null;
}
export function widgetSize(item: Placement): Size {
  return item.w >= 7 && item.h >= 4
    ? 'large'
    : item.w >= 4 && item.h >= 3
      ? 'medium'
      : 'small';
}
// Pure projection: no changes are ever written back to the configured layout.
export function presentationLayout(
  items: Placement[],
  attention?: Candidate,
): Placement[] {
  if (!attention?.widget || attention.presentation === 'NORMAL')
    return items.map((item) => ({ ...item }));
  if (attention.presentation === 'FULLSCREEN')
    return [{ id: attention.widget, x: 0, y: 0, w: 12, h: 6 }];
  const others = items
    .filter((item) => item.id !== attention.widget)
    .slice(0, 4);
  return [
    { id: attention.widget, x: 0, y: 0, w: 8, h: 6 },
    ...others.map((item, i) => ({ ...item, x: 8, y: i * 2, w: 4, h: 2 })),
  ];
}

export const PINS_KEY = 'apartment-display.pins.v1';
export function parsePins(raw: string): WidgetId[] {
  const pins: unknown = JSON.parse(raw);
  if (
    !Array.isArray(pins) ||
    pins.some((id) => !widgetIds.includes(id)) ||
    new Set(pins).size !== pins.length
  )
    throw new Error('Invalid widget pins');
  return pins;
}

// Pins are a display preference; never mutate saved screen geometry.
export function pinnedLayout(
  items: Placement[],
  pins: WidgetId[],
  fullscreen = false,
): Placement[] {
  if (fullscreen) return items.map((item) => ({ ...item }));
  const missing = pins.filter((id) => !items.some((item) => item.id === id));
  if (!missing.length) return items.map((item) => ({ ...item }));
  const all = [
    ...items,
    ...missing.map((id) => ({ id, x: 0, y: 0, w: 4, h: 2 })),
  ];
  const [primary, ...others] = all;
  return [
    { ...primary, x: 0, y: 0, w: 8, h: 6 },
    ...others.map((item, i) => ({ ...item, x: 8, y: i * 2, w: 4, h: 2 })),
  ];
}

// Reflow by saved reading order into a portrait composition. The source layout survives rotation.
export function portraitLayout(
  items: Placement[],
  narrow = false,
): Placement[] {
  const ordered = [...items].sort((a, b) => a.y - b.y || a.x - b.x);
  if (ordered.length <= 1)
    return ordered.map((item) => ({ ...item, x: 0, y: 0, w: 12, h: 6 }));
  if (narrow)
    return ordered.map((item, i) => ({ ...item, x: 0, y: i * 2, w: 12, h: 2 }));
  const [primary, ...rest] = ordered;
  const height = primary.id === 'clock' ? 2 : 4;
  return [
    { ...primary, x: 0, y: 0, w: 12, h: height },
    ...rest.map((item, i) => ({
      ...item,
      x: (i % 2) * 6,
      y: height + Math.floor(i / 2) * 3,
      w: 6,
      h: 3,
    })),
  ];
}
