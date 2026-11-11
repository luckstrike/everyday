<script lang="ts">
  import type { DashboardState, Size, SourceStatus } from '../../lib/types.ts';
  import { when } from '../../lib/format.ts';
  import { sourceNotice } from '../../lib/sources.ts';
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
  // Up to three events are rendered; the stylesheet hides the ones that don't fit.
  const shown = $derived(events.slice(0, 3));
  const message = $derived(
    notice ??
      (shown.length === 0
        ? { title: 'A little breathing room', detail: 'No upcoming events.' }
        : null),
  );
  // When it happens, plus where once there is room for the longer line.
  const detail = (event: DashboardState['calendar'][number]) =>
    size === 'large' && event.location
      ? `${when(event, now)} · ${event.location}`
      : when(event, now);
</script>

<WidgetFrame title="Up next" {message}>
  <ol class="events" class:large={size === 'large'}>
    {#each shown as event, i}<li class="event">
        <p class="event-title" title={event.title}>{event.title}</p>
        <p class="event-when quiet" class:accent={i === 0}>
          {detail(event)}
        </p>
      </li>{/each}
  </ol>
</WidgetFrame>

<style>
  /* The list is its own container, so how many cards fit depends only on the
     height it is given, not on the widget's padding or text sizes. */
  .events {
    list-style: none;
    margin: 0;
    padding: 0;
    flex: 1 1 0;
    min-height: 0;
    overflow: hidden;
    container: events / size;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  /* Each event is a card: the widget's outline, one radius step smaller, with
     every other card filled with the theme's raised surface colour. */
  .event {
    min-width: 0;
    padding: 8px 12px;
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
  /* Cards have a fixed height (about 60px, or 76px when large) so the thresholds
     below can show only cards that fit whole, counting the gaps between them. */
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
    padding: 12px 14px;
  }
  .large .event-title {
    font-size: 22px;
  }
  .large .event .event-when {
    font-size: 16px;
  }
  /* Three cards need about 196px, two about 128px. */
  @container events (max-height: 197px) {
    .event:nth-child(n + 3) {
      display: none;
    }
  }
  @container events (max-height: 129px) {
    .event:nth-child(n + 2) {
      display: none;
    }
  }
  /* Large cards: three need about 248px, two about 162px. */
  @container events (max-height: 249px) {
    .large .event:nth-child(n + 3) {
      display: none;
    }
  }
  @container events (max-height: 163px) {
    .large .event:nth-child(n + 2) {
      display: none;
    }
  }
</style>
