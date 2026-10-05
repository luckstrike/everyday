<script lang="ts">
  import type { DashboardState, Size } from '../../lib/types.ts';
  import { day, time } from '../../lib/format.ts';
  let {
    data,
    size,
    now,
  }: { data: DashboardState['calendar']; size: Size; now: number } = $props();
  const events = $derived(
    data.filter((event) => Date.parse(event.endsAt) > now),
  );
</script>

<div class="widget-body calendar">
  <p class="eyebrow">Up next <span>Calendar</span></p>
  {#if events[0]}
    <p class="event-when">
      {day(events[0].startsAt, now)} <span>{time(events[0].startsAt)}</span>
    </p>
    <p class="widget-title">{events[0].title}</p>
    {#if size !== 'small'}<p class="quiet">{events[0].location}</p>{/if}
    {#if size === 'large'}<div class="section-note">
        {#each events.slice(1) as event}<p class="eyebrow">
            {day(event.startsAt, now)} · {time(event.startsAt)}
          </p>
          <p class="widget-title">{event.title}</p>
          <p class="quiet">{event.location}</p>{/each}
      </div>{/if}
  {:else}<p class="widget-title">Room to breathe.</p>
    <p class="quiet">Nothing else on the calendar.</p>{/if}
</div>
