import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  eventModifier,
  rankWidgets,
  type Candidate,
} from '../src/lib/priority.ts';
import {
  initialPresentation,
  type PresentationSession,
  updatePresentation,
  attentionFor,
  displayLayout,
} from '../src/lib/presentation.ts';
import { initialBoards, widgetIds } from '../src/lib/boards.ts';
import type { AttentionEvent, WidgetDefinition } from '../src/lib/types.ts';
const widgets = widgetIds.map((id) => ({
  id,
  title: id,
  basePriority:
    id === 'finance' || id === 'calendar' ? 25 : id === 'clock' ? 10 : 20,
  minWidth: 6,
  minHeight: 2,
  defaultWidth: 4,
  defaultHeight: 3,
})) satisfies WidgetDefinition[];
const event = (
  widget: AttentionEvent['widget'] = 'weather',
  kind = 'rain',
  decay = false,
): AttentionEvent => ({
  id: `${widget}:${kind}`,
  widget,
  kind,
  startsAt: new Date(0).toISOString(),
  expiresAt: new Date(120000).toISOString(),
  decay,
});
const rank = (now: number, events: AttentionEvent[] = []) =>
  rankWidgets(widgets, events, now);
const updateAlertPresentation = (
  session: PresentationSession,
  ranks: Candidate[],
  now: number,
) => updatePresentation(session, ranks, now, true);
test('ambient policy keeps severe alerts in place and does not add absent widgets', () => {
  const saved = initialBoards().boards[0].layouts.portrait;
  const original = JSON.stringify(saved);
  const ranks = rank(0, [event('weather', 'severe')]);
  const session = updatePresentation(initialPresentation(), ranks, 0);
  assert.equal(session.active, undefined);
  assert.equal(attentionFor(session, ranks, saved)?.presentation, 'EMPHASIZED');
  assert.deepEqual(
    displayLayout(saved, attentionFor(session, ranks, saved)),
    saved,
  );
  assert.equal(attentionFor(session, ranks, []), undefined);
  assert.equal(
    attentionFor(session, rank(120000, [event('weather', 'severe')]), saved),
    undefined,
  );
  assert.equal(JSON.stringify(saved), original);
  const active = updateAlertPresentation(initialPresentation(), ranks, 0);
  assert.equal(updatePresentation(active, ranks, 1).active, undefined);
});
test('widget priorities are deterministic and strongest event wins without accumulation', () => {
  assert.equal(rank(0)[0].id, 'calendar');
  assert.equal(rank(0, [event('sports', 'live')])[0].score, 60);
  assert.equal(
    rank(0, [event('finance', 'tomorrow'), event('finance', 'soon')])[0].score,
    50,
  );
  assert.equal(rank(0, [event('calendar', 'soon')])[0].score, 75);
  assert.deepEqual(
    rank(0, [event('finance', 'soon'), event('finance', 'tomorrow')]),
    rank(0, [event('finance', 'tomorrow'), event('finance', 'soon')]),
  );
});
test('events do not start early, expire exactly and decay linearly', () => {
  const final = event('sports', 'final', true);
  assert.equal(eventModifier(final, -1), 0);
  assert.equal(eventModifier(final, 0), 20);
  assert.equal(eventModifier(final, 60000), 10);
  assert.equal(eventModifier(final, 120000), 0);
  assert.equal(eventModifier({ ...final, expiresAt: 'invalid' }, 10), 0);
});
test('opt-in fullscreen preserves geometry and restores after minimum duration', () => {
  const saved = initialBoards().boards[0].layouts.landscape,
    copy = JSON.stringify(saved);
  let session = updateAlertPresentation(
    initialPresentation(),
    rank(0, [event('weather', 'severe')]),
    0,
  );
  assert.deepEqual(
    displayLayout(saved, attentionFor(session, rank(0), saved)),
    [{ id: 'weather', x: 0, y: 0, w: 24, h: 24 }],
  );
  assert.equal(updateAlertPresentation(session, rank(9999), 9999), session);
  session = updateAlertPresentation(session, rank(10000), 10000);
  assert.deepEqual(
    displayLayout(saved, attentionFor(session, rank(10000), saved)),
    saved,
  );
  assert.deepEqual(
    displayLayout(
      saved,
      attentionFor(session, rank(10000, [event('finance', 'tomorrow')]), saved),
    ),
    saved,
  );
  assert.equal(JSON.stringify(saved), copy);
});
test('opt-in fullscreen is bounded, consumed alerts do not repeat and new alerts respect cooldown', () => {
  const severe = event('weather', 'severe');
  let session = updateAlertPresentation(
    initialPresentation(),
    rank(0, [severe]),
    0,
  );
  assert.ok(session.active);
  session = updateAlertPresentation(session, rank(25000, [severe]), 25000);
  assert.equal(session.active, undefined);
  assert.equal(
    updateAlertPresentation(session, rank(90000, [severe]), 90000),
    session,
  );
  const next = {
    ...severe,
    id: 'second',
    startsAt: new Date(25000).toISOString(),
  };
  assert.equal(
    updateAlertPresentation(session, rank(89999, [next]), 89999),
    session,
  );
  assert.ok(
    updateAlertPresentation(session, rank(90000, [next]), 90000).active,
  );
});
test('emphasis ignores absent widgets; opt-in fullscreen may show alerts', () => {
  const session = initialPresentation();
  assert.equal(
    attentionFor(session, rank(0, [event('finance', 'tomorrow')]), []),
    undefined,
  );
  const active = updateAlertPresentation(
    session,
    rank(0, [event('weather', 'severe')]),
    0,
  );
  assert.equal(
    displayLayout([], attentionFor(active, rank(0), []))[0].id,
    'weather',
  );
});
