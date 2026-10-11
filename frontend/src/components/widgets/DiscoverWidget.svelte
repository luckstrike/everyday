<script lang="ts">
  import type { DashboardState, Size } from '../../lib/types.ts';
  import { day, time } from '../../lib/format.ts';
  import Icon from '../Icon.svelte';
  let {
    data,
    size,
    now,
  }: { data: DashboardState['localEvents']; size: Size; now: number } =
    $props();
  let index = $state(0);
  const upcoming = $derived(data.filter((e) => Date.parse(e.startsAt) > now));
  const event = $derived(upcoming[index % (upcoming.length || 1)]);
</script>

<div class="widget-body discover">
  <p class="eyebrow accent">Around Dallas</p>
  {#if event}<div class="discovery-title">
      <p class="widget-title">{event.title}</p>
      <button
        class="icon-button"
        aria-label="Next event recommendation"
        disabled={upcoming.length < 2}
        onclick={() => (index = (index + 1) % upcoming.length)}
        ><Icon name="arrow" size={20} /></button
      >
    </div>
    <div class="discovery-details quiet">
      <span>{day(event.startsAt, now)} · {time(event.startsAt)}</span><span
        >{event.venue}</span
      ><span class="accent">{event.price}</span>
    </div>
    {#if size === 'large'}<p class="quiet discovery-reason">
        {event.reason}
      </p>{/if}{:else}<p class="widget-title">More to come</p>
    <p class="quiet">No upcoming sample events.</p>{/if}
</div>
