import { test } from 'node:test';
import assert from 'node:assert/strict';
import { billEvents, daysUntil } from '../src/lib/bills.ts';
import { eventModifier, rankScreens } from '../src/lib/priority.ts';
import { displayDuration, selectScreen } from '../src/lib/screens.ts';
import {
  defaultLayouts,
  pinnedLayout,
  portraitLayout,
  overlaps,
  parsePins,
  presentationLayout,
} from '../src/lib/layout.ts';
import type { ScreenDefinition } from '../src/lib/types.ts';
const now = new Date(2026, 9, 4, 23, 50).getTime();
const bill = (days: number) => ({
  name: 'Electricity',
  amount: 84.5,
  autopay: true,
  dueAt: new Date(2026, 9, 4 + days, 0, 10).toISOString(),
});
const finance: ScreenDefinition = {
  id: 'finance',
  title: 'Bills',
  basePriority: 25,
  cooldownMs: 65000,
  minDurationMs: 10000,
  maxDurationMs: 25000,
};
const home: ScreenDefinition = {
  ...finance,
  id: 'home',
  title: 'Home',
  basePriority: 10,
  cooldownMs: 0,
};
const weather: ScreenDefinition = {
  ...finance,
  id: 'weather',
  title: 'Weather',
  basePriority: 20,
};
test('bills gain priority by calendar day, including midnight and overdue', () => {
  assert.equal(daysUntil(bill(1).dueAt, now), 1);
  for (const [days, score, duration] of [
    [8, 0, 25000],
    [7, 5, 45000],
    [3, 15, 60000],
    [1, 25, 90000],
    [0, 35, 120000],
    [-1, 40, 120000],
  ]) {
    const events = billEvents([bill(days)], now);
    assert.equal(
      events.reduce((sum, e) => sum + eventModifier(e, now), 0),
      score,
    );
    const candidate = rankScreens([finance], [], events, now)[0];
    assert.equal(displayDuration(finance, candidate), duration);
  }
  assert.deepEqual(billEvents([{ ...bill(1), dueAt: 'invalid' }], now), []);
});
test('date priority is deterministic and competing finance events do not accumulate', () => {
  const events = billEvents([bill(1), { ...bill(3), name: 'Rent' }], now);
  const ranks = rankScreens([finance], [], events, now);
  assert.equal(ranks[0].score, 50);
  assert.equal(
    billEvents([bill(1)], now)[0].id,
    billEvents([bill(1)], now + 1000)[0].id,
  );
});
test('important bill visits last longer but return Home and yield to severe weather', () => {
  const defs = [home, finance, weather],
    events = billEvents([bill(1)], now);
  const session = {
    current: 'finance' as const,
    since: now,
    lastShown: { finance: now },
    bypassed: [],
  };
  assert.equal(
    selectScreen(
      session,
      rankScreens(defs, [], events, now + 30000),
      defs,
      now + 30000,
    ).current,
    'finance',
  );
  assert.equal(
    selectScreen(
      session,
      rankScreens(defs, [], events, now + 90000),
      defs,
      now + 90000,
    ).current,
    'home',
  );
  const severe = {
    id: 'severe',
    widget: 'weather' as const,
    kind: 'severe',
    startsAt: new Date(now).toISOString(),
    expiresAt: new Date(now + 120000).toISOString(),
    decay: false,
  };
  assert.equal(
    selectScreen(
      session,
      rankScreens(defs, [], [...events, severe], now + 11000),
      defs,
      now + 11000,
    ).current,
    'weather',
  );
});
test('pins persist across screens without duplicate widgets or saved geometry changes', () => {
  const layouts = defaultLayouts(),
    snapshot = JSON.stringify(layouts);
  for (const items of Object.values(layouts)) {
    const projected = pinnedLayout(items, ['finance', 'sports']);
    assert.ok(projected.some((item) => item.id === 'finance'));
    assert.ok(projected.some((item) => item.id === 'sports'));
    assert.equal(
      new Set(projected.map((item) => item.id)).size,
      projected.length,
    );
    for (const item of projected)
      assert.ok(
        !projected.some((other) => other !== item && overlaps(item, other)),
      );
  }
  assert.equal(JSON.stringify(layouts), snapshot);
  assert.deepEqual(pinnedLayout(layouts.weather, []), layouts.weather);
  assert.deepEqual(parsePins('["finance"]'), ['finance']);
  assert.throws(() => parsePins('["unknown"]'));
  assert.throws(() => parsePins('["finance","finance"]'));
});
test('portrait and narrow projections fit their columns and preserve layout restoration', () => {
  const saved = defaultLayouts().home,
    snapshot = JSON.stringify(saved);
  for (const narrow of [false, true]) {
    const projected = portraitLayout(pinnedLayout(saved, ['finance']), narrow);
    assert.equal(projected.length, 5);
    for (const item of projected) {
      assert.ok(item.x + item.w <= 12);
      assert.ok(
        !projected.some((other) => other !== item && overlaps(item, other)),
      );
    }
  }
  assert.equal(JSON.stringify(saved), snapshot);
  const alert = rankScreens(
    [weather],
    [],
    [
      {
        id: 'severe',
        widget: 'weather',
        kind: 'severe',
        startsAt: new Date(now).toISOString(),
        expiresAt: new Date(now + 60000).toISOString(),
        decay: false,
      },
    ],
    now,
  )[0];
  assert.equal(
    pinnedLayout(presentationLayout(saved, alert), ['finance'], true).length,
    1,
  );
  assert.deepEqual(pinnedLayout(saved, []), saved);
});

test('persistent bill reminders do not starve the other screens', () => {
  const sports = { ...weather, id: 'sports' as const };
  const defs = [{ ...home, maxDurationMs: 20000 }, finance, weather, sports];
  for (const due of [7, 1]) {
    let session: import('../src/lib/screens.ts').ScreenSession = {
      current: 'home',
      since: now,
      lastShown: { home: now },
      bypassed: [],
    };
    const seen = new Set(['home']);
    for (let elapsed = 1000; elapsed < 3600000; elapsed += 1000) {
      const time = now + elapsed;
      session = selectScreen(
        session,
        rankScreens(defs, [], billEvents([bill(due)], time), time),
        defs,
        time,
      );
      seen.add(session.current);
    }
    assert.deepEqual(seen, new Set(['home', 'finance', 'weather', 'sports']));
  }
});
