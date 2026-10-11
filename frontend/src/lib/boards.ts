import type { Placement, Size, WidgetDefinition, WidgetId } from './types.ts';

export type Orientation = 'landscape' | 'portrait';
export interface Board {
  id: string;
  name: string;
  layouts: Record<Orientation, Placement[]>;
}
export interface BoardStore {
  version: 2;
  activeId: string;
  boards: Board[];
}
export const BOARDS_KEY = 'apartment-display.boards.v2';
export const GRID = 24;
export const widgetIds: WidgetId[] = [
  'clock',
  'weather',
  'calendar',
  'finance',
  'discover',
  'sports',
];

export function everyday(): Board {
  return {
    id: 'everyday',
    name: 'Everyday',
    layouts: {
      landscape: [
        { id: 'clock', x: 0, y: 0, w: 14, h: 12 },
        { id: 'weather', x: 14, y: 0, w: 10, h: 12 },
        { id: 'calendar', x: 0, y: 12, w: 7, h: 8 },
        { id: 'finance', x: 7, y: 12, w: 7, h: 8 },
        { id: 'discover', x: 14, y: 12, w: 10, h: 8 },
        { id: 'sports', x: 0, y: 20, w: 24, h: 4 },
      ],
      portrait: [
        { id: 'clock', x: 0, y: 0, w: 24, h: 5 },
        { id: 'weather', x: 0, y: 5, w: 24, h: 6 },
        { id: 'calendar', x: 0, y: 11, w: 12, h: 5 },
        { id: 'finance', x: 12, y: 11, w: 12, h: 5 },
        { id: 'discover', x: 0, y: 16, w: 24, h: 5 },
        { id: 'sports', x: 0, y: 21, w: 24, h: 3 },
      ],
    },
  };
}
export function initialBoards(): BoardStore {
  return { version: 2, activeId: 'everyday', boards: [everyday()] };
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
  const def = definitions.find((d) => d.id === item.id);
  return (
    !!def &&
    [item.x, item.y, item.w, item.h].every(Number.isInteger) &&
    item.x >= 0 &&
    item.y >= 0 &&
    item.w >= def.minWidth * 2 &&
    item.h >= def.minHeight &&
    item.x + item.w <= GRID &&
    item.y + item.h <= GRID &&
    !items.some((other) => other.id !== item.id && overlaps(item, other))
  );
}
export function footprint(
  id: WidgetId,
  size: Size,
  orientation: Orientation,
): Pick<Placement, 'w' | 'h'> {
  if (orientation === 'portrait')
    return size === 'small'
      ? { w: 12, h: 4 }
      : size === 'medium'
        ? { w: 24, h: 5 }
        : { w: 24, h: 8 };
  if (size === 'small') return { w: 6, h: id === 'weather' ? 6 : 4 };
  return size === 'medium' ? { w: 10, h: 8 } : { w: 14, h: 12 };
}
export function validPositions(
  id: WidgetId,
  size: Size,
  orientation: Orientation,
  items: Placement[],
  definitions: WidgetDefinition[],
): Placement[] {
  const shape = footprint(id, size, orientation);
  const result: Placement[] = [];
  for (let y = 0; y <= GRID - shape.h; y++)
    for (let x = 0; x <= GRID - shape.w; x++) {
      const item = { id, x, y, ...shape };
      if (canPlace(item, items, definitions)) result.push(item);
    }
  return result;
}
function validLayout(
  items: unknown,
  definitions: WidgetDefinition[],
): items is Placement[] {
  return (
    Array.isArray(items) &&
    items.length <= widgetIds.length &&
    items.every((item) => item && widgetIds.includes(item.id)) &&
    new Set(items.map((item) => item.id)).size === items.length &&
    items.every((item) => canPlace(item, items, definitions))
  );
}
export function parseBoards(
  raw: string,
  definitions: WidgetDefinition[],
): BoardStore {
  const value = JSON.parse(raw);
  if (
    !value ||
    value.version !== 2 ||
    !Array.isArray(value.boards) ||
    !value.boards.length ||
    value.boards.length > 50 ||
    value.boards.some(
      (b: Board) =>
        !b ||
        typeof b.id !== 'string' ||
        !b.id ||
        typeof b.name !== 'string' ||
        !b.name.trim() ||
        b.name.length > 40 ||
        !validLayout(b.layouts?.landscape, definitions) ||
        !validLayout(b.layouts?.portrait, definitions),
    ) ||
    new Set(value.boards.map((b: Board) => b.id)).size !==
      value.boards.length ||
    !value.boards.some((b: Board) => b.id === value.activeId)
  )
    throw new Error('Invalid saved boards');
  return value;
}
// Legacy geometry is retained on landscape boards. The old key is never overwritten.
export function migrateLayouts(
  raw: string,
  definitions: WidgetDefinition[],
): BoardStore {
  const legacy = JSON.parse(raw);
  const boards: Board[] = [];
  for (const [id, name] of [
    ['home', 'Everyday'],
    ['weather', 'Weather'],
    ['sports', 'Sports'],
    ['finance', 'Bills'],
  ]) {
    const items = legacy?.[id];
    if (!Array.isArray(items)) throw new Error('Invalid legacy layout');
    const height = Math.max(6, ...items.map((p: Placement) => p.y + p.h));
    const factor = height <= 6 ? 4 : 2;
    const landscape = items.map((p: Placement) => ({
      ...p,
      x: p.x * 2,
      y: p.y * factor,
      w: p.w * 2,
      h: p.h * factor,
    }));
    if (!validLayout(landscape, definitions))
      throw new Error('Invalid legacy layout');
    // Untouched defaults adopt the prototype. Customized geometry becomes an extra board.
    const isDefault =
      id === 'home' &&
      items.length === 6 &&
      items.every(
        (p: Placement) =>
          p.w === 4 && p.h === 3 && p.x % 4 === 0 && p.y % 3 === 0,
      );
    boards.push(
      isDefault
        ? everyday()
        : {
            id,
            name,
            layouts: {
              landscape,
              portrait: everyday().layouts.portrait.filter((p) =>
                items.some((old: Placement) => old.id === p.id),
              ),
            },
          },
    );
  }
  return { version: 2, activeId: boards[0].id, boards };
}
