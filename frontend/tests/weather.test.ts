import { test } from 'node:test';
import assert from 'node:assert/strict';
import { weatherIconName } from '../src/lib/weather.ts';
import { iconPaths } from '../src/lib/icon-paths.ts';

const kinds = [
  'clear',
  'partly-cloudy',
  'cloudy',
  'fog',
  'rain',
  'snow',
  'storm',
];

test('every kind of weather has its own drawing', () => {
  const names = kinds.map(weatherIconName);
  assert.equal(new Set(names).size, kinds.length);
  for (const name of names) {
    assert.ok(iconPaths[name], `${name} has no path to draw`);
  }
});
test('a kind of weather the display does not know shows a cloud', () => {
  assert.equal(weatherIconName('hail-meteor'), 'weather-cloudy');
  assert.equal(weatherIconName(''), 'weather-cloudy');
});
