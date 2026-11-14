<script lang="ts">
  import { monthGrid, weekContaining } from '../lib/month.ts';
  let { now, week = false }: { now: number; week?: boolean } = $props();
  const year = $derived(new Date(now).getFullYear());
  const month = $derived(new Date(now).getMonth());
  const today = $derived(new Date(now).getDate());
  const calendar = $derived(monthGrid(year, month));
  const rows = $derived(
    week ? weekContaining(calendar.weeks, today) : calendar.weeks,
  );
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

<div class="month-calendar" style:--calendar-rows={rows.length}>
  <table aria-label={calendar.label}>
    <thead
      ><tr
        >{#each weekdays as weekday}<th scope="col"
            ><abbr title={weekday}>{weekday.slice(0, 2)}</abbr></th
          >{/each}</tr
      ></thead
    >
    <tbody
      >{#each rows as row}<tr
          >{#each row as day}<td>
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
  .month-calendar {
    --calendar-day-size: min(
      72px,
      calc(100cqw / 7 - 8px),
      calc((100cqh - 32px) / var(--calendar-rows) - 4px)
    );
    width: 100%;
    height: 100%;
  }
  table {
    width: 100%;
    height: 100%;
    border-collapse: separate;
    border-spacing: 0 4px;
    table-layout: fixed;
    text-align: center;
    font-variant-numeric: tabular-nums;
  }
  th {
    height: 24px;
    padding: 0;
    font-size: clamp(12px, calc(var(--calendar-day-size) * 0.3), 18px);
    line-height: 1.2;
    font-weight: 400;
    color: var(--muted);
  }
  abbr {
    text-decoration: none;
  }
  td {
    padding: 0;
    font-size: clamp(14px, calc(var(--calendar-day-size) * 0.45), 32px);
    line-height: 1;
  }
  time {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: var(--calendar-day-size);
    height: var(--calendar-day-size);
    border-radius: 50%;
  }
  .today {
    color: var(--bg);
    background: var(--accent);
    font-weight: 500;
  }
</style>
