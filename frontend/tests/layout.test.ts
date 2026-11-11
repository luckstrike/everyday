import { test } from 'node:test';
import assert from 'node:assert/strict';
import { GRID, layout, overlaps, widgetTitles } from '../src/lib/layout.ts';

test('every widget sits inside the grid on whole cells', () => {
  for (const item of layout) {
    for (const value of [item.x, item.y, item.w, item.h])
      assert.ok(Number.isInteger(value), item.id);
    assert.ok(item.x >= 0 && item.y >= 0, item.id);
    assert.ok(item.w > 0 && item.h > 0, item.id);
    assert.ok(item.x + item.w <= GRID && item.y + item.h <= GRID, item.id);
  }
});

test('widgets never overlap and appear only once', () => {
  const ids = layout.map((item) => item.id);
  assert.equal(new Set(ids).size, ids.length);
  for (const [i, a] of layout.entries())
    for (const b of layout.slice(i + 1))
      assert.ok(!overlaps(a, b), `${a.id} overlaps ${b.id}`);
});

test('every placed widget has a title', () => {
  for (const item of layout) assert.ok(widgetTitles[item.id], item.id);
});
