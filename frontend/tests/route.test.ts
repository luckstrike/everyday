import { test } from 'node:test';
import assert from 'node:assert/strict';
import { hashFor, parseHash, sections } from '../src/lib/route.ts';

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
test('settings opens on the first section unless a known one is named', () => {
  const first = sections[0].id;
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
});
test('an address always reads back as the screen it was made from', () => {
  assert.equal(hashFor({ name: 'dashboard' }), '');
  for (const { id } of sections) {
    const view = { name: 'settings', section: id } as const;
    assert.deepEqual(parseHash(hashFor(view)), view);
  }
});
