import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  initialBoards,
  parseBoards,
  canPlace,
  validPositions,
  footprint,
  migrateLayouts,
  widgetIds,
} from '../src/lib/boards.ts';
import type { WidgetDefinition } from '../src/lib/types.ts';
const definitions = widgetIds.map((id) => ({
  id,
  title: id,
  minWidth: 6,
  minHeight: 2,
  basePriority: 10,
  defaultWidth: 4,
  defaultHeight: 3,
})) satisfies WidgetDefinition[];
test('prototype boards fit in both orientations and round trip without losing geometry', () => {
  const store = initialBoards();
  assert.deepEqual(parseBoards(JSON.stringify(store), definitions), store);
  for (const items of Object.values(store.boards[0].layouts))
    for (const item of items) assert.ok(canPlace(item, items, definitions));
});
test('placement uses actual footprint, protects occupied space and leaves spare space untouched', () => {
  const items = initialBoards().boards[0].layouts.landscape.filter(
    (p) => p.id !== 'calendar',
  );
  const original = JSON.stringify(items);
  const small = validPositions(
    'calendar',
    'small',
    'landscape',
    items,
    definitions,
  );
  assert.ok(small.length > 0);
  assert.ok(
    small.every(
      (p) =>
        p.w === footprint('calendar', 'small', 'landscape').w &&
        canPlace(p, items, definitions),
    ),
  );
  assert.equal(
    validPositions('calendar', 'large', 'landscape', items, definitions).length,
    0,
  );
  assert.equal(JSON.stringify(items), original);
  assert.equal(
    validPositions(
      'calendar',
      'small',
      'landscape',
      [{ id: 'clock', x: 0, y: 0, w: 24, h: 24 }],
      definitions,
    ).length,
    0,
  );
});
test('saved boards reject collision, duplicate widgets, invalid active board, unknown ID and out of bounds', () => {
  for (const patch of [
    { x: -1 },
    { w: 25 },
    { id: 'unknown' },
    { x: 14 },
    { x: 0.5 },
  ]) {
    const store = initialBoards();
    Object.assign(store.boards[0].layouts.landscape[0], patch);
    assert.throws(() => parseBoards(JSON.stringify(store), definitions));
  }
  const store = initialBoards();
  store.boards[0].layouts.landscape.push(store.boards[0].layouts.landscape[0]);
  assert.throws(() => parseBoards(JSON.stringify(store), definitions));
  assert.throws(() =>
    parseBoards(
      JSON.stringify({ ...initialBoards(), activeId: 'missing' }),
      definitions,
    ),
  );
});
test('legacy custom layouts survive migration and leave source intact', () => {
  const items = [{ id: 'clock', x: 1, y: 0, w: 3, h: 3 }];
  const raw = JSON.stringify({
    home: items,
    weather: [],
    sports: [],
    finance: [],
  });
  const migrated = migrateLayouts(raw, definitions);
  assert.deepEqual(migrated.boards[0].layouts.landscape, [
    { id: 'clock', x: 2, y: 0, w: 6, h: 12 },
  ]);
  assert.deepEqual(
    parseBoards(JSON.stringify(migrated), definitions),
    migrated,
  );
  assert.deepEqual(JSON.parse(raw).home, items);
});
test('migration never resets a rearranged six-widget default', () => {
  const home = widgetIds.map((id, i) => ({
    id,
    x: (i % 3) * 4,
    y: Math.floor(i / 3) * 3,
    w: 4,
    h: 3,
  }));
  const migrated = migrateLayouts(
    JSON.stringify({ home, weather: [], sports: [], finance: [] }),
    definitions,
  );
  assert.deepEqual(
    migrated.boards[0].layouts.landscape,
    home.map((p) => ({ ...p, x: p.x * 2, y: p.y * 4, w: p.w * 2, h: p.h * 4 })),
  );
});
