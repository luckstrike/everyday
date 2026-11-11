import assert from 'node:assert/strict';
import { test } from 'node:test';
import { fittingRowCount } from '../src/lib/fit.ts';

test('fits any number of whole rows without a three-row cap', () => {
  const rows = Array.from({ length: 10 }, () => 60);

  assert.equal(fittingRowCount(60, rows, 8), 1);
  assert.equal(fittingRowCount(128, rows, 8), 2);
  assert.equal(fittingRowCount(196, rows, 8), 3);
  assert.equal(fittingRowCount(264, rows, 8), 4);
  assert.equal(fittingRowCount(400, rows, 8), 6);
  assert.equal(fittingRowCount(1000, rows, 8), 10);
});

test('never shows a partial row or charges a trailing gap', () => {
  assert.equal(fittingRowCount(0, [60, 60], 8), 0);
  assert.equal(fittingRowCount(59, [60, 60], 8), 0);
  assert.equal(fittingRowCount(127, [60, 60], 8), 1);
  assert.equal(fittingRowCount(128, [60, 60], 8), 2);
  assert.equal(fittingRowCount(400, [], 8), 0);
});

test('uses measured row heights and gaps rather than size presets', () => {
  assert.equal(fittingRowCount(162, [76, 76, 76], 10), 2);
  assert.equal(fittingRowCount(248, [76, 76, 76], 10), 3);
  assert.equal(fittingRowCount(169, [45, 70, 36], 9), 3);
  assert.equal(fittingRowCount(168, [45, 70, 36], 9), 2);
  assert.equal(fittingRowCount(121.5, [56.25, 56.25], 9), 2);
  assert.equal(fittingRowCount(121.49, [56.25, 56.25], 9), 1);
});

test('recomputes capacity when space, rows, or their sizes change', () => {
  assert.equal(fittingRowCount(264, [60, 60, 60, 60], 8), 4);
  assert.equal(fittingRowCount(128, [60, 60, 60, 60], 8), 2);
  assert.equal(fittingRowCount(264, [76, 76, 76, 76], 10), 3);
  assert.equal(fittingRowCount(264, [60, 60], 8), 2);
  assert.equal(fittingRowCount(264, [0, 60], 8), 0);
});
