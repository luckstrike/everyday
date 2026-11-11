<script lang="ts">
  import type { DashboardState, Size, SourceStatus } from '../../lib/types.ts';
  import Icon from '../Icon.svelte';
  import WidgetFrame from '../WidgetFrame.svelte';
  import TemperatureRange from '../TemperatureRange.svelte';
  import { sourceNotice } from '../../lib/sources.ts';
  import { fittingItemCount } from '../../lib/fit.ts';
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
  let forecastStrip = $state<HTMLDivElement>();
  let visibleDayCount = $state(0);
  const notice = $derived(sourceNotice(status, 'Weather'));

  $effect(() => {
    const strip = forecastStrip;
    if (!strip || forecast !== 'days' || data.daily.length === 0) return;
    const days = Array.from(strip.children);
    const measure = () => {
      const gap = Number.parseFloat(getComputedStyle(strip).columnGap) || 0;
      visibleDayCount = fittingItemCount(
        strip.getBoundingClientRect().width,
        days.map((day) => {
          const styles = getComputedStyle(day);
          return (
            day.firstElementChild!.getBoundingClientRect().width +
            Number.parseFloat(styles.paddingLeft) +
            Number.parseFloat(styles.paddingRight) +
            Number.parseFloat(styles.borderLeftWidth) +
            Number.parseFloat(styles.borderRightWidth)
          );
        }),
        gap,
      );
    };
    const observer = new ResizeObserver(measure);
    observer.observe(strip);
    for (const day of days) {
      observer.observe(day);
      observer.observe(day.firstElementChild!);
    }
    measure();

    return () => observer.disconnect();
  });
</script>

<WidgetFrame title="Weather" message={notice}>
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
    <div class="weather-current">
      <span class="weather-icon accent"><Icon name="weather" size={56} /></span
      ><span class="temperature">{data.temperature}°</span>
      <div>
        <p class="widget-title">{data.condition}</p>
        <p class="quiet">
          <TemperatureRange low={data.low} high={data.high} />
        </p>
      </div>
    </div>
    {#if data.alert}<p class="alert-copy">{data.alert}</p>{/if}
    <div
      class="forecast-strip"
      class:daily={forecast === 'days'}
      bind:this={forecastStrip}
    >
      {#if forecast === 'hours'}{#each data.hourly.slice(0, 4) as hour}<div>
            <span>{hour.label}</span><strong>{hour.temperature}°</strong
            >{#if size === 'large' && data.alert}<small>{hour.condition}</small
              >{/if}
          </div>{/each}{:else}{#each data.daily.slice(0, 5) as day, i}<div
            class="forecast-day"
            class:overflow-day={i >= visibleDayCount}
            style:visibility={i < visibleDayCount ? 'visible' : 'hidden'}
            aria-hidden={i >= visibleDayCount}
          >
            <div class="forecast-day-content">
              <span
                >{i === 0
                  ? 'Today'
                  : new Date(day.date).toLocaleDateString('en-US', {
                      weekday: 'short',
                    })}</span
              ><strong
                ><TemperatureRange low={day.low} high={day.high} /></strong
              >{#if size === 'large' && data.alert}<small>{day.condition}</small
                >{/if}
            </div>
          </div>{/each}{/if}
    </div>
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
  .forecast-strip.daily {
    display: flex;
    position: relative;
    overflow: hidden;
    gap: 8px;
  }
  .forecast-day {
    flex: 1 0 auto;
    min-width: 0;
    padding: 12px clamp(4px, 1cqw, 8px);
    border: 1px solid var(--line);
    border-radius: 12px;
  }
  .forecast-day:nth-child(odd) {
    background: var(--surface);
  }
  .forecast-day.overflow-day {
    position: absolute;
  }
  .forecast-day-content {
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: max-content;
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
  .forecast-strip.daily strong {
    font-size: clamp(18px, 3cqw, 24px);
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
    .forecast-strip.daily {
      gap: 4px;
    }
    .forecast-strip span {
      font-size: 12px;
    }
    .forecast-strip strong {
      font-size: 18px;
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
