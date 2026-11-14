// Calls onIdle once ms pass without a touch(). Each touch() starts the wait over, so
// something is only "idle" after a full quiet spell. stop() cancels it for good.
export function idleTimer(ms: number, onIdle: () => void) {
  let handle: ReturnType<typeof setTimeout> | undefined;
  return {
    touch() {
      clearTimeout(handle);
      handle = setTimeout(onIdle, ms);
    },
    stop() {
      clearTimeout(handle);
    },
  };
}
