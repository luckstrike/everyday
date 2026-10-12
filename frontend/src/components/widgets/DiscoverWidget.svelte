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
      <p class="widget-title" title={event.title}>{event.title}</p>
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
        class="discovery-venue">{event.venue}</span
      ><span class="accent">{event.price}</span>
    </div>
    {#if size === 'large'}<p class="quiet discovery-reason">
        {event.reason}
      </p>{/if}{:else}<p class="widget-title">More to come</p>
    <p class="quiet">No upcoming sample events.</p>{/if}
</div>

<style>
  .discovery-title .widget-title {
    min-width: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
  @container (max-height: 150px) and (max-width: 340px) {
    .discovery-venue {
      display: none;
    }
  }
  @container (max-height: 120px) {
    .discover .eyebrow {
      display: none;
    }
    .discovery-venue {
      display: none;
    }
  }
  @container (max-width: 210px) and (max-height: 150px) {
    .discover .eyebrow,
    .discovery-details .accent {
      display: none;
    }
    .discovery-title .widget-title {
      font-size: 18px;
      white-space: normal;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      line-clamp: 2;
      -webkit-box-orient: vertical;
    }
  }
</style>
