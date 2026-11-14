<script lang="ts">
  import Icon from './Icon.svelte';
  import SettingsHeader from './SettingsHeader.svelte';
  import type { Theme } from '../lib/theme.ts';
  let { theme, onchange }: { theme: Theme; onchange: (theme: Theme) => void } =
    $props();
</script>

<div class="appearance">
  <SettingsHeader title="Appearance" description="A color for your space." />
  <div class="theme-options" role="group" aria-label="Dashboard theme">
    {#each ['evergreen', 'turquoise'] as value}<button
        class:active={theme === value}
        aria-pressed={theme === value}
        onclick={() => onchange(value as Theme)}
        ><span
          class="theme-swatch"
          class:turquoise={value === 'turquoise'}
          aria-hidden="true"
        ></span><span>{value === 'evergreen' ? 'Evergreen' : 'Turquoise'}</span
        >{#if theme === value}<Icon name="check" size={18} />{/if}</button
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
