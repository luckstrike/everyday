import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  type ListedBill,
  toEntry,
  withPaid,
  draftFrom,
  dueNote,
  groupByDay,
  emptyDraft,
  entryFrom,
  ordinal,
  parseAmount,
  validateDraft,
} from '../src/lib/bills-editor.ts';

test('amounts are read the way people write them', () => {
  assert.equal(parseAmount('1850'), 1850);
  assert.equal(parseAmount('84.5'), 84.5);
  assert.equal(parseAmount('$84.50'), 84.5);
  assert.equal(parseAmount(' 1,850.25 '), 1850.25);
  assert.equal(parseAmount('.5'), 0.5);
  assert.equal(parseAmount('0'), 0);
  // Cents are rounded, so 10.005 style typing cannot store a fraction of a cent.
  assert.equal(parseAmount('19.999'), 20);
});
test('things that are not amounts are refused', () => {
  for (const bad of [
    '',
    ' ',
    'abc',
    '-5',
    '1e3',
    '12 dollars',
    '1.2.3',
    '$',
    '--1',
  ]) {
    assert.equal(parseAmount(bad), null, JSON.stringify(bad));
  }
});
test('a draft can be saved only with a name, an amount and a real day', () => {
  const good = { name: 'Rent', amount: '1850', day: 1 };
  assert.equal(validateDraft(good), null);
  assert.match(validateDraft({ ...good, name: '   ' })!, /name/i);
  assert.match(validateDraft({ ...good, amount: '' })!, /amount/i);
  assert.match(validateDraft({ ...good, amount: '-5' })!, /amount/i);
  assert.match(validateDraft({ ...good, day: 0 })!, /day/i);
  assert.match(validateDraft({ ...good, day: 32 })!, /day/i);
  assert.match(validateDraft({ ...good, day: 1.5 })!, /day/i);
});
test('a new draft is not saveable until it is filled in', () => {
  assert.notEqual(validateDraft(emptyDraft()), null);
});
test('a bill survives being turned into a draft and back', () => {
  const entry = { name: 'Internet', amount: 65.5, day: 15 };
  assert.deepEqual(entryFrom(draftFrom(entry)), entry);
});
test('the entry sent to the backend is tidied', () => {
  assert.deepEqual(entryFrom({ name: '  Rent  ', amount: '$1,850', day: 1 }), {
    name: 'Rent',
    amount: 1850,
    day: 1,
  });
});
test('due days read aloud correctly, including the teens', () => {
  const spoken = [1, 2, 3, 4, 11, 12, 13, 21, 22, 23, 30, 31].map(ordinal);
  assert.deepEqual(spoken, [
    '1st',
    '2nd',
    '3rd',
    '4th',
    '11th',
    '12th',
    '13th',
    '21st',
    '22nd',
    '23rd',
    '30th',
    '31st',
  ]);
});
test('the due note warns about short months only when it matters', () => {
  assert.equal(dueNote(15), '15th of every month.');
  assert.equal(dueNote(28), '28th of every month.');
  assert.match(dueNote(29), /last day of a shorter month/);
  assert.match(dueNote(31), /last day of a shorter month/);
});
test('the due note does not put "the" before the day', () => {
  for (const day of [1, 2, 15, 22, 29, 31]) {
    assert.doesNotMatch(dueNote(day), /^the\s/i);
  }
});
test('a bill is prefilled with its cents, so 65 reads 65.00', () => {
  const draft = draftFrom({ name: 'Internet', amount: 65, day: 15 });
  assert.equal(draft.amount, '65.00');
  assert.equal(draftFrom({ name: 'X', amount: 84.5, day: 1 }).amount, '84.50');
  // And it still reads back as the same amount.
  assert.equal(entryFrom(draft).amount, 65);
});
const bill = (name: string, day: number) => ({
  name,
  amount: 10,
  day,
});
test('bills are grouped by due day, earliest first, however they were saved', () => {
  const groups = groupByDay([
    bill('Rent', 25),
    bill('Phone', 1),
    bill('Water', 15),
  ]);
  assert.deepEqual(
    groups.map((g) => [g.day, g.bills.map((b) => b.bill.name)]),
    [
      [1, ['Phone']],
      [15, ['Water']],
      [25, ['Rent']],
    ],
  );
});
test('bills due on the same day share a group, in name order', () => {
  const groups = groupByDay([
    bill('Water', 1),
    bill('rent', 1),
    bill('Internet', 15),
    bill('Electricity', 1),
  ]);
  assert.deepEqual(
    groups.map((g) => [g.day, g.bills.map((b) => b.bill.name)]),
    [
      [1, ['Electricity', 'rent', 'Water']],
      [15, ['Internet']],
    ],
  );
});
test('each grouped bill still points at its place in the saved list', () => {
  const saved = [bill('Rent', 25), bill('Phone', 1), bill('Water', 15)];
  const placed = groupByDay(saved).flatMap((g) => g.bills);
  for (const { bill: b, index } of placed) assert.equal(saved[index], b);
  assert.deepEqual(
    placed.map((p) => p.index),
    [1, 2, 0],
  );
});
test('screen order runs straight through the groups, for alternating fills', () => {
  const placed = groupByDay([bill('A', 1), bill('B', 1), bill('C', 9)]).flatMap(
    (g) => g.bills,
  );
  assert.deepEqual(
    placed.map((p) => p.position),
    [0, 1, 2],
  );
});
test('no bills means no groups', () => {
  assert.deepEqual(groupByDay([]), []);
});

const listed = (over: Partial<ListedBill> = {}): ListedBill => ({
  name: 'Electricity',
  amount: 84.5,
  day: 18,
  due: '2026-11-18',
  paid: false,
  ...over,
});

test('ticking a bill records the due date it was paid for', () => {
  assert.deepEqual(withPaid(listed(), true), {
    name: 'Electricity',
    amount: 84.5,
    day: 18,
    paidFor: '2026-11-18',
  });
});
test('unticking forgets the paid date, and leaves nothing behind in the file', () => {
  const paid = listed({ paid: true, paidFor: '2026-11-18' });
  const unticked = withPaid(paid, false);
  assert.equal('paidFor' in unticked, false);
  assert.deepEqual(unticked, { name: 'Electricity', amount: 84.5, day: 18 });
});
test('what is saved never includes the dates worked out for display', () => {
  const entry = toEntry(listed({ paid: true, paidFor: '2026-11-18' }));
  assert.deepEqual(Object.keys(entry).sort(), [
    'amount',
    'day',
    'name',
    'paidFor',
  ]);
  assert.deepEqual(Object.keys(toEntry(listed())).sort(), [
    'amount',
    'day',
    'name',
  ]);
});
test('editing a bill keeps it paid, so fixing a typo does not un-pay it', () => {
  const existing = {
    name: 'Electrcity',
    amount: 84.5,
    day: 18,
    paidFor: '2026-11-18',
  };
  const fixed = entryFrom(
    { name: 'Electricity', amount: '84.50', day: 18 },
    existing,
  );
  assert.equal(fixed.paidFor, '2026-11-18');
  // A brand-new bill starts unpaid.
  assert.equal(
    'paidFor' in entryFrom({ name: 'New', amount: '1', day: 1 }),
    false,
  );
});
