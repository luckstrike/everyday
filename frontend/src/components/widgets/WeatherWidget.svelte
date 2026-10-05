<script lang="ts">
  import type { DashboardState, Size } from '../../lib/types.ts';
  let { data, size }: { data: DashboardState['weather']; size: Size } =
    $props();
</script>

<div class="widget-body weather">
  <p class="eyebrow">Weather <span>{data.location}</span></p>
  <div class="weather-main">
    <span class="temperature">{data.temperature}°</span><svg
      class="weather-icon"
      viewBox="0 0 80 70"
      fill="none"
      aria-hidden="true"
      ><circle
        cx="48"
        cy="25"
        r="14"
        stroke="currentColor"
        stroke-width="2"
      /><path
        d="M48 3v5M70 25h6M65 8l-4 5M28 7l4 5M22 25h5"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
      /><path
        d="M18 59a12 12 0 010-24 17 17 0 0132-3 14 14 0 016 27H18Z"
        fill="var(--surface)"
        stroke="currentColor"
        stroke-width="2"
      /></svg
    >
  </div>
  <p class="widget-title">{data.condition}</p>
  {#if data.alert}<p class="alert-copy">{data.alert}</p>{/if}
  {#if size !== 'small'}<p class="quiet weather-range">
      High {data.high}° <span>Low {data.low}°</span>
      <span>{data.precipitation}% rain</span>
    </p>{/if}
  {#if size === 'large'}
    <div class="forecast">
      {#each data.hourly as hour}<div>
          <p class="quiet">{hour.label}</p>
          <strong>{hour.temperature}°</strong><small>{hour.condition}</small>
        </div>{/each}
    </div>
    <p class="tomorrow"><span class="eyebrow">Tomorrow</span>{data.tomorrow}</p>
  {/if}
</div>
