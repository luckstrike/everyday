<script lang="ts">
  import MonthCalendar from '../MonthCalendar.svelte';
  import WidgetFrame from '../WidgetFrame.svelte';
  let { now }: { now: number } = $props();
  const title = $derived(
    new Date(now).toLocaleDateString('en-US', {
      month: 'long',
      year: 'numeric',
    }),
  );
</script>

<!-- Roomy widgets show the whole month; short ones show only the current week. -->
<WidgetFrame {title}>
  <div class="full"><MonthCalendar {now} /></div>
  <div class="week"><MonthCalendar {now} week /></div>
</WidgetFrame>

<style>
  .full,
  .week {
    flex: 1 1 0;
    min-height: 0;
    width: 100%;
    container-type: size;
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
