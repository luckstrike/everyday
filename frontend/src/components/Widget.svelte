<script lang="ts">
  import type { DashboardState, Size, WidgetId } from '../lib/types.ts';
  import ClockWidget from './widgets/ClockWidget.svelte';
  import CalendarWidget from './widgets/CalendarWidget.svelte';
  import WeatherWidget from './widgets/WeatherWidget.svelte';
  import UpNextWidget from './widgets/UpNextWidget.svelte';
  import FinanceWidget from './widgets/FinanceWidget.svelte';
  let {
    id,
    dashboard,
    size,
    now,
  }: { id: WidgetId; dashboard: DashboardState; size: Size; now: number } =
    $props();
</script>

{#if id === 'clock'}<ClockWidget {now} />
{:else if id === 'calendar'}<CalendarWidget {now} />
{:else if id === 'weather'}<WeatherWidget data={dashboard.weather} {size} />
{:else if id === 'upnext'}<UpNextWidget
    data={dashboard.calendar}
    {size}
    {now}
  />
{:else}<FinanceWidget data={dashboard.finance} {size} {now} />{/if}
