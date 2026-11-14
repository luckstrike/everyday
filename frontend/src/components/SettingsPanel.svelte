<script lang="ts">
  import { onMount } from 'svelte';
  import AppearanceSettings from './AppearanceSettings.svelte';
  import BillsEditor from './BillsEditor.svelte';
  import { idleTimer } from '../lib/idle.ts';
  import { sections, type Section } from '../lib/route.ts';
  import type { Theme } from '../lib/theme.ts';

  let {
    section,
    theme,
    onsection,
    onclose,
    onchoosetheme,
    onbillschanged,
  }: {
    section: Section;
    theme: Theme;
    onsection: (section: Section) => void;
    onclose: () => void;
    onchoosetheme: (theme: Theme) => void;
    onbillschanged: () => void;
  } = $props();

  // This is a wall display. Settings opens on purpose, but nobody should have to
  // remember to close it, so it hands the screen back after a quiet spell.
  const IDLE_MS = 90_000;
  const idle = idleTimer(IDLE_MS, () => onclose());
  let panel = $state<HTMLElement>();
  onMount(() => {
    idle.touch();
    panel?.focus();
    return () => idle.stop();
  });
</script>

<svelte:window
  onkeydown={(event) => {
    if (event.key === 'Escape') onclose();
  }}
/>

<div
  class="settings-panel"
  role="dialog"
  aria-modal="true"
  aria-label="Settings"
  tabindex="-1"
  bind:this={panel}
  onpointerdown={idle.touch}
  onkeydown={idle.touch}
  oninput={idle.touch}
  onwheel={idle.touch}
>
  <header class="panel-header">
    <h2>Settings</h2>
    <button class="primary" onclick={onclose}>Done</button>
  </header>
  <div class="panel-body">
    <nav class="rail" aria-label="Settings sections">
      {#each sections as item}<button
          class:active={section === item.id}
          aria-current={section === item.id ? 'page' : undefined}
          onclick={() => onsection(item.id)}>{item.label}</button
        >{/each}
    </nav>
    <section class="panel-content">
      {#if section === 'bills'}
        <BillsEditor onchanged={onbillschanged} />
      {:else}
        <AppearanceSettings {theme} onchange={onchoosetheme} />
      {/if}
    </section>
  </div>
</div>

<style>
  .settings-panel {
    position: absolute;
    inset: 0;
    z-index: 30;
    display: flex;
    flex-direction: column;
    gap: 20px;
    padding: 28px 32px;
    background: var(--bg);
    outline: none;
  }
  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }
  .panel-body {
    flex: 1 1 0;
    min-height: 0;
    display: flex;
    gap: 28px;
  }
  .rail {
    flex: 0 0 180px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .rail button {
    justify-content: flex-start;
    min-height: 52px;
    border-radius: 14px;
  }
  .rail button.active {
    border-color: var(--accent);
    background: var(--surface);
  }
  /* The content scrolls on its own, so the header and rail stay put and a form
     that outgrows the screen (say, with a touch keyboard up) is still reachable. */
  .panel-content {
    flex: 1 1 0;
    min-width: 0;
    min-height: 0;
    overflow-y: auto;
    padding: 0 4px 24px 0;
  }
  @media (max-width: 640px) {
    .settings-panel {
      padding: 16px;
      gap: 12px;
    }
    .panel-body {
      flex-direction: column;
      gap: 16px;
    }
    .rail {
      flex: none;
      flex-direction: row;
    }
    .rail button {
      flex: 1;
      justify-content: center;
    }
  }
</style>
