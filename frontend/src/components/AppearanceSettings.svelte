<script lang="ts">
  import Icon from './Icon.svelte';
  import type { Theme } from '../lib/theme.ts';
  let { theme, onchange }: { theme: Theme; onchange: (theme: Theme) => void } =
    $props();
</script>

<div class="appearance">
  <div>
    <h3>Appearance</h3>
    <p class="quiet">A color for your space.</p>
  </div>
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
    gap: 16px;
    max-width: 360px;
  }
  h3 {
    margin: 0 0 4px;
    font-size: 20px;
    font-weight: 400;
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
