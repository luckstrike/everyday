<script lang="ts">
  import Icon from './Icon.svelte';

  let { low, high }: { low: number; high: number } = $props();
</script>

<span class="temperature-range">
  <span class="sr-only">{'Low '}</span><span class="low">{low}°</span>
  <Icon name="arrow-right" size={16} />
  <span class="sr-only">{' to high '}</span><span class="high">{high}°</span>
</span>

<style>
  .temperature-range {
    display: inline-flex;
    align-items: center;
    gap: 0.25em;
    white-space: nowrap;
    font-variant-numeric: tabular-nums;
  }
  .low {
    color: var(--muted);
  }
  .high {
    color: var(--text);
  }
  .temperature-range :global(svg) {
    width: 0.75em;
    height: 0.75em;
    color: var(--muted);
  }
</style>
