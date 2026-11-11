<script lang="ts">
  import type { DashboardState, Size } from '../../lib/types.ts';
  import { day, time } from '../../lib/format.ts';
  let {
    data,
    size,
    now,
  }: { data: DashboardState['calendar']; size: Size; now: number } = $props();
  const events = $derived(
    data
      .filter((e) => Date.parse(e.endsAt) > now)
      .sort((a, b) => Date.parse(a.startsAt) - Date.parse(b.startsAt)),
  );
</script>

<div class="widget-body calendar">
  <p class="eyebrow">Up next</p>
  {#if events[0]}<p class="widget-title" title={events[0].title}>
      {events[0].title}
    </p>
    <p class="quiet accent">
      {day(events[0].startsAt, now)} · {time(events[0].startsAt)}
    </p>
    {#if size === 'large'}<p class="quiet">{events[0].location}</p>
      {#each events.slice(1, 3) as event}<div class="extra-event">
          <p>{event.title}</p>
          <small>{day(event.startsAt, now)} · {time(event.startsAt)}</small>
        </div>{/each}{/if}{:else}<p class="widget-title">
      A little breathing room
    </p>
    <p class="quiet">No upcoming events.</p>{/if}
</div>

<style>
  .widget-title {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  @container (max-height: 120px) {
    .calendar .eyebrow {
      display: none;
    }
    .widget-title {
      -webkit-line-clamp: 1;
      line-clamp: 1;
    }
  }
  @container (max-width: 210px) and (max-height: 150px) {
    .calendar .eyebrow {
      display: none;
    }
  }
</style>
