<script lang="ts">
  import MonthCalendar from '../MonthCalendar.svelte';
  import WidgetHeader from '../WidgetHeader.svelte';
  let { now }: { now: number } = $props();
  const title = $derived(
    new Date(now).toLocaleDateString('en-US', {
      month: 'long',
      year: 'numeric',
    }),
  );
</script>

<!-- Roomy widgets show the whole month; short ones show only the current week. -->
<div class="widget-body month">
  <WidgetHeader {title} />
  <div class="widget-content month-content">
    <div class="full"><MonthCalendar {now} /></div>
    <div class="week"><MonthCalendar {now} week /></div>
  </div>
</div>

<style>
  .month-content {
    align-items: center;
    width: 100%;
  }
  .full,
  .week {
    width: 100%;
    max-width: 420px;
  }
  .week {
    display: none;
  }
  @container (max-height: 279px) {
    .full {
      display: none;
    }
    .week {
      display: block;
    }
  }
</style>
