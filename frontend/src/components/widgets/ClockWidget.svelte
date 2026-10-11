<script lang="ts">
  import type { Size } from '../../lib/types.ts';
  import Icon from '../Icon.svelte';
  import MonthCalendar from '../MonthCalendar.svelte';
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
  let monthPopover: HTMLDivElement;
  let calendarButton: HTMLButtonElement;
  let calendarOpen = $state(false);
  let popupLeft = $state(16),
    popupTop = $state(16);
  function toggleCalendar() {
    monthPopover.togglePopover();
    if (!monthPopover.matches(':popover-open')) return;
    const trigger = calendarButton.getBoundingClientRect();
    const card = calendarButton.closest('.widget')?.getBoundingClientRect();
    const panel = monthPopover.getBoundingClientRect();
    popupLeft = Math.max(
      16,
      Math.min(
        (card?.right ?? trigger.right) - panel.width,
        window.innerWidth - panel.width - 16,
      ),
    );
    const below = (card?.bottom ?? trigger.bottom) + 8;
    popupTop =
      below + panel.height <= window.innerHeight - 16
        ? below
        : Math.max(16, (card?.top ?? trigger.top) - panel.height - 8);
  }
  function closeCalendar() {
    monthPopover.hidePopover();
    calendarButton.focus();
  }
</script>

<svelte:window onresize={() => monthPopover?.hidePopover()} />
<div class="widget-body clock" class:expanded={size === 'large'}>
  <div class="clock-summary">
    <div class="clock-time">
      <span>{clock}</span><small
        >{parts.find((p) => p.type === 'dayPeriod')?.value}</small
      >
    </div>
    <div class="clock-date-block">
      <div class="clock-date">
        <span
          >{new Date(now).toLocaleDateString('en-US', {
            weekday: 'long',
          })}</span
        ><span
          >{new Date(now).toLocaleDateString('en-US', {
            month: 'long',
            day: 'numeric',
          })}</span
        >
      </div>
      <button
        class="icon-button month-trigger"
        bind:this={calendarButton}
        aria-label="Show month calendar"
        aria-expanded={calendarOpen}
        aria-controls="clock-month-popover"
        onclick={toggleCalendar}><Icon name="calendar" size={18} /></button
      >
    </div>
  </div>
  <div class="clock-month"><MonthCalendar {now} /></div>
  <div
    id="clock-month-popover"
    class="month-popover popover"
    popover="auto"
    bind:this={monthPopover}
    style:left={`${popupLeft}px`}
    style:top={`${popupTop}px`}
    ontoggle={() => (calendarOpen = monthPopover.matches(':popover-open'))}
  >
    <MonthCalendar {now} onclose={closeCalendar} />
  </div>
</div>

<style>
  .clock {
    gap: 0;
  }
  .clock-summary {
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 20px;
    width: 100%;
    min-width: 0;
    height: 100%;
  }
  .clock-date-block {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-shrink: 0;
  }
  .clock-time {
    flex-shrink: 0;
  }
  .clock-time > span {
    font-size: min(27cqw, calc(100cqh - 2 * var(--pad) - 56px), 192px);
  }
  .month-trigger {
    color: var(--muted);
    flex-shrink: 0;
    margin-block: -10px;
  }
  .clock-month {
    display: none;
  }
  .month-popover {
    position: fixed;
    inset: auto;
    margin: 0;
    width: 320px;
    max-width: calc(100vw - 32px);
    max-height: calc(100dvh - 32px);
    overflow-y: auto;
    padding: 20px;
  }
  @container (min-aspect-ratio: 3) {
    .clock-summary {
      flex-direction: row;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
    }
    .clock-time > span {
      font-size: min(27cqw, 70cqh, 192px);
    }
  }
  @container (max-width: 559px) and (max-aspect-ratio: 3) {
    .clock-summary {
      gap: 8px;
    }
    .clock-time > span {
      font-size: min(24cqw, 53cqh, 144px);
    }
  }
  @container (min-width: 640px) and (min-height: 320px) {
    .clock {
      display: grid;
      grid-template-columns: minmax(0, 1fr) 224px;
      gap: 24px;
      align-items: center;
    }
    .clock-summary {
      flex-direction: column;
      align-items: flex-start;
      gap: 20px;
    }
    .clock-time > span {
      font-size: min(21cqw, 38cqh, 144px);
    }
    .clock-date {
      flex-direction: row;
      flex-wrap: wrap;
      text-align: left;
      font-size: 18px;
    }
    .clock-date span + span {
      font-size: 18px;
    }
    .clock-month {
      display: block;
      padding-left: 20px;
      border-left: 1px solid var(--line);
    }
    .month-trigger {
      display: none;
    }
  }
  @container (max-height: 150px) {
    .clock-summary {
      gap: 8px;
    }
  }
  @container (max-height: 100px) {
    .clock-summary {
      flex-direction: row;
      align-items: center;
      justify-content: space-between;
      gap: 8px;
    }
    .clock-date {
      display: none;
    }
    .clock-time > span {
      font-size: min(24cqw, 60cqh, 96px);
    }
  }
</style>
