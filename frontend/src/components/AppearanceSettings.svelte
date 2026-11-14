<script lang="ts">
  import Icon from './Icon.svelte';
  import SettingsHeader from './SettingsHeader.svelte';
  import { themes, type Theme } from '../lib/theme.ts';
  let { theme, onchange }: { theme: Theme; onchange: (theme: Theme) => void } =
    $props();
</script>

<div class="appearance">
  <SettingsHeader title="Appearance" description="A color for your space." />
  <div class="theme-options" role="group" aria-label="Dashboard theme">
    {#each themes as { id, label }}<button
        class:active={theme === id}
        aria-pressed={theme === id}
        onclick={() => onchange(id)}
        ><span class="theme-swatch" data-theme={id} aria-hidden="true"
        ></span><span>{label}</span>{#if theme === id}<Icon
            name="check"
            size={18}
          />{/if}</button
      >{/each}
  </div>
</div>

<style>
  .appearance {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  /* Themes sit side by side and share the width, instead of one narrow stack. */
  .theme-options {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 12px;
  }
  .theme-options button {
    justify-content: flex-start;
    min-height: 56px;
    padding: 10px 12px;
    border-radius: 12px;
  }
  .theme-options button.active {
    border-color: var(--accent);
    background: var(--surface);
  }
  .theme-options button > span:nth-child(2) {
    flex: 1;
    text-align: left;
  }
  .theme-options :global(svg) {
    color: var(--accent);
  }
  /* The swatch carries its theme's data-theme, so it reads that theme's colours from
     app.css whichever theme is showing: the fill is its accent, the ring its line. */
  .theme-swatch {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    border: 6px solid var(--line);
    background: var(--accent);
  }
</style>
