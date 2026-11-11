import assert from 'node:assert/strict';
import { test } from 'node:test';
import { fittingItemCount } from '../src/lib/fit.ts';

test('fits any number of whole rows without a three-row cap', () => {
  const rows = Array.from({ length: 10 }, () => 60);

  assert.equal(fittingItemCount(60, rows, 8), 1);
  assert.equal(fittingItemCount(128, rows, 8), 2);
  assert.equal(fittingItemCount(196, rows, 8), 3);
  assert.equal(fittingItemCount(264, rows, 8), 4);
  assert.equal(fittingItemCount(400, rows, 8), 6);
  assert.equal(fittingItemCount(1000, rows, 8), 10);
});

test('never shows a partial row or charges a trailing gap', () => {
  assert.equal(fittingItemCount(0, [60, 60], 8), 0);
  assert.equal(fittingItemCount(59, [60, 60], 8), 0);
  assert.equal(fittingItemCount(127, [60, 60], 8), 1);
  assert.equal(fittingItemCount(128, [60, 60], 8), 2);
  assert.equal(fittingItemCount(400, [], 8), 0);
});

test('uses measured row heights and gaps rather than size presets', () => {
  assert.equal(fittingItemCount(162, [76, 76, 76], 10), 2);
  assert.equal(fittingItemCount(248, [76, 76, 76], 10), 3);
  assert.equal(fittingItemCount(169, [45, 70, 36], 9), 3);
  assert.equal(fittingItemCount(168, [45, 70, 36], 9), 2);
  assert.equal(fittingItemCount(121.5, [56.25, 56.25], 9), 2);
  assert.equal(fittingItemCount(121.49, [56.25, 56.25], 9), 1);
});

test('recomputes capacity when space, rows, or their sizes change', () => {
  assert.equal(fittingItemCount(264, [60, 60, 60, 60], 8), 4);
  assert.equal(fittingItemCount(128, [60, 60, 60, 60], 8), 2);
  assert.equal(fittingItemCount(264, [76, 76, 76, 76], 10), 3);
  assert.equal(fittingItemCount(264, [60, 60], 8), 2);
  assert.equal(fittingItemCount(264, [0, 60], 8), 0);
});

test('fits forecast cards using content, padding, borders and horizontal gaps', () => {
  const contentWidth = 82;
  const borderWidth = 2;
  const compactCard = contentWidth + 4 * 2 + borderWidth;
  const roomyCard = contentWidth + 8 * 2 + borderWidth;

  assert.equal(fittingItemCount(296, Array(5).fill(roomyCard), 8), 2);
  assert.equal(fittingItemCount(296, Array(5).fill(compactCard), 8), 3);
  assert.equal(fittingItemCount(401, Array(5).fill(compactCard), 8), 4);
  assert.equal(fittingItemCount(716, Array(5).fill(roomyCard), 8), 5);
  assert.equal(fittingItemCount(296, [92, 108, 80, 92], 8), 3);
  assert.equal(fittingItemCount(295, [92, 108, 80, 92], 8), 2);
});

test('a section fits only when every block above it fits as well', () => {
  // Current conditions 72px tall, forecast cards 81px, 28px apart: 181px in all.
  const blocks = [72, 81];
  const fits = (room: number) =>
    fittingItemCount(room, blocks, 28) === blocks.length;

  assert.equal(fits(181), true);
  assert.equal(fits(300), true);
  assert.equal(fits(180.9), false);
  // Without room even for the current conditions, the forecast cannot fit either.
  assert.equal(fits(60), false);
});
