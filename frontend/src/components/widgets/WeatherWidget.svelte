<script lang="ts">
  import type { DashboardState, Size } from '../../lib/types.ts';
  let { data, size }: { data: DashboardState['weather']; size: Size } =
    $props();
  let forecast = $state<'hours' | 'days'>('days');
</script>

<div class="widget-body weather">
  <p class="eyebrow">Weather <span>{data.location}</span></p>
  <div class="weather-current">
    <span class="temperature">{data.temperature}°</span>
    <div>
      <p class="widget-title">{data.condition}</p>
      {#if size !== 'small'}<p class="quiet">
          High {data.high}° · Low {data.low}° · {data.precipitation}% rain
        </p>{/if}
    </div>
  </div>
  {#if data.alert}<p class="alert-copy">{data.alert}</p>{/if}
  <div class="forecast-switch" aria-label="Weather forecast">
    <button
      aria-pressed={forecast === 'hours'}
      onclick={() => (forecast = 'hours')}>Hours</button
    ><button
      aria-pressed={forecast === 'days'}
      onclick={() => (forecast = 'days')}>Days</button
    >
  </div>
  {#if forecast === 'hours'}
    <div class="forecast-strip">
      {#each data.hourly.slice(0, size === 'small' ? 3 : 4) as hour}<div>
          <span>{hour.label}</span><strong>{hour.temperature}°</strong
          >{#if size === 'large'}<small>{hour.condition}</small>{/if}
        </div>{/each}
    </div>
  {:else}
    <div class="forecast-strip">
      {#each (data.daily ?? []).slice(0, size === 'small' ? 3 : 4) as day}<div>
          <span
            >{new Date(day.date).toLocaleDateString('en-US', {
              weekday: 'short',
            })}</span
          ><strong>{day.high}° <small>{day.low}°</small></strong
          >{#if size === 'large'}<small
              >{day.condition} · {day.precipitation}%</small
            >{/if}
        </div>{/each}
    </div>
  {/if}
</div>
