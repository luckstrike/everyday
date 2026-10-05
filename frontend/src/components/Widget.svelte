<script lang="ts">
  import type { DashboardState, Size, WidgetId } from '../lib/types.ts';
  import DiscoverWidget from './widgets/DiscoverWidget.svelte';
  import ClockWidget from './widgets/ClockWidget.svelte';
  import WeatherWidget from './widgets/WeatherWidget.svelte';
  import SportsWidget from './widgets/SportsWidget.svelte';
  import FinanceWidget from './widgets/FinanceWidget.svelte';
  import CalendarWidget from './widgets/CalendarWidget.svelte';
  let {
    id,
    dashboard,
    size,
    now,
  }: { id: WidgetId; dashboard: DashboardState; size: Size; now: number } =
    $props();
</script>

{#if id === 'clock'}<ClockWidget {now} {size} />
{:else if id === 'weather'}<WeatherWidget data={dashboard.weather} {size} />
{:else if id === 'sports'}<SportsWidget data={dashboard.sports} {size} {now} />
{:else if id === 'finance'}<FinanceWidget
    data={dashboard.finance}
    {size}
    {now}
  />
{:else if id === 'discover'}<DiscoverWidget
    data={dashboard.localEvents ?? []}
    {size}
    {now}
  />{:else}<CalendarWidget data={dashboard.calendar} {size} {now} />{/if}
