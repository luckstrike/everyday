import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import {
  defaultTheme,
  loadTheme,
  THEME_KEY,
  themes,
} from '../src/lib/theme.ts';

const withStorage = (stored: string | null | 'throws', check: () => void) => {
  const original = Object.getOwnPropertyDescriptor(globalThis, 'localStorage');
  Object.defineProperty(globalThis, 'localStorage', {
    configurable: true,
    value: {
      getItem: (key: string) => {
        if (stored === 'throws') throw new Error('storage is blocked');
        return key === THEME_KEY ? stored : null;
      },
    },
  });
  try {
    check();
  } finally {
    if (original) Object.defineProperty(globalThis, 'localStorage', original);
    else delete (globalThis as { localStorage?: unknown }).localStorage;
  }
};

test('every theme loads back as itself', () => {
  for (const { id } of themes)
    withStorage(id, () => assert.equal(loadTheme(), id));
});
test('a missing, unknown or unreadable saved theme falls back to the default', () => {
  for (const stored of [null, '', 'nope', 'Ocean', 'throws'] as const)
    withStorage(stored, () => assert.equal(loadTheme(), defaultTheme));
});
test('the default theme is one that exists', () => {
  assert.ok(themes.some((t) => t.id === defaultTheme));
});
test('themes are listed alphabetically, with no repeats', () => {
  const labels = themes.map((t) => t.label);
  assert.deepEqual(
    labels,
    [...labels].sort((a, b) => a.localeCompare(b)),
  );
  assert.equal(new Set(themes.map((t) => t.id)).size, themes.length);
});

const css = readFileSync(new URL('../src/app.css', import.meta.url), 'utf8');
const palette = [
  '--bg',
  '--surface',
  '--accent',
  '--accent-hover',
  '--text',
  '--muted',
  '--line',
  '--placement-bg',
];
test('every theme defines the whole palette in app.css', () => {
  for (const { id } of themes) {
    const match = css.match(
      new RegExp(`\\[data-theme='${id}'\\]\\s*\\{([^}]*)\\}`),
    );
    assert.ok(match, `${id} has no block in app.css`);
    for (const name of palette)
      assert.ok(match[1].includes(`${name}:`), `${id} is missing ${name}`);
  }
});
