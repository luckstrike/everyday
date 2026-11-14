<script lang="ts">
  import type { DashboardState, Size, SourceStatus } from '../../lib/types.ts';
  import Icon from '../Icon.svelte';
  import WidgetFrame from '../WidgetFrame.svelte';
  import TemperatureRange from '../TemperatureRange.svelte';
  import { sourceNotice } from '../../lib/sources.ts';
  import { fittingItemCount } from '../../lib/fit.ts';
  import { weatherIconName } from '../../lib/weather.ts';
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
  let dailyStrip = $state<HTMLDivElement>();
  let weatherArea = $state<HTMLDivElement>();
  let currentBlock = $state<HTMLDivElement>();
  let alertCopy = $state<HTMLParagraphElement>();
  let visibleDayCount = $state(0);
  // Whether the forecast, and the switch that chooses between days and hours, fit
  // whole beneath the current conditions. Measured, never guessed from a height.
  let forecastFits = $state(true);
  const notice = $derived(sourceNotice(status, 'Weather'));

  $effect(() => {
    const strip = dailyStrip;
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

  // The daily cards always keep their natural height, so measuring them says what
  // the forecast needs even while it is hidden. If everything above it plus the
  // cards does not fit the space available, the forecast is hidden rather than
  // squeezed, so a card is never shown clipped. The hourly strip is shorter than
  // the daily one, so the daily height is the requirement for both.
  $effect(() => {
    const area = weatherArea;
    const current = currentBlock;
    const daily = dailyStrip;
    if (!area || !current || !daily || data.daily.length === 0) return;
    const measure = () => {
      const gap = Number.parseFloat(getComputedStyle(area).rowGap) || 0;
      const blocks = [current.getBoundingClientRect().height];
      if (alertCopy) blocks.push(alertCopy.getBoundingClientRect().height);
      blocks.push(daily.getBoundingClientRect().height);
      // The widget's own padding absorbs a small shortfall: the cards keep their full
      // size and sit a little nearer the edge, which looks better than squeezing
      // them. Only half the padding is borrowed, so a margin always remains.
      const padding = Number.parseFloat(
        getComputedStyle(area.closest('.widget-frame') ?? area).paddingBottom,
      );
      const room = area.getBoundingClientRect().height + (padding || 0) / 2;
      forecastFits = fittingItemCount(room, blocks, gap) === blocks.length;
    };
    const observer = new ResizeObserver(measure);
    observer.observe(area);
    observer.observe(current);
    observer.observe(daily);
    if (alertCopy) observer.observe(alertCopy);
    measure();

    return () => observer.disconnect();
  });
</script>

<WidgetFrame title="Weather" message={notice}>
  {#snippet actions()}
    {#if !notice && forecastFits}<div
        class="forecast-control"
        class:has-alert={!!data.alert}
      >
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
  <div
    class="weather-content"
    class:has-alert={!!data.alert}
    bind:this={weatherArea}
  >
    <div class="weather-current" bind:this={currentBlock}>
      <span class="weather-icon accent"
        ><Icon name={weatherIconName(data.icon)} size={56} /></span
      ><span class="temperature">{data.temperature}°</span>
      <div>
        <p class="widget-title">{data.condition}</p>
        <p class="quiet">
          <TemperatureRange low={data.low} high={data.high} />
        </p>
      </div>
    </div>
    {#if data.alert}<p class="alert-copy" bind:this={alertCopy}>
        {data.alert}
      </p>{/if}
    <!-- Always rendered so its height can be measured; it leaves the layout, unseen,
         when hours are showing or the forecast does not fit. -->
    <div
      class="forecast-strip daily"
      class:forecast-hidden={forecast !== 'days' || !forecastFits}
      aria-hidden={forecast !== 'days' || !forecastFits}
      inert={forecast !== 'days' || !forecastFits}
      bind:this={dailyStrip}
    >
      {#each data.daily.slice(0, 5) as day, i}<div
          class="forecast-day"
          class:overflow-day={i >= visibleDayCount}
          style:visibility={i < visibleDayCount ? 'visible' : 'hidden'}
          aria-hidden={i >= visibleDayCount}
        >
          <div class="forecast-day-content">
            <div class="forecast-day-head">
              <span class="forecast-day-label"
                >{i === 0
                  ? 'Today'
                  : new Date(day.date).toLocaleDateString('en-US', {
                      weekday: 'short',
                    })}</span
              ><span class="forecast-day-icon" title={day.condition}
                ><Icon name={weatherIconName(day.icon)} size={22} /></span
              ><span class="sr-only">{day.condition}</span>
            </div>
            <strong><TemperatureRange low={day.low} high={day.high} /></strong
            >{#if size === 'large' && data.alert}<small>{day.condition}</small
              >{/if}
          </div>
        </div>{/each}
    </div>
    {#if forecast === 'hours' && forecastFits}<div class="forecast-strip">
        {#each data.hourly.slice(0, 4) as hour}<div>
            <span>{hour.label}</span><strong>{hour.temperature}°</strong
            >{#if size === 'large' && data.alert}<small>{hour.condition}</small
              >{/if}
          </div>{/each}
      </div>{/if}
  </div>
</WidgetFrame>

<style>
  .weather-content {
    position: relative;
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
    /* Keep the natural height, so the forecast is hidden to fit and never squeezed. */
    flex-shrink: 0;
  }
  /* Out of the layout and unseen, but still laid out at full width so it can be
     measured. */
  .forecast-strip.forecast-hidden {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    visibility: hidden;
    pointer-events: none;
  }
  /* Each card sets its own visibility inline (to hide the days that do not fit across),
     and a child's own setting beats its parent's. Without this, "hidden" cards that
     fit across would still be drawn, on top of the current conditions. */
  .forecast-strip.forecast-hidden .forecast-day {
    visibility: hidden !important;
  }
  .forecast-day {
    flex: 1 0 auto;
    min-width: 0;
    /* A card is wider than its text whenever the widget has room to spare. Centering
       the text spreads that spare width evenly instead of leaving it on the right. */
    align-items: center;
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
  /* Day label on the left and the day's weather on the right, so the icon adds
     no width to the card and the temperatures keep their full size. The icon is
     taller than the label's line, so negative margins let it overhang into the
     card's padding instead of making every card taller. */
  .forecast-day-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }
  .forecast-strip .forecast-day-icon {
    display: flex;
    margin-block: -3px;
    color: var(--accent);
  }
  .forecast-day-icon :global(svg) {
    display: block;
  }
  .forecast-strip.daily strong {
    font-size: clamp(18px, 3cqw, 24px);
  }
  /* On a wide widget the cards are much wider than their text, which used to leave
     an empty strip on the right of each one. There the icon leaves the corner and
     sits large beside the day and temperatures, so the card's width is used. */
  @container (min-width: 960px) {
    .forecast-day-content {
      display: grid;
      grid-template-columns: auto auto;
      grid-template-areas: 'icon label' 'icon range';
      align-items: center;
      justify-content: start;
      column-gap: clamp(10px, 1.2cqw, 16px);
      row-gap: 8px;
    }
    .forecast-day-head {
      display: contents;
    }
    .forecast-day-label {
      grid-area: label;
    }
    .forecast-strip .forecast-day-icon {
      grid-area: icon;
      margin-block: 0;
    }
    .forecast-day-icon :global(svg) {
      width: clamp(32px, 3.2cqw, 48px);
      height: clamp(32px, 3.2cqw, 48px);
    }
    .forecast-day-content > strong {
      grid-area: range;
    }
    .forecast-day-content > small {
      grid-column: 1 / -1;
    }
  }
  /* Mid-width widgets, just short of the side-by-side layout, still have a little
     spare width per card, so the temperatures grow to use it. */
  @container (min-width: 780px) and (max-width: 959px) and (min-height: 280px) {
    .forecast-strip.daily strong {
      font-size: 26px;
    }
  }
  /* Taller as well as wider: the temperatures can grow too without needing more
     height than the widget has. */
  @container (min-width: 960px) and (min-height: 300px) {
    .forecast-strip.daily strong {
      font-size: clamp(24px, 2.2cqw, 32px);
    }
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
    .forecast-strip .forecast-day-icon {
      margin-block: -1px;
    }
    .forecast-day-icon :global(svg) {
      width: 18px;
      height: 18px;
    }
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
    .forecast-strip .forecast-day-icon {
      display: none;
    }
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
    .forecast-strip .forecast-day-icon {
      display: none;
    }
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
