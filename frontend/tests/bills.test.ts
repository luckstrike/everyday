import { test } from 'node:test';
import assert from 'node:assert/strict';
import { orderBills } from '../src/lib/bills.ts';

const bill = (name: string, dueAt: string, paid: boolean) => ({
  name,
  dueAt,
  paid,
});

test('bills left to pay come first, soonest due, then the paid ones in the same order', () => {
  const ordered = orderBills([
    bill('Internet', '2026-11-12T00:00:00', true),
    bill('Rent', '2026-11-20T00:00:00', false),
    bill('Phone', '2026-11-14T00:00:00', false),
    bill('Water', '2026-11-13T00:00:00', true),
  ]);
  assert.deepEqual(
    ordered.map((b) => b.name),
    ['Phone', 'Rent', 'Internet', 'Water'],
  );
});
test('ordering leaves the list it was given alone', () => {
  const given = [
    bill('B', '2026-11-20T00:00:00', false),
    bill('A', '2026-11-10T00:00:00', false),
  ];
  orderBills(given);
  assert.deepEqual(
    given.map((b) => b.name),
    ['B', 'A'],
  );
});
