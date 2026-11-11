import assert from 'node:assert/strict';
import { after, before, test } from 'node:test';
import type { Component, Snippet } from 'svelte';
import { createServer, type ViteDevServer } from 'vite';

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
});

for (const widget of [
  {
    name: 'Bills',
    component: 'FinanceWidget',
    sourceName: 'Bills',
    data: { bills: [] },
    emptyTitle: 'All clear',
    emptyDetail: 'No upcoming bills.',
  },
  {
    name: 'Up next',
    component: 'UpNextWidget',
    sourceName: 'Calendar',
    data: [],
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
      assert.doesNotMatch(body, /Not connected|Waiting for the latest data/);
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
          ? 'Not connected'
          : `${widget.sourceName} unavailable`;

      assert.ok(body.includes(expectedTitle));
      assert.ok(!body.includes(widget.emptyTitle));
      assert.ok(!body.includes(widget.emptyDetail));
    }
  });
}

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
