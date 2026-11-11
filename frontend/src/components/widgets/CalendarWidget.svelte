<script lang="ts">
  import MonthCalendar from '../MonthCalendar.svelte';
  let { now }: { now: number } = $props();
</script>

<!-- Roomy widgets show the whole month; short ones show only the current week. -->
<div class="widget-body month">
  <div class="full"><MonthCalendar {now} /></div>
  <div class="week"><MonthCalendar {now} week /></div>
</div>

<style>
  .month {
    align-items: center;
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
