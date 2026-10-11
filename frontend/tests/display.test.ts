import { test } from 'node:test';
import assert from 'node:assert/strict';
import { billEvents, daysUntil } from '../src/lib/bills.ts';
import { eventModifier, rankWidgets } from '../src/lib/priority.ts';
import { initialBoards } from '../src/lib/boards.ts';
import {
  initialPresentation,
  attentionFor,
  displayLayout,
} from '../src/lib/presentation.ts';
import type { WidgetDefinition } from '../src/lib/types.ts';
const now = new Date(2026, 9, 4, 23, 50).getTime();
const bill = (days: number) => ({
  name: 'Electricity',
  amount: 84.5,
  autopay: true,
  dueAt: new Date(2026, 9, 4 + days, 0, 10).toISOString(),
});
const finance: WidgetDefinition = {
  id: 'finance',
  title: 'Bills',
  basePriority: 25,
  minWidth: 6,
  minHeight: 2,
  defaultWidth: 4,
  defaultHeight: 3,
};
test('bill attention uses calendar days across midnight and overdue boundaries', () => {
  assert.equal(daysUntil(bill(1).dueAt, now), 1);
  for (const [days, score] of [
    [8, 0],
    [7, 5],
    [3, 15],
    [1, 25],
    [0, 35],
    [-1, 40],
  ])
    assert.equal(
      billEvents([bill(days)], now).reduce(
        (sum, e) => sum + eventModifier(e, now),
        0,
      ),
      score,
    );
  assert.deepEqual(billEvents([{ ...bill(1), dueAt: 'invalid' }], now), []);
});
test('bill date events are stable and never rearrange the board', () => {
  const saved = initialBoards().boards[0].layouts.landscape;
  assert.equal(
    billEvents([bill(1)], now)[0].id,
    billEvents([bill(1)], now + 1000)[0].id,
  );
  for (const elapsed of [0, 60000, 3600000]) {
    const ranks = rankWidgets(
      [finance],
      billEvents([bill(1), { ...bill(3), name: 'Rent' }], now + elapsed),
      now + elapsed,
    );
    assert.equal(ranks[0].score, elapsed >= 3600000 ? 60 : 50);
    assert.deepEqual(
      displayLayout(saved, attentionFor(initialPresentation(), ranks, saved)),
      saved,
    );
  }
});
