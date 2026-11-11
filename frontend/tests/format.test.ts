import { test } from 'node:test';
import assert from 'node:assert/strict';
import { when } from '../src/lib/format.ts';

// Dates are built in local time, like the display's own clock, so the tests hold in any zone.
const at = (month: number, date: number, hour = 0) =>
  new Date(2026, month - 1, date, hour);
const now = at(11, 11, 10).getTime();

test('a timed event reads as a day and a clock time', () => {
  const text = when({ startsAt: at(11, 12, 18).toISOString() }, now);
  assert.ok(text.startsWith('Tomorrow · '));
  assert.ok(!text.includes('All day'));
});
test('an all-day event says all day instead of midnight', () => {
  assert.equal(
    when({ startsAt: at(11, 12).toISOString(), allDay: true }, now),
    'Tomorrow · All day',
  );
});
test('an all-day event already under way reads as today', () => {
  assert.equal(
    when({ startsAt: at(11, 9).toISOString(), allDay: true }, now),
    'Today · All day',
  );
});
