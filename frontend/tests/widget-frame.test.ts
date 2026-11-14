import assert from 'node:assert/strict';
import { after, before, test } from 'node:test';
import type { Component, Snippet } from 'svelte';
import { createServer, type ViteDevServer } from 'vite';
import { iconPaths } from '../src/lib/icon-paths.ts';
import { weatherIconName } from '../src/lib/weather.ts';

// One of each kind of weather the backend reports, for the forecast tests.
const weatherKinds = ['clear', 'rain', 'snow', 'storm', 'fog', 'cloudy'];
let server: ViteDevServer;
let render: typeof import('svelte/server').render;
let createRawSnippet: typeof import('svelte').createRawSnippet;
let WidgetFrame: Component<{
  title: string;
  message?: { title: string; detail: string } | null;
  actions?: Snippet;
  children: Snippet;
}>;

before(async () => {
  server = await createServer({
    server: { middlewareMode: true, hmr: false, watch: null },
    appType: 'custom',
    optimizeDeps: { noDiscovery: true, include: [] },
  });
  ({ render } = await server.ssrLoadModule('svelte/server'));
  ({ createRawSnippet } = await server.ssrLoadModule('svelte'));
  WidgetFrame = (
    await server.ssrLoadModule('/src/components/WidgetFrame.svelte')
  ).default;
});

after(async () => {
  await server?.close();
});

test('month and compact calendars retain date placement and today across different week counts', async () => {
  const { default: MonthCalendar } = await server.ssrLoadModule(
    '/src/components/MonthCalendar.svelte',
  );
  for (const { year, month, day, weeks } of [
    { year: 2026, month: 1, day: 14, weeks: 4 },
    { year: 2026, month: 7, day: 31, weeks: 6 },
    { year: 2024, month: 1, day: 29, weeks: 5 },
  ]) {
    for (const week of [false, true]) {
      const { body } = render(MonthCalendar, {
        props: { now: new Date(year, month, day, 12).getTime(), week },
      });
      const dates = body.match(/<tbody[^>]*>([\s\S]*?)<\/tbody>/)![1];
      assert.equal((dates.match(/<tr\b/g) ?? []).length, week ? 1 : weeks);
      assert.equal((body.match(/scope="col"/g) ?? []).length, 7);
      assert.equal((body.match(/aria-current="date"/g) ?? []).length, 1);
      const date = `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
      assert.match(
        body,
        new RegExp(`datetime="${date}"[^>]*aria-current="date"`),
      );
    }
  }
});

test('a widget renders its heading before content without requiring actions', () => {
  const children = createRawSnippet(() => ({
    render: () => '<p>No upcoming events.</p>',
  }));
  const { body } = render(WidgetFrame, {
    props: { title: 'Up next', children },
  });

  assert.match(body, /<h2[^>]*>Up next<\/h2>/);
  assert.doesNotMatch(body, /class="widget-actions/);
  assert.ok(body.indexOf('</header>') < body.indexOf('No upcoming events.'));
});

test('multiple actions stay in the header in supplied reading order', () => {
  const actions = createRawSnippet(() => ({
    render: () => '<button>Refresh</button><button>Days</button>',
  }));
  const children = createRawSnippet(() => ({
    render: () => '<p>Current conditions</p>',
  }));
  const { body } = render(WidgetFrame, {
    props: { title: 'Weather', actions, children },
  });
  const headerEnd = body.indexOf('</header>');

  assert.match(body, /class="widget-actions/);
  assert.ok(body.indexOf('Refresh') < body.indexOf('Days'));
  assert.ok(body.indexOf('Days') < headerEnd);
  assert.ok(headerEnd < body.indexOf('Current conditions'));
});

test('a message replaces content without rendering it or hiding header actions', () => {
  let contentRendered = false;
  const actions = createRawSnippet(() => ({
    render: () => '<button>Retry</button>',
  }));
  const children = createRawSnippet(() => {
    contentRendered = true;
    return { render: () => '<p>Normal content</p>' };
  });
  const { body } = render(WidgetFrame, {
    props: {
      title: 'Weather',
      message: {
        title: 'Weather unavailable',
        detail: 'Waiting for the latest data.',
      },
      actions,
      children,
    },
  });

  assert.match(body, /<h2[^>]*>Weather<\/h2>/);
  assert.match(body, /<button>Retry<\/button>/);
  assert.match(body, /Weather unavailable/);
  assert.match(body, /Waiting for the latest data\./);
  assert.match(body, /class="widget-content[^"\n]*\bhas-message\b/);
  assert.ok(body.indexOf('</header>') < body.indexOf('Weather unavailable'));
  assert.doesNotMatch(body, /Normal content/);
  assert.equal(contentRendered, false);
});

test('an explicit null message displays normal content', () => {
  const children = createRawSnippet(() => ({
    render: () => '<p>Current conditions</p>',
  }));
  const { body } = render(WidgetFrame, {
    props: { title: 'Weather', message: null, children },
  });

  assert.match(body, /Current conditions/);
  assert.doesNotMatch(body, /class="widget-title message-title/);
  assert.doesNotMatch(body, /has-message/);
});

for (const widget of [
  {
    name: 'Bills',
    component: 'FinanceWidget',
    sourceName: 'Bills',
    data: { bills: [] },
    unconfiguredTitle: 'No bills yet',
    emptyTitle: 'All clear',
    emptyDetail: 'No upcoming bills.',
  },
  {
    name: 'Up next',
    component: 'UpNextWidget',
    sourceName: 'Calendar',
    data: [],
    unconfiguredTitle: 'Not connected',
    emptyTitle: 'A little breathing room',
    emptyDetail: 'No upcoming events.',
  },
]) {
  test(`${widget.name} shows its empty message for working sources`, async () => {
    const { default: Widget } = await server.ssrLoadModule(
      `/src/components/widgets/${widget.component}.svelte`,
    );

    for (const status of ['live', 'mock']) {
      const { body } = render(Widget, {
        props: { data: widget.data, status, size: 'medium', now: 0 },
      });

      assert.ok(body.includes(widget.emptyTitle));
      assert.ok(body.includes(widget.emptyDetail));
      assert.match(body, /class="widget-content[^"\n]*\bhas-message\b/);
      assert.doesNotMatch(
        body,
        /Not connected|No bills yet|Waiting for the latest data|Check your bills file/,
      );
    }
  });

  test(`${widget.name} prioritizes source notices over empty results`, async () => {
    const { default: Widget } = await server.ssrLoadModule(
      `/src/components/widgets/${widget.component}.svelte`,
    );

    for (const status of ['unconfigured', 'unavailable']) {
      const { body } = render(Widget, {
        props: { data: widget.data, status, size: 'medium', now: 0 },
      });
      const expectedTitle =
        status === 'unconfigured'
          ? widget.unconfiguredTitle
          : `${widget.sourceName} unavailable`;

      assert.ok(body.includes(expectedTitle));
      assert.match(body, /class="widget-content[^"\n]*\bhas-message\b/);
      assert.ok(!body.includes(widget.emptyTitle));
      assert.ok(!body.includes(widget.emptyDetail));
    }
  });
}

test('Bills lists what is left to pay first, soonest due, then what is already paid', async () => {
  const { default: FinanceWidget } = await server.ssrLoadModule(
    '/src/components/widgets/FinanceWidget.svelte',
  );
  const now = new Date(2026, 10, 11, 12).getTime();
  const due = (month: number, date: number) =>
    new Date(2026, month - 1, date).toISOString();
  // Deliberately out of order. Phone is due soonest but already paid.
  const data = {
    bills: [
      {
        name: 'Rent',
        amount: 1850,
        dueAt: due(12, 1),
        paid: false,
        overdue: false,
        daysLate: 0,
      },
      {
        name: 'Phone',
        amount: 40,
        dueAt: due(11, 11),
        paid: true,
        overdue: false,
        daysLate: 0,
      },
      {
        name: 'Electricity',
        amount: 84.5,
        dueAt: due(11, 15),
        paid: false,
        overdue: false,
        daysLate: 0,
      },
      {
        name: 'Internet',
        amount: 65,
        dueAt: due(11, 20),
        paid: false,
        overdue: false,
        daysLate: 0,
      },
    ],
  };

  for (const size of ['small', 'medium', 'large']) {
    const { body } = render(FinanceWidget, {
      props: { data, status: 'live', size, now },
    });

    // Every bill is rendered, so the measuring code can decide how many fit.
    assert.equal(body.match(/class="bill[\s"]/g)?.length, 4);
    assert.doesNotMatch(body, /has-message/);
    const order = ['Electricity', 'Internet', 'Rent', 'Phone'].map((name) =>
      body.indexOf(`>${name}<`),
    );
    assert.ok(order.every((position) => position >= 0));
    assert.deepEqual(
      order,
      [...order].sort((a, b) => a - b),
    );
  }
  const { body } = render(FinanceWidget, {
    props: { data, status: 'live', size: 'medium', now },
  });
  assert.match(body, /\$1,850\.00/);
  assert.match(body, /\$84\.50/);
  // Only the paid bill is marked, and nothing mentions how bills are paid.
  assert.equal(body.match(/>Paid</g)?.length, 1);
  assert.equal(body.match(/class="bill[^"]*\bpaid\b/g)?.length, 1);
  assert.doesNotMatch(body, /Autopay|Manual payment/);
  // Nothing is late, so there is no chip and no warning.
  assert.doesNotMatch(body, /overdue/i);
});

test('Bills puts missed bills first, says how late, and counts them in the header', async () => {
  const { default: FinanceWidget } = await server.ssrLoadModule(
    '/src/components/widgets/FinanceWidget.svelte',
  );
  const now = new Date(2026, 10, 14, 12).getTime();
  const due = (month: number, date: number) =>
    new Date(2026, month - 1, date).toISOString();
  const data = {
    bills: [
      {
        name: 'Rent',
        amount: 1850,
        dueAt: due(12, 1),
        paid: false,
        overdue: false,
        daysLate: 0,
      },
      {
        name: 'Phone',
        amount: 40,
        dueAt: due(11, 13),
        paid: false,
        overdue: true,
        daysLate: 1,
      },
      {
        name: 'Electricity',
        amount: 84.5,
        dueAt: due(11, 10),
        paid: false,
        overdue: true,
        daysLate: 4,
      },
    ],
  };
  const { body } = render(FinanceWidget, {
    props: { data, status: 'live', size: 'medium', now },
  });

  // Oldest miss first, then what is still to pay; each bill appears once.
  const order = ['Electricity', 'Phone', 'Rent'].map((name) =>
    body.indexOf(`>${name}<`),
  );
  assert.deepEqual(
    order,
    [...order].sort((a, b) => a - b),
  );
  assert.equal(body.match(/class="bill[^"]*\boverdue\b/g)?.length, 2);
  assert.match(body, />4 days overdue</);
  assert.match(body, />1 day overdue</);
  assert.match(body, /2 overdue/);
});

test('Bills with no fixed amount say Due, and say nothing once paid', async () => {
  const { default: FinanceWidget } = await server.ssrLoadModule(
    '/src/components/widgets/FinanceWidget.svelte',
  );
  const now = new Date(2026, 10, 14, 12).getTime();
  const due = (date: number) => new Date(2026, 10, date).toISOString();
  const bill = (name: string, dueAt: string, paid = false, amount = 0) => ({
    name,
    amount,
    dueAt,
    paid,
    overdue: false,
    daysLate: 0,
  });
  const { body } = render(FinanceWidget, {
    props: {
      data: {
        bills: [
          bill('Amex', due(20)),
          bill('Visa', due(15), true),
          bill('Rent', due(25), false, 1850),
        ],
      },
      status: 'live',
      size: 'medium',
      now,
    },
  });
  // Only the unpaid card with no amount says Due; the real amount is still quoted.
  assert.equal(body.match(/>Due</g)?.length, 1);
  assert.match(body, /\$1,850\.00/);
  assert.doesNotMatch(body, /\$0\.00/);
});

test('Up next keeps all upcoming events available for measured fitting in date order', async () => {
  const { default: UpNextWidget } = await server.ssrLoadModule(
    '/src/components/widgets/UpNextWidget.svelte',
  );
  const now = Date.UTC(2026, 10, 11, 12);
  const hour = 60 * 60 * 1000;
  const data = Array.from({ length: 6 }, (_, index) => ({
    title: `Event ${index + 1}`,
    location: '',
    startsAt: new Date(now + index * hour).toISOString(),
    endsAt: new Date(now + (index + 1) * hour).toISOString(),
    allDay: false,
  })).reverse();
  data.push({
    title: 'Already ended',
    location: '',
    startsAt: new Date(now - hour).toISOString(),
    endsAt: new Date(now).toISOString(),
    allDay: false,
  });

  for (const size of ['small', 'medium', 'large']) {
    const { body } = render(UpNextWidget, {
      props: { data, status: 'live', size, now },
    });

    assert.equal(body.match(/class="event[\s"]/g)?.length, 6);
    assert.doesNotMatch(body, /Already ended|has-message/);
    for (let index = 1; index < 6; index += 1) {
      assert.ok(
        body.indexOf(`Event ${index}`) < body.indexOf(`Event ${index + 1}`),
      );
    }
  }
});

test('Weather replaces cached conditions with source notices and hides its toggle', async () => {
  const { default: WeatherWidget } = await server.ssrLoadModule(
    '/src/components/widgets/WeatherWidget.svelte',
  );
  const data = {
    location: 'Dallas',
    temperature: 72,
    condition: 'Clear',
    high: 75,
    low: 55,
    precipitation: 0,
    tomorrow: 'Clear',
    hourly: [],
    daily: [],
    alert: '',
  };

  for (const status of ['live', 'unconfigured', 'unavailable']) {
    const { body } = render(WeatherWidget, {
      props: { data, status, size: 'medium' },
    });

    assert.match(body, /<h2[^>]*>Weather<\/h2>/);
    if (status === 'live') {
      assert.match(body, /Clear/);
      assert.match(body, /Show hourly forecast/);
    } else {
      assert.ok(
        body.includes(
          status === 'unconfigured' ? 'Not connected' : 'Weather unavailable',
        ),
      );
      assert.doesNotMatch(body, /Clear|Show hourly forecast/);
    }
  }
});

test('temperature ranges show low before high with an accessible arrow alternative', async () => {
  const { default: TemperatureRange } = await server.ssrLoadModule(
    '/src/components/TemperatureRange.svelte',
  );

  for (const [low, high] of [
    [62, 78],
    [-12, -4],
    [89, 104],
  ]) {
    const { body } = render(TemperatureRange, { props: { low, high } });

    assert.ok(body.indexOf(`${low}°`) < body.indexOf(`${high}°`));
    assert.match(body, />Low <\/span>/);
    assert.match(body, /> to high <\/span>/);
    assert.match(body, /<svg[^>]*aria-hidden="true"/);
    assert.match(body, /M4 12h16m-6-6 6 6-6 6/);
  }
});

test('Weather shows five days with low-to-high ranges, starting with today', async () => {
  const { default: WeatherWidget } = await server.ssrLoadModule(
    '/src/components/widgets/WeatherWidget.svelte',
  );
  const data = {
    location: 'Dallas',
    temperature: 72,
    condition: 'Clear',
    high: 90,
    low: 55,
    precipitation: 0,
    tomorrow: 'Clear',
    hourly: [],
    daily: Array.from({ length: 6 }, (_, index) => ({
      date: new Date(Date.UTC(2026, 10, 11 + index, 12)).toISOString(),
      high: 80 + index,
      low: 60 + index,
      condition: `Weather ${index}`,
      icon: weatherKinds[index],
      precipitation: 0,
    })),
    alert: '',
  };
  const { body } = render(WeatherWidget, {
    props: { data, status: 'live', size: 'medium' },
  });
  // One piece of markup per forecast card, however deeply each card nests its content.
  const days = body.split(/(?=<div class="forecast-day(?:\s[^"]*)?")/).slice(1);

  assert.equal(days.length, 5);
  assert.match(days[0], /Today/);
  for (const [index, day] of days.entries()) {
    assert.match(day, /aria-hidden="true"/);
    assert.match(day, /visibility:\s*hidden/);
    assert.ok(day.includes(`${60 + index}°`));
    assert.ok(day.includes(`${80 + index}°`));
    assert.ok(day.indexOf(`${60 + index}°`) < day.indexOf(`${80 + index}°`));
  }
  assert.doesNotMatch(body, /65°|85°/);
});

test('Weather draws the icon for the current conditions, not the first forecast day', async () => {
  const { default: WeatherWidget } = await server.ssrLoadModule(
    '/src/components/widgets/WeatherWidget.svelte',
  );
  const data = {
    location: 'Dallas',
    temperature: 41,
    condition: 'Rain',
    icon: 'rain',
    high: 50,
    low: 38,
    precipitation: 80,
    tomorrow: 'Clear',
    hourly: [],
    // Today's forecast is clear, so a card can differ from what is happening now.
    daily: [
      {
        date: new Date(Date.UTC(2026, 10, 11, 12)).toISOString(),
        high: 50,
        low: 38,
        condition: 'Clear skies',
        icon: 'clear',
        precipitation: 0,
      },
    ],
    alert: '',
  };
  const { body } = render(WeatherWidget, {
    props: { data, status: 'live', size: 'medium' },
  });
  const current = body.match(
    /class="weather-icon[^"]*"[^>]*>([\s\S]*?)<\/span>/,
  );

  assert.ok(current, 'the current-conditions icon should render');
  assert.ok(current[1].includes(iconPaths[weatherIconName('rain')]));
  assert.ok(!current[1].includes(iconPaths[weatherIconName('clear')]));
});

test('each forecast day shows the icon for its own weather, with the condition for screen readers', async () => {
  const { default: WeatherWidget } = await server.ssrLoadModule(
    '/src/components/widgets/WeatherWidget.svelte',
  );
  const data = {
    location: 'Dallas',
    temperature: 72,
    condition: 'Clear',
    high: 90,
    low: 55,
    precipitation: 0,
    tomorrow: 'Clear',
    hourly: [],
    daily: weatherKinds.slice(0, 5).map((kind, index) => ({
      date: new Date(Date.UTC(2026, 10, 11 + index, 12)).toISOString(),
      high: 80,
      low: 60,
      condition: `Weather ${index}`,
      icon: kind,
      precipitation: 0,
    })),
    alert: '',
  };
  const { body } = render(WeatherWidget, {
    props: { data, status: 'live', size: 'medium' },
  });
  const days = body.split(/(?=<div class="forecast-day(?:\s[^"]*)?")/).slice(1);

  assert.equal(days.length, 5);
  for (const [index, day] of days.entries()) {
    assert.ok(
      day.includes(iconPaths[weatherIconName(weatherKinds[index])]),
      `day ${index} should draw its ${weatherKinds[index]} icon`,
    );
    assert.match(day, new RegExp(`sr-only[^>]*>Weather ${index}<`));
  }
  // Different kinds of weather must not all draw the same picture.
  const drawn = new Set(days.map((day) => day.match(/<path d="([^"]+)"/g)![0]));
  assert.ok(drawn.size > 1);
});
