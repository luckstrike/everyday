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

test('money always shows cents, so whole amounts read like the rest', async () => {
  const { money } = await import('../src/lib/format.ts');
  assert.equal(money(65), '$65.00');
  assert.equal(money(84.5), '$84.50');
  assert.equal(money(1850), '$1,850.00');
  assert.equal(money(0), '$0.00');
  assert.equal(money(1234567.891), '$1,234,567.89');
});

test('a date is written for people as read, never shifted to another day', async () => {
  const { shortDate } = await import('../src/lib/format.ts');
  assert.equal(shortDate('2026-11-18'), 'Nov 18');
  assert.equal(shortDate('2026-12-01'), 'Dec 1');
  assert.equal(shortDate('2027-01-31'), 'Jan 31');
});
