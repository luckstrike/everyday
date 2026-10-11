<script lang="ts">
  import type { Size } from '../../lib/types.ts';
  let { now, size }: { now: number; size: Size } = $props();
  const parts = $derived(
    new Intl.DateTimeFormat('en-US', {
      hour: 'numeric',
      minute: '2-digit',
      hour12: true,
    }).formatToParts(now),
  );
  const clock = $derived(
    parts
      .filter(
        (p) =>
          (p.type !== 'dayPeriod' && p.type !== 'literal') || p.value === ':',
      )
      .map((p) => p.value)
      .join(''),
  );
</script>

<div class="widget-body clock" class:expanded={size === 'large'}>
  <div class="clock-time">
    <span>{clock}</span><small
      >{parts.find((p) => p.type === 'dayPeriod')?.value}</small
    >
  </div>
  <div class="clock-date">
    <span>{new Date(now).toLocaleDateString('en-US', { weekday: 'long' })}</span
    ><span
      >{new Date(now).toLocaleDateString('en-US', {
        month: 'long',
        day: 'numeric',
      })}</span
    >
  </div>
</div>
