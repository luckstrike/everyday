<script lang="ts">
  import type { DashboardState, Size } from '../../lib/types.ts';
  import { day, time } from '../../lib/format.ts';
  let {
    data,
    size,
    now,
  }: { data: DashboardState['localEvents']; size: Size; now: number } =
    $props();
  let index = $state(0);
  const upcoming = $derived(
    data.filter((event) => Date.parse(event.startsAt) > now),
  );
  const event = $derived(upcoming[index % (upcoming.length || 1)]);
</script>

<div class="widget-body discover">
  <p class="eyebrow">Around Dallas <span>Sample picks</span></p>
  {#if event}
    <p class="discovery-date">
      {day(event.startsAt, now)}{#if size !== 'small'}
        · {time(event.startsAt)}
      {/if}
    </p>
    <p class="widget-title">{event.title}</p>
    <p class="quiet discovery-venue">{event.venue}</p>
    {#if size !== 'small'}<p class="quiet discovery-price">
        {event.price}
      </p>{/if}
    {#if size === 'large'}<p class="discovery-reason">{event.reason}</p>{/if}
    <div class="discovery-controls">
      <span>{(index % upcoming.length) + 1} / {upcoming.length} · Samples</span
      ><button
        aria-label="Next event recommendation"
        onclick={() => (index = (index + 1) % upcoming.length)}>Next →</button
      >
    </div>
  {:else}<p class="widget-title">More to come</p>
    <p class="quiet">No upcoming sample events.</p>{/if}
</div>
