<script lang="ts">
  import type { DashboardState, Size } from '../../lib/types.ts';
  import Icon from '../Icon.svelte';
  let { data, size }: { data: DashboardState['weather']; size: Size } =
    $props();
  let forecast = $state<'hours' | 'days'>('days');
</script>

<div class="widget-body weather" class:has-alert={!!data.alert}>
  <div class="weather-current">
    <span class="weather-icon accent"><Icon name="weather" size={56} /></span
    ><span class="temperature">{data.temperature}°</span>
    <div>
      <p class="widget-title">{data.condition}</p>
      <p class="quiet">High {data.high}° · Low {data.low}°</p>
    </div>
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

<style>
  .widget-body.weather {
    --weather-inset: 32px;
    position: relative;
    padding: var(--weather-inset);
    justify-content: center;
    gap: 28px;
  }
  .weather-current {
    display: flex;
    align-items: center;
    gap: 20px;
    min-width: 0;
    padding-right: 72px;
  }
  .weather-icon {
    display: flex;
    flex-shrink: 0;
    margin-left: -4px;
  }
  .weather-icon :global(svg) {
    display: block;
  }
  .temperature {
    font-size: 72px;
    font-weight: 300;
    line-height: 1;
    letter-spacing: -0.05em;
    flex-shrink: 0;
  }
  .weather-current > div {
    min-width: 0;
  }
  .weather-current .widget-title {
    font-size: 24px;
    line-height: 1.25;
  }
  .weather-current .quiet {
    margin-top: 8px;
    font-size: 16px;
    line-height: 1.4;
  }
  .forecast-strip {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 12px;
  }
  .forecast-strip > div {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .forecast-strip span,
  .forecast-strip small {
    font-size: 16px;
    color: var(--muted);
    font-weight: 400;
    line-height: 1.2;
  }
  .forecast-strip strong {
    font-size: 24px;
    font-weight: 400;
    line-height: 1.2;
    white-space: nowrap;
  }
  .forecast-strip strong small {
    font-size: 18px;
    margin-left: 4px;
  }
  .forecast-toggle {
    position: absolute;
    top: 12px;
    right: 16px;
    font-size: 12px;
    color: var(--muted);
    border: 0;
    padding: 4px 8px;
    min-width: 64px;
  }
  :global(.editing) .forecast-toggle {
    right: 48px;
  }
  .alert-copy {
    color: var(--accent);
    font-size: clamp(16px, 3cqw, 32px);
    max-width: 60ch;
    line-height: 1.4;
  }
  @container (max-height:230px) {
    .widget-body.weather {
      --weather-inset: 24px;
      gap: 16px;
    }
    .weather-current {
      gap: 16px;
    }
    .weather-icon {
      margin-left: -3.5px;
    }
    .weather-icon :global(svg) {
      width: 44px;
      height: 44px;
    }
    .temperature {
      font-size: 56px;
    }
    .weather-current .widget-title {
      font-size: 20px;
    }
    .weather-current .quiet {
      margin-top: 4px;
      font-size: 14px;
    }
    .forecast-strip > div {
      gap: 6px;
    }
    .forecast-strip span {
      font-size: 14px;
    }
    .forecast-strip strong {
      font-size: 22px;
    }
    .forecast-strip strong small {
      font-size: 16px;
    }
    .has-alert .alert-copy {
      font-size: 14px;
      display: -webkit-box;
      -webkit-line-clamp: 1;
      line-clamp: 1;
      -webkit-box-orient: vertical;
      overflow: hidden;
    }
    .has-alert.widget-body.weather {
      gap: 8px;
    }
  }
  @container (max-width:480px) {
    .widget-body.weather {
      --weather-inset: 20px;
      gap: 16px;
    }
    .weather-icon {
      display: none;
    }
    .weather-current {
      gap: 12px;
    }
    .temperature {
      font-size: 48px;
    }
    .weather-current .widget-title {
      font-size: 18px;
    }
    .weather-current .quiet {
      font-size: 12px;
      margin-top: 4px;
    }
    .forecast-strip {
      gap: 8px;
    }
    .forecast-strip span {
      font-size: 13px;
    }
    .forecast-strip strong {
      font-size: 20px;
    }
    .forecast-strip strong small {
      font-size: 14px;
      margin-left: 2px;
    }
    .forecast-toggle {
      top: 8px;
      right: 12px;
    }
  }
  @container (max-height:180px) {
    .widget-body.weather {
      gap: 12px;
    }
  }
  @container (max-width:340px) {
    .widget-body.weather {
      --weather-inset: 16px;
      gap: 12px;
    }
    .weather-current {
      gap: 8px;
    }
    .temperature {
      font-size: 40px;
    }
    .weather-current .widget-title {
      font-size: 16px;
    }
    .weather-current .quiet {
      font-size: 11px;
    }
    .forecast-strip {
      gap: 4px;
    }
    .forecast-strip span {
      font-size: 12px;
    }
    .forecast-strip strong {
      font-size: 18px;
    }
    .forecast-strip strong small {
      font-size: 12px;
    }
  }
  @container (max-height:120px) {
    .widget-body.weather {
      --weather-inset: 12px;
      gap: 8px;
    }
    .temperature {
      font-size: 32px;
    }
    .weather-current .quiet {
      display: none;
    }
    .weather-current .widget-title {
      font-size: 14px;
    }
    .forecast-strip > div {
      flex-direction: row;
      align-items: baseline;
      gap: 4px;
    }
    .forecast-strip span {
      font-size: 11px;
    }
    .forecast-strip strong {
      font-size: 14px;
    }
    .forecast-strip strong small {
      display: none;
    }
    .forecast-toggle {
      top: 2px;
      right: 8px;
    }
  }
</style>
