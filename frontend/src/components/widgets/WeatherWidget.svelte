<script lang="ts">
  import type { DashboardState, Size } from '../../lib/types.ts';
  import Icon from '../Icon.svelte';
  let { data, size }: { data: DashboardState['weather']; size: Size } =
    $props();
  let forecast = $state<'hours' | 'days'>('days');
</script>

<div class="widget-body weather">
  <div class="weather-current">
    <span class="weather-icon accent"><Icon name="weather" size={56} /></span
    ><span class="temperature">{data.temperature}°</span>
    <div>
      <p class="widget-title">{data.condition}</p>
      <p class="quiet">High {data.high}° · Low {data.low}°</p>
    </div>
    <button
      class="forecast-toggle"
      aria-label={`Show ${forecast === 'days' ? 'hourly' : 'daily'} forecast`}
      onclick={() => (forecast = forecast === 'days' ? 'hours' : 'days')}
      >{forecast === 'days' ? 'Days' : 'Hours'}<Icon
        name="down"
        size={12}
      /></button
    >
  </div>
  {#if data.alert}<p class="alert-copy">{data.alert}</p>{/if}
  <div class="forecast-strip">
    {#if forecast === 'hours'}{#each data.hourly.slice(0, 4) as hour}<div>
          <span>{hour.label}</span><strong>{hour.temperature}°</strong
          >{#if size === 'large' && data.alert}<small>{hour.condition}</small
            >{/if}
        </div>{/each}{:else}{#each data.daily.slice(0, 4) as day, i}<div>
          <span
            >{i === 0
              ? 'Today'
              : new Date(day.date).toLocaleDateString('en-US', {
                  weekday: 'short',
                })}</span
          ><strong>{day.high}° <small>{day.low}°</small></strong
          >{#if size === 'large' && data.alert}<small>{day.condition}</small
            >{/if}
        </div>{/each}{/if}
  </div>
</div>
