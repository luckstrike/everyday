import { test } from 'node:test';
import assert from 'node:assert/strict';
import { monthGrid } from '../src/lib/month.ts';

test('month grid places dates under the correct weekdays and pads complete weeks', () => {
  const month = monthGrid(2026, 9);
  assert.equal(month.label, 'October 2026');
  assert.deepEqual(month.weeks[0], [null, null, null, null, 1, 2, 3]);
  assert.deepEqual(month.weeks.at(-1), [25, 26, 27, 28, 29, 30, 31]);
  assert.equal(month.weeks.flat().filter((day) => day !== null).length, 31);
  assert.ok(month.weeks.every((week) => week.length === 7));
});
test('month grid handles leap years, four-week months, six-week months and year boundaries', () => {
  assert.equal(monthGrid(2024, 1).weeks.flat().filter(Boolean).length, 29);
  assert.equal(monthGrid(2025, 1).weeks.flat().filter(Boolean).length, 28);
  assert.equal(monthGrid(2026, 1).weeks.length, 4);
  assert.equal(monthGrid(2026, 7).weeks.length, 6);
  assert.equal(monthGrid(2026, 11).weeks.flat().filter(Boolean).length, 31);
  assert.deepEqual(monthGrid(2027, 0).weeks[0], [
    null,
    null,
    null,
    null,
    null,
    1,
    2,
  ]);
});
