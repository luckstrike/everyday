import assert from 'node:assert/strict';
import { after, before, test } from 'node:test';
import type { Component, Snippet } from 'svelte';
import { createServer, type ViteDevServer } from 'vite';

let server: ViteDevServer;
let render: typeof import('svelte/server').render;
let createRawSnippet: typeof import('svelte').createRawSnippet;
let WidgetFrame: Component<{
  title: string;
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
