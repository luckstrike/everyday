<script lang="ts">
  import Icon from './Icon.svelte';
  let {
    message,
    token = 0,
    duration = 4500,
    ondismiss,
  }: {
    message: string;
    token?: number;
    duration?: number;
    ondismiss: () => void;
  } = $props();
  let timer: ReturnType<typeof setTimeout> | undefined;
  let remaining = 0,
    started = 0,
    hovered = false,
    focused = false;
  function schedule() {
    clearTimeout(timer);
    if (!hovered && !focused) {
      started = performance.now();
      timer = setTimeout(ondismiss, remaining);
    }
  }
  function pause() {
    clearTimeout(timer);
    remaining = Math.max(0, remaining - (performance.now() - started));
  }
  $effect(() => {
    // The token restarts repeated messages, even when their text is identical.
    void token;
    if (!message) {
      hovered = false;
      focused = false;
      return;
    }
    remaining = duration;
    schedule();
    return () => clearTimeout(timer);
  });
</script>

{#if message}
  <aside
    class="toast"
    aria-label="Notification"
    onpointerenter={() => {
      if (!focused) pause();
      hovered = true;
    }}
    onpointerleave={() => {
      hovered = false;
      if (!focused) schedule();
    }}
    onfocusin={() => {
      if (!hovered) pause();
      focused = true;
    }}
    onfocusout={() => {
      focused = false;
      if (!hovered) schedule();
    }}
  >
    <span role="status" aria-live="polite" aria-atomic="true">{message}</span
    ><button class="icon-button" aria-label="Dismiss notice" onclick={ondismiss}
      ><Icon name="close" size={16} /></button
    >
  </aside>
{/if}

<style>
  .toast {
    position: fixed;
    left: 48px;
    bottom: 90px;
    max-width: min(560px, calc(100vw - 96px));
    background: var(--surface);
    border: 1px solid var(--line);
    padding: 8px 12px 8px 20px;
    border-radius: 12px;
    z-index: 40;
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 14px;
    color: var(--text);
  }
  @media (max-width: 600px) {
    .toast {
      left: 16px;
      max-width: calc(100vw - 32px);
      bottom: 80px;
    }
  }
</style>
