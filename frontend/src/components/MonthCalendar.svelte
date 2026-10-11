<script lang="ts">
  import Icon from './Icon.svelte';
  import { monthGrid } from '../lib/month.ts';
  let { now, onclose }: { now: number; onclose?: () => void } = $props();
  const year = $derived(new Date(now).getFullYear());
  const month = $derived(new Date(now).getMonth());
  const today = $derived(new Date(now).getDate());
  const calendar = $derived(monthGrid(year, month));
  const weekdays = [
    'Sunday',
    'Monday',
    'Tuesday',
    'Wednesday',
    'Thursday',
    'Friday',
    'Saturday',
  ];
  const dateString = (day: number) =>
    `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
</script>

<div class="month-calendar">
  <div class="month-heading">
    <h2>{calendar.label}</h2>
    {#if onclose}<button
        class="icon-button"
        aria-label="Close month calendar"
        onclick={onclose}><Icon name="close" size={18} /></button
      >{/if}
  </div>
  <table aria-label={calendar.label}>
    <thead
      ><tr
        >{#each weekdays as weekday}<th scope="col"
            ><abbr title={weekday}>{weekday.slice(0, 2)}</abbr></th
          >{/each}</tr
      ></thead
    >
    <tbody
      >{#each calendar.weeks as week}<tr
          >{#each week as day}<td>
              {#if day !== null}<time
                  datetime={dateString(day)}
                  aria-current={day === today ? 'date' : undefined}
                  class:today={day === today}>{day}</time
                >{/if}
            </td>{/each}</tr
        >{/each}</tbody
    >
  </table>
</div>

<style>
  .month-heading {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-bottom: 12px;
  }
  h2 {
    font-size: 18px;
  }
  .month-heading button {
    margin: -8px -8px -8px 0;
    color: var(--muted);
  }
  table {
    width: 100%;
    border-collapse: separate;
    border-spacing: 0 4px;
    table-layout: fixed;
    text-align: center;
    font-variant-numeric: tabular-nums;
  }
  th {
    padding: 0 0 6px;
    font-size: 12px;
    font-weight: 400;
    color: var(--muted);
  }
  abbr {
    text-decoration: none;
  }
  td {
    height: 28px;
    padding: 0;
    font-size: 14px;
  }
  time {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 50%;
  }
  .today {
    color: var(--bg);
    background: var(--accent);
    font-weight: 500;
  }
</style>
