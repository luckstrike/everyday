import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  amountText,
  daysLateText,
  orderBills,
  overdueCount,
} from '../src/lib/bills.ts';

const bill = (name: string, dueAt: string, paid = false, overdue = false) => ({
  name,
  dueAt,
  paid,
  overdue,
});

test('bills left to pay come first, soonest due, then the paid ones in the same order', () => {
  const ordered = orderBills([
    bill('Internet', '2026-11-12T00:00:00', true),
    bill('Rent', '2026-11-20T00:00:00'),
    bill('Phone', '2026-11-14T00:00:00'),
    bill('Water', '2026-11-13T00:00:00', true),
  ]);
  assert.deepEqual(
    ordered.map((b) => b.name),
    ['Phone', 'Rent', 'Internet', 'Water'],
  );
});
test('missed bills lead, the longest overdue first, ahead of anything still to pay', () => {
  const ordered = orderBills([
    bill('Rent', '2026-11-20T00:00:00'),
    bill('Phone', '2026-11-12T00:00:00', false, true),
    bill('Internet', '2026-11-01T00:00:00', true),
    bill('Water', '2026-11-05T00:00:00', false, true),
  ]);
  assert.deepEqual(
    ordered.map((b) => b.name),
    ['Water', 'Phone', 'Rent', 'Internet'],
  );
});
test('ordering leaves the list it was given alone', () => {
  const given = [
    bill('B', '2026-11-20T00:00:00'),
    bill('A', '2026-11-10T00:00:00'),
  ];
  orderBills(given);
  assert.deepEqual(
    given.map((b) => b.name),
    ['B', 'A'],
  );
});
test('overdue bills are counted for the header chip', () => {
  assert.equal(overdueCount([]), 0);
  assert.equal(
    overdueCount([
      bill('A', '2026-11-10T00:00:00', false, true),
      bill('B', '2026-11-20T00:00:00'),
      bill('C', '2026-11-11T00:00:00', false, true),
    ]),
    2,
  );
});
test('days late are worded in the singular for one day', () => {
  assert.equal(daysLateText(1), '1 day overdue');
  assert.equal(daysLateText(4), '4 days overdue');
});
test('a bill with no fixed amount says Due until it is paid, then nothing', () => {
  assert.equal(amountText(84.5, false), '$84.50');
  assert.equal(amountText(84.5, true), '$84.50');
  assert.equal(amountText(0, false), 'Due');
  assert.equal(amountText(0, true), '');
});
