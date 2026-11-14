import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  draftFrom,
  dueNote,
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
  const good = { name: 'Rent', amount: '1850', day: 1, autopay: false };
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
  const entry = { name: 'Internet', amount: 65.5, day: 15, autopay: true };
  assert.deepEqual(entryFrom(draftFrom(entry)), entry);
});
test('the entry sent to the backend is tidied', () => {
  assert.deepEqual(
    entryFrom({ name: '  Rent  ', amount: '$1,850', day: 1, autopay: false }),
    { name: 'Rent', amount: 1850, day: 1, autopay: false },
  );
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
  assert.equal(dueNote(15), 'The 15th of every month.');
  assert.match(dueNote(28), /^The 28th of every month\.$/);
  assert.match(dueNote(29), /last day of a shorter month/);
  assert.match(dueNote(31), /last day of a shorter month/);
});
