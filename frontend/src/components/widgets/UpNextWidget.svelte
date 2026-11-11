<script lang="ts">
  import type { DashboardState, Size, SourceStatus } from '../../lib/types.ts';
  import { when } from '../../lib/format.ts';
  import { sourceNotice } from '../../lib/sources.ts';
  import { fittingRowCount } from '../../lib/fit.ts';
  import WidgetFrame from '../WidgetFrame.svelte';
  let {
    data,
    status,
    size,
    now,
  }: {
    data: DashboardState['calendar'];
    status: SourceStatus;
    size: Size;
    now: number;
  } = $props();
  const notice = $derived(sourceNotice(status, 'Calendar'));
  const events = $derived(
    data
      .filter((e) => Date.parse(e.endsAt) > now)
      .sort((a, b) => Date.parse(a.startsAt) - Date.parse(b.startsAt)),
  );
  let eventList = $state<HTMLOListElement>();
  let visibleCount = $state(0);
  const message = $derived(
    notice ??
      (events.length === 0
        ? { title: 'A little breathing room', detail: 'No upcoming events.' }
        : null),
  );
  // When it happens, plus where once there is room for the longer line.
  const detail = (event: DashboardState['calendar'][number]) =>
    size === 'large' && event.location
      ? `${when(event, now)} · ${event.location}`
      : when(event, now);

  $effect(() => {
    const list = eventList;
    if (!list || events.length === 0) return;
    const rows = Array.from(list.children);
    const measure = () => {
      const gap = Number.parseFloat(getComputedStyle(list).rowGap) || 0;
      visibleCount = fittingRowCount(
        list.getBoundingClientRect().height,
        rows.map((row) => row.getBoundingClientRect().height),
        gap,
      );
    };
    const observer = new ResizeObserver(measure);
    observer.observe(list);
    for (const row of rows) observer.observe(row);
    measure();

    return () => observer.disconnect();
  });
</script>

<WidgetFrame title="Up next" {message}>
  <ol class="events" class:large={size === 'large'} bind:this={eventList}>
    {#each events as event, i}<li
        class="event"
        style:visibility={i < visibleCount ? 'visible' : 'hidden'}
        aria-hidden={i >= visibleCount}
      >
        <p class="event-title" title={event.title}>{event.title}</p>
        <p class="event-when quiet" class:accent={i === 0}>
          {detail(event)}
        </p>
      </li>{/each}
  </ol>
</WidgetFrame>

<style>
  .events {
    list-style: none;
    margin: 0;
    padding: 0;
    flex: 1 1 0;
    min-height: 0;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  /* Each event is a card: the widget's outline, one radius step smaller, with
     every other card filled with the theme's raised surface colour. */
  .event {
    flex-shrink: 0;
    min-width: 0;
    padding: 12px;
    border: 1px solid var(--line);
    border-radius: 12px;
  }
  .event:nth-child(odd) {
    background: var(--surface);
  }
  .event-title {
    font-size: 18px;
    line-height: 1.25;
    font-weight: 400;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .event .event-when {
    font-size: 14px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .events.large {
    gap: 10px;
  }
  .large .event {
    padding: 16px 14px;
  }
  .large .event-title {
    font-size: 22px;
  }
  .large .event .event-when {
    font-size: 16px;
  }
</style>
