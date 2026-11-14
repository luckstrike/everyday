import { test, mock } from 'node:test';
import assert from 'node:assert/strict';
import { idleTimer } from '../src/lib/idle.ts';

test('goes idle only after a full quiet spell', () => {
  mock.timers.enable({ apis: ['setTimeout'] });
  try {
    let idle = 0;
    const timer = idleTimer(1000, () => (idle += 1));
    timer.touch();
    mock.timers.tick(999);
    assert.equal(idle, 0);
    mock.timers.tick(1);
    assert.equal(idle, 1);
  } finally {
    mock.timers.reset();
  }
});
test('every touch starts the wait over', () => {
  mock.timers.enable({ apis: ['setTimeout'] });
  try {
    let idle = 0;
    const timer = idleTimer(1000, () => (idle += 1));
    timer.touch();
    mock.timers.tick(900);
    timer.touch();
    mock.timers.tick(900);
    assert.equal(idle, 0, 'still in use: the second touch reset the clock');
    mock.timers.tick(100);
    assert.equal(idle, 1);
  } finally {
    mock.timers.reset();
  }
});
test('a stopped timer never fires', () => {
  mock.timers.enable({ apis: ['setTimeout'] });
  try {
    let idle = 0;
    const timer = idleTimer(1000, () => (idle += 1));
    timer.touch();
    timer.stop();
    mock.timers.tick(5000);
    assert.equal(idle, 0);
  } finally {
    mock.timers.reset();
  }
});
