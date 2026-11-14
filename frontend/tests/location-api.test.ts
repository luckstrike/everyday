import assert from 'node:assert/strict';
import { afterEach, test } from 'node:test';
import {
  ApiError,
  getLocation,
  saveLocation,
  searchLocations,
} from '../src/lib/api.ts';

const originalFetch = globalThis.fetch;
afterEach(() => {
  globalThis.fetch = originalFetch;
});

test('location search encodes cities and qualifiers without changing the API path', async () => {
  globalThis.fetch = async (input) => {
    assert.equal(input, '/api/locations?q=S%C3%A3o%20Paulo%2C%20Brazil');
    return Response.json({ locations: [] });
  };
  assert.deepEqual(await searchLocations('São Paulo, Brazil'), {
    locations: [],
  });
});

test('saving location sends coordinates as JSON and reads the saved location', async () => {
  const location = {
    name: 'Austin, Texas, United States',
    latitude: 30.27,
    longitude: -97.74,
  };
  globalThis.fetch = async (input, init) => {
    assert.equal(input, '/api/location');
    assert.equal(init?.method, 'PUT');
    assert.deepEqual(init?.headers, { 'Content-Type': 'application/json' });
    assert.deepEqual(JSON.parse(init?.body as string), location);
    return Response.json({ location, editable: true });
  };
  assert.deepEqual(await saveLocation(location), { location, editable: true });
});

test('location settings surface server explanations instead of claiming a successful save', async () => {
  globalThis.fetch = async () =>
    Response.json(
      { error: 'Location is set by environment variables.' },
      { status: 409 },
    );
  await assert.rejects(
    getLocation(),
    (error: unknown) =>
      error instanceof ApiError &&
      error.status === 409 &&
      error.message === 'Location is set by environment variables.',
  );
});
