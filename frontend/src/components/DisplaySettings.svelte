<script lang="ts">
  import { tick } from 'svelte';
  import Icon from './Icon.svelte';
  import type { Theme } from '../lib/theme.ts';
  let { theme, onchange }: { theme: Theme; onchange: (theme: Theme) => void } =
    $props();
  let open = $state(false);
  let trigger: HTMLButtonElement;
  let panel = $state<HTMLElement>();
  function close() {
    open = false;
    trigger?.focus();
  }
  async function toggle() {
    open = !open;
    if (open) {
      await tick();
      panel?.focus();
    }
  }
</script>

<svelte:window
  onclick={(e) => {
    if (
      !e
        .composedPath()
        .some(
          (node) =>
            node instanceof Element &&
            node.classList.contains('display-settings'),
        )
    )
      open = false;
  }}
  onkeydown={(e) => {
    if (open && e.key === 'Escape') {
      e.stopImmediatePropagation();
      close();
    }
  }}
/>
<div class="display-settings">
  <button
    bind:this={trigger}
    class="icon-button settings-trigger"
    aria-label="Display settings"
    aria-expanded={open}
    aria-controls="display-settings-panel"
    onclick={toggle}><Icon name="settings" size={20} /></button
  >
  {#if open}<section
      id="display-settings-panel"
      bind:this={panel}
      class="settings-panel popover"
      aria-label="Display settings"
      tabindex="-1"
    >
      <div class="popover-heading">
        <h2>Appearance</h2>
        <button class="icon-button" aria-label="Close settings" onclick={close}
          ><Icon name="close" size={20} /></button
        >
      </div>
      <p class="quiet">A color for your space.</p>
      <div class="theme-options" aria-label="Dashboard theme">
        {#each ['evergreen', 'turquoise'] as value}<button
            class:active={theme === value}
            aria-pressed={theme === value}
            onclick={() => onchange(value as Theme)}
            ><span
              class="theme-swatch"
              class:turquoise={value === 'turquoise'}
              aria-hidden="true"
            ></span><span
              >{value === 'evergreen' ? 'Evergreen' : 'Turquoise'}</span
            >{#if theme === value}<Icon name="check" size={18} />{/if}</button
          >{/each}
      </div>
    </section>{/if}
</div>

<style>
  .display-settings {
    position: relative;
    z-index: 20;
  }
  .settings-trigger {
    color: var(--muted);
  }
  .settings-trigger[aria-expanded='true'] {
    color: var(--accent);
    background: var(--surface);
  }
  .settings-panel {
    position: absolute;
    top: 52px;
    right: 0;
    width: 300px;
    max-width: calc(100vw - 32px);
    max-height: calc(100dvh - 120px);
    overflow-y: auto;
    padding: 20px;
    display: grid;
    gap: 16px;
  }
  .theme-options {
    display: grid;
    gap: 8px;
  }
  .theme-options button {
    justify-content: flex-start;
    min-height: 56px;
    padding: 10px 12px;
    border-radius: 12px;
  }
  .theme-options button.active {
    border-color: var(--accent);
    background: var(--bg);
  }
  .theme-options button > span:nth-child(2) {
    flex: 1;
    text-align: left;
  }
  .theme-options :global(svg) {
    color: var(--accent);
  }
  .theme-swatch {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    border: 6px solid oklch(39.45% 0.0284 156.49);
    background: oklch(82.38% 0.0619 130.22);
  }
  .theme-swatch.turquoise {
    border-color: oklch(42% 0.035 190);
    background: oklch(83% 0.084 185);
  }
</style>
