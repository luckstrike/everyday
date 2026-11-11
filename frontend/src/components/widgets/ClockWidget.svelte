<script lang="ts">
  import WidgetHeader from '../WidgetHeader.svelte';
  let { now }: { now: number } = $props();
  const parts = $derived(
    new Intl.DateTimeFormat('en-US', {
      hour: 'numeric',
      minute: '2-digit',
      hour12: true,
    }).formatToParts(now),
  );
  const clock = $derived(
    parts
      .filter(
        (p) =>
          (p.type !== 'dayPeriod' && p.type !== 'literal') || p.value === ':',
      )
      .map((p) => p.value)
      .join(''),
  );
</script>

<div class="widget-body clock">
  <WidgetHeader title="Clock" />
  <div class="clock-summary">
    <div class="clock-time">
      <span>{clock}</span><small
        >{parts.find((p) => p.type === 'dayPeriod')?.value}</small
      >
    </div>
    <div class="clock-date-block">
      <div class="clock-date">
        <span
          >{new Date(now).toLocaleDateString('en-US', {
            weekday: 'long',
          })}</span
        ><span
          >{new Date(now).toLocaleDateString('en-US', {
            month: 'long',
            day: 'numeric',
          })}</span
        >
      </div>
    </div>
  </div>
</div>

<style>
  .clock {
    gap: 14px;
  }
  .clock-summary {
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 20px;
    width: 100%;
    min-width: 0;
    flex: 1 1 0;
    min-height: 0;
  }
  .clock-date-block {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-shrink: 0;
  }
  .clock-time {
    flex-shrink: 0;
  }
  .clock-time > span {
    font-size: max(
      24px,
      min(27cqw, calc(100cqh - 2 * var(--pad) - 90px), 192px)
    );
  }
  @container (min-aspect-ratio: 3) {
    .clock-summary {
      flex-direction: row;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
    }
    .clock-time > span {
      font-size: max(
        24px,
        min(27cqw, calc(100cqh - 2 * var(--pad) - 34px), 192px)
      );
    }
  }
  @container (max-width: 559px) and (max-aspect-ratio: 3) {
    .clock-summary {
      gap: 8px;
    }
    .clock-time > span {
      font-size: max(
        24px,
        min(24cqw, calc(100cqh - 2 * var(--pad) - 70px), 144px)
      );
    }
  }
  @container (max-height: 150px) {
    .clock-summary {
      gap: 8px;
    }
  }
  @container (max-width: 250px) and (min-height: 101px) {
    .clock-date-block {
      position: relative;
      padding-right: 32px;
    }
    .clock-date {
      flex-direction: column;
      align-items: flex-start;
      gap: 2px;
      font-size: 14px;
    }
    .clock-date span + span {
      font-size: 14px;
    }
  }
  @container (max-height: 100px) {
    .clock-summary {
      flex-direction: row;
      align-items: center;
      justify-content: space-between;
      gap: 8px;
    }
    .clock-date {
      display: none;
    }
    .clock-time > span {
      font-size: max(24px, min(24cqw, calc(100cqh - 44px), 96px));
    }
  }
</style>
