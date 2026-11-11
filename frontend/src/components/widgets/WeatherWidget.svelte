<script lang="ts">
  import type { DashboardState, Size, SourceStatus } from '../../lib/types.ts';
  import Icon from '../Icon.svelte';
  import WidgetFrame from '../WidgetFrame.svelte';
  import { sourceNotice } from '../../lib/sources.ts';
  let {
    data,
    status,
    size,
  }: {
    data: DashboardState['weather'];
    status: SourceStatus;
    size: Size;
  } = $props();
  let forecast = $state<'hours' | 'days'>('days');
  const notice = $derived(sourceNotice(status, 'Weather'));
</script>

<WidgetFrame title="Weather">
  {#snippet actions()}
    {#if !notice}<div class="forecast-control" class:has-alert={!!data.alert}>
        <button
          class="forecast-toggle"
          aria-label={`Show ${forecast === 'days' ? 'hourly' : 'daily'} forecast`}
          onclick={() => (forecast = forecast === 'days' ? 'hours' : 'days')}
          >{forecast === 'days' ? 'Days' : 'Hours'}<Icon
            name="down"
            size={12}
          /></button
        >
      </div>{/if}
  {/snippet}
  <div class="weather-content" class:has-alert={!!data.alert}>
    {#if notice}
      <p class="widget-title">{notice.title}</p>
      <p class="quiet">{notice.detail}</p>
    {:else}
      <div class="weather-current">
        <span class="weather-icon accent"
          ><Icon name="weather" size={56} /></span
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
              >{#if size === 'large' && data.alert}<small
                  >{hour.condition}</small
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
    {/if}
  </div>
</WidgetFrame>

<style>
  .weather-content {
    display: flex;
    flex-direction: column;
    justify-content: center;
    flex: 1 1 0;
    min-width: 0;
    min-height: 0;
    gap: 28px;
  }
  .weather-current {
    display: flex;
    align-items: center;
    gap: 20px;
    min-width: 0;
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
    font-size: 12px;
    color: var(--muted);
    border: 0;
    padding: 4px 8px;
    min-width: 64px;
  }
  .alert-copy {
    color: var(--accent);
    font-size: clamp(16px, 3cqw, 32px);
    max-width: 60ch;
    line-height: 1.4;
  }
  @container (max-height: 320px) {
    .has-alert .forecast-strip,
    .has-alert.forecast-control {
      display: none;
    }
  }
  @container (max-height:230px) {
    .weather-content {
      gap: 16px;
    }
    .forecast-strip,
    .forecast-control {
      display: none;
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
    }
    .has-alert.weather-content {
      gap: 8px;
    }
  }
  @container (max-width:480px) {
    .weather-content {
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
  }
  @container (max-height:180px) {
    .weather-content {
      gap: 12px;
    }
  }
  @container (max-width:340px) {
    .weather-content {
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
    .weather-content {
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
    .has-alert .weather-current {
      display: none;
    }
  }
  @container (max-height: 100px) {
    .has-alert .alert-copy {
      font-size: 12px;
    }
  }
</style>
