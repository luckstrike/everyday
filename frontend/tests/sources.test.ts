import { test } from 'node:test';
import assert from 'node:assert/strict';
import { sourceNotice } from '../src/lib/sources.ts';

test('live and mock data need no notice, even when a list is empty', () => {
  assert.equal(sourceNotice('live', 'Calendar'), null);
  assert.equal(sourceNotice('mock', 'Calendar'), null);
});
test('a section with nothing hooked up says it is not connected', () => {
  assert.equal(sourceNotice('unconfigured', 'Bills')?.title, 'Not connected');
});
test('a source that is not delivering names the section', () => {
  assert.equal(
    sourceNotice('unavailable', 'Weather')?.title,
    'Weather unavailable',
  );
});
