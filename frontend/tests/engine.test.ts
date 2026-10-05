import { test } from 'node:test';
import assert from 'node:assert/strict';
import { eventModifier, rankScreens } from '../src/lib/priority.ts';
import { initialSession, selectScreen } from '../src/lib/screens.ts';
import {
  defaultLayouts,
  presentationLayout,
  parseLayouts,
  canPlace,
  addWidget,
} from '../src/lib/layout.ts';
import type {
  AttentionEvent,
  ScreenDefinition,
  WidgetDefinition,
} from '../src/lib/types.ts';
const screens: ScreenDefinition[] = [
  ['home', 10],
  ['weather', 20],
  ['sports', 20],
  ['finance', 25],
].map(([id, basePriority]) => ({
  id,
  basePriority,
  title: id,
  cooldownMs: id === 'home' ? 0 : 65000,
  minDurationMs: 10000,
  maxDurationMs: 25000,
})) as ScreenDefinition[];
const widgets: WidgetDefinition[] = [
  'clock',
  'weather',
  'sports',
  'finance',
  'calendar',
].map((id) => ({
  id,
  title: id,
  basePriority:
    id === 'clock' ? 10 : id === 'calendar' || id === 'finance' ? 25 : 20,
  minWidth: 3,
  minHeight: 2,
  defaultWidth: 4,
  defaultHeight: 3,
})) as WidgetDefinition[];
const event = (
  widget = 'weather',
  kind = 'rain',
  decay = false,
): AttentionEvent => ({
  id: `${widget}:${kind}`,
  widget: widget as AttentionEvent['widget'],
  kind,
  startsAt: new Date(0).toISOString(),
  expiresAt: new Date(120000).toISOString(),
  decay,
});
const rank = (now: number, events: AttentionEvent[] = []) =>
  rankScreens(screens, widgets, events, now);
test('base priorities and event modifiers are deterministic', () => {
  assert.deepEqual(
    rank(0).map((c) => [c.id, c.score]),
    [
      ['finance', 25],
      ['sports', 20],
      ['weather', 20],
      ['home', 10],
    ],
  );
  assert.equal(rank(0, [event('sports', 'live')])[0].score, 60);
  assert.equal(rank(0, [event('finance', 'tomorrow')])[0].score, 50);
  assert.equal(rank(0, [event('calendar', 'soon')])[0].id, 'home');
  assert.equal(rank(0, [event('calendar', 'soon')])[0].score, 75);
});
test('events do not start early, expire exactly, and decay linearly', () => {
  const final = event('sports', 'final', true);
  assert.equal(eventModifier(final, -1), 0);
  assert.equal(eventModifier(final, 0), 20);
  assert.equal(eventModifier(final, 60000), 10);
  assert.equal(eventModifier(final, 120000), 0);
  assert.equal(eventModifier({ ...final, expiresAt: 'invalid' }, 10), 0);
});
test('emphasis and fullscreen restore without changing saved layout', () => {
  const saved = defaultLayouts().home,
    copy = JSON.stringify(saved);
  const normal = rank(0).find((c) => c.id === 'home');
  const severe = rank(0, [event('weather', 'severe')])[0];
  assert.equal(severe.presentation, 'FULLSCREEN');
  assert.deepEqual(presentationLayout(saved, severe), [
    { id: 'weather', x: 0, y: 0, w: 12, h: 6 },
  ]);
  assert.equal(rank(0, [event()])[0].presentation, 'EMPHASIZED');
  assert.equal(JSON.stringify(saved), copy);
  assert.deepEqual(presentationLayout(saved, normal), saved);
  assert.deepEqual(
    presentationLayout(
      saved,
      rank(120000, [event('weather', 'severe')]).find((c) => c.id === 'home'),
    ),
    saved,
  );
});
test('minimum reading time is respected, even for severe weather', () => {
  const initial = initialSession(0);
  assert.equal(
    selectScreen(
      initial,
      rank(9999, [event('weather', 'severe')]),
      screens,
      9999,
    ),
    initial,
  );
  assert.equal(
    selectScreen(
      initial,
      rank(10000, [event('weather', 'severe')]),
      screens,
      10000,
    ).current,
    'weather',
  );
});
test('maximum duration forces Home even with an ongoing severe alert', () => {
  let session = selectScreen(
    initialSession(0),
    rank(10000, [event('weather', 'severe')]),
    screens,
    10000,
  );
  session = selectScreen(
    session,
    rank(35000, [event('weather', 'severe')]),
    screens,
    35000,
  );
  assert.equal(session.current, 'home');
  assert.equal(
    selectScreen(
      session,
      rank(45000, [event('weather', 'severe')]),
      screens,
      45000,
    ).current,
    'home',
  );
});
test('cooldown is measured from exit and a new urgent event can bypass it once', () => {
  const initial = { ...initialSession(0), lastShown: { weather: 0 } };
  assert.equal(
    selectScreen(initial, rank(10000, [event()]), screens, 10000).current,
    'home',
  );
  const severe = event('weather', 'severe');
  const selected = selectScreen(initial, rank(10000, [severe]), screens, 10000);
  assert.equal(selected.current, 'weather');
  const home = selectScreen(selected, rank(35000, [severe]), screens, 35000);
  assert.equal(home.lastShown.weather, 35000);
  assert.equal(
    selectScreen(home, rank(45000, [severe]), screens, 45000).current,
    'home',
  );
  assert.equal(
    selectScreen(
      home,
      rank(45000, [{ ...severe, id: 'new-alert' }]),
      screens,
      45000,
    ).current,
    'weather',
  );
});
test('cleared attention returns to Home after minimum duration', () => {
  const selected = selectScreen(
    initialSession(0),
    rank(10000, [event()]),
    screens,
    10000,
  );
  assert.equal(selected.current, 'weather');
  assert.equal(
    selectScreen(selected, rank(15000), screens, 15000).current,
    'weather',
  );
  assert.equal(
    selectScreen(selected, rank(20000), screens, 20000).current,
    'home',
  );
});
test('normal rotation includes all screens and regularly rests at Home', () => {
  let session = initialSession(0);
  const visited: string[] = ['home'];
  for (let now = 1000; now <= 3600000; now += 1000) {
    const next = selectScreen(session, rank(now), screens, now);
    if (next.current !== session.current) {
      assert.ok(now - session.since >= 10000);
      if (session.current !== 'home') assert.equal(next.current, 'home');
      visited.push(next.current);
    }
    session = next;
  }
  assert.deepEqual(
    new Set(visited),
    new Set(['home', 'weather', 'sports', 'finance']),
  );
  assert.ok(visited.length > 100);
  assert.ok(session.bypassed.length <= 1);
});
test('saved layout validation protects bounds, collisions, unknown IDs and duplicates', () => {
  const layouts = defaultLayouts();
  assert.deepEqual(parseLayouts(JSON.stringify(layouts), widgets), layouts);
  for (const bad of [
    { ...layouts.home[0], w: 13 },
    { ...layouts.home[0], x: -1 },
    { ...layouts.home[0], id: 'oops' },
    { ...layouts.home[0], x: 7 },
  ])
    assert.throws(() =>
      parseLayouts(
        JSON.stringify({ ...layouts, home: [bad, ...layouts.home.slice(1)] }),
        widgets,
      ),
    );
  assert.throws(() =>
    parseLayouts(
      JSON.stringify({ ...layouts, home: [layouts.home[0], layouts.home[0]] }),
      widgets,
    ),
  );
  assert.equal(
    canPlace({ ...layouts.home[0], w: 6 }, layouts.home, widgets),
    true,
  );
  const removed = layouts.home.filter((p) => p.id !== 'sports');
  assert.equal(
    addWidget(
      removed,
      widgets.find((w) => w.id === 'sports')!,
    )?.length,
    5,
  );
});
