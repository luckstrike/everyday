<script lang="ts">
  import type { Size } from '../../lib/types.ts';
  let { now, size }: { now: number; size: Size } = $props();
  const date = $derived(new Date(now));
</script>

<div class="clock widget-body" class:compact={size === 'small'}>
  <p class="eyebrow">{date.toLocaleDateString('en-US', { weekday: 'long' })}</p>
  <div class="clock-time">
    {date.toLocaleTimeString('en-US', {
      hour: 'numeric',
      minute: '2-digit',
      hour12: false,
    })}<span class="seconds"
      >{date
        .toLocaleTimeString('en-US', { second: '2-digit' })
        .padStart(2, '0')}</span
    >
  </div>
  <p class="clock-date">
    {date.toLocaleDateString('en-US', {
      month: 'long',
      day: 'numeric',
      year: 'numeric',
    })}
  </p>
  {#if size === 'large'}<p class="quiet">A little space for your day.</p>{/if}
</div>
