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
test('a widget can word its own notices, without changing the others', () => {
  const own = { title: 'No bills yet', detail: 'Add some to your bills file.' };
  const wording = { unconfigured: own, unavailableDetail: 'Check your file.' };
  assert.deepEqual(sourceNotice('unconfigured', 'Bills', wording), own);
  assert.deepEqual(sourceNotice('unavailable', 'Bills', wording), {
    title: 'Bills unavailable',
    detail: 'Check your file.',
  });
  assert.equal(sourceNotice('live', 'Bills', wording), null);
  assert.equal(sourceNotice('unconfigured', 'Bills')?.title, 'Not connected');
  assert.equal(
    sourceNotice('unavailable', 'Bills')?.detail,
    'Waiting for the latest data.',
  );
});
