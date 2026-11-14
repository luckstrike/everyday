import assert from 'node:assert/strict';
import { after, before, test } from 'node:test';
import { createServer, type ViteDevServer } from 'vite';

let server: ViteDevServer;
let render: typeof import('svelte/server').render;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
let SettingsPanel: any;

before(async () => {
  server = await createServer({
    server: { middlewareMode: true, hmr: false, watch: null },
    appType: 'custom',
    optimizeDeps: { noDiscovery: true, include: [] },
  });
  ({ render } = await server.ssrLoadModule('svelte/server'));
  SettingsPanel = (
    await server.ssrLoadModule('/src/components/SettingsPanel.svelte')
  ).default;
});

after(async () => {
  await server?.close();
});

const noop = () => {};
const panel = (section: string, theme = 'evergreen') =>
  render(SettingsPanel, {
    props: {
      section,
      theme,
      onsection: noop,
      onclose: noop,
      onchoosetheme: noop,
      onbillschanged: noop,
    },
  }).body;

test('settings is a dialog with a heading, a Done button and every section on its rail', () => {
  const body = panel('bills');
  assert.match(body, /role="dialog"[^>]*aria-modal="true"/);
  assert.match(body, /<h2>Settings<\/h2>/);
  assert.match(body, />Done</);
  assert.match(body, /aria-label="Settings sections"/);
  assert.match(body, />Bills</);
  assert.match(body, />Appearance</);
});

test('the rail marks the section that is showing, and only that one', () => {
  for (const [section, label, other] of [
    ['bills', 'Bills', 'Appearance'],
    ['appearance', 'Appearance', 'Bills'],
  ]) {
    const body = panel(section);
    assert.match(body, new RegExp(`aria-current="page"[^>]*>${label}<`));
    assert.doesNotMatch(body, new RegExp(`aria-current="page"[^>]*>${other}<`));
    assert.equal(body.match(/aria-current="page"/g)?.length, 1);
  }
});

test('the bills section starts by saying it is loading, before anything is shown', () => {
  const body = panel('bills');
  assert.match(body, /Loading your bills/);
  assert.doesNotMatch(body, /class="bill-form"|Add bill/);
});

test('the appearance section offers each theme and marks the current one', () => {
  for (const [theme, current, other] of [
    ['evergreen', 'Evergreen', 'Turquoise'],
    ['turquoise', 'Turquoise', 'Evergreen'],
  ]) {
    const body = panel('appearance', theme);
    assert.match(
      body,
      new RegExp(`aria-pressed="true"[^>]*>[\\s\\S]*?${current}`),
    );
    assert.equal(body.match(/aria-pressed="true"/g)?.length, 1);
    assert.ok(body.includes(other));
  }
});
