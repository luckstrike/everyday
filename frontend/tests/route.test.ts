import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  defaultSection,
  hashFor,
  parseHash,
  sections,
} from '../src/lib/route.ts';

test('the dashboard is the default, whatever the address says', () => {
  for (const hash of [
    '',
    '#',
    '#/',
    '#/nonsense',
    '#/dashboard',
    '#settings2',
  ]) {
    assert.deepEqual(parseHash(hash), { name: 'dashboard' });
  }
});
test('settings opens on Bills unless a known section is named', () => {
  const first = defaultSection;
  assert.deepEqual(parseHash('#/settings'), {
    name: 'settings',
    section: first,
  });
  assert.deepEqual(parseHash('#/settings/nope'), {
    name: 'settings',
    section: first,
  });
  assert.deepEqual(parseHash('#/settings/appearance'), {
    name: 'settings',
    section: 'appearance',
  });
  assert.deepEqual(parseHash('#settings/bills'), {
    name: 'settings',
    section: 'bills',
  });
  assert.deepEqual(parseHash('#/settings/location'), {
    name: 'settings',
    section: 'location',
  });
});
test('an address always reads back as the screen it was made from', () => {
  assert.equal(hashFor({ name: 'dashboard' }), '');
  for (const { id } of sections) {
    const view = { name: 'settings', section: id } as const;
    assert.deepEqual(parseHash(hashFor(view)), view);
  }
});
test('the rail lists sections alphabetically, while Bills stays the default', () => {
  const labels = sections.map((s) => s.label);
  assert.deepEqual(
    labels,
    [...labels].sort((a, b) => a.localeCompare(b)),
  );
  assert.equal(defaultSection, 'bills');
});
