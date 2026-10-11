<script lang="ts">
  import type { DashboardState, Size } from '../../lib/types.ts';
  import { day, time } from '../../lib/format.ts';
  let {
    data,
    size,
    now,
  }: { data: DashboardState['sports']; size: Size; now: number } = $props();
</script>

<div class="widget-body sports">
  <div>
    <p class="eyebrow">
      Dallas basketball{#if data.status === 'live'}<span class="accent"
          >Live</span
        >{/if}
    </p>
    <p class="widget-title">
      <span class="team-full">{data.home} vs. {data.away}</span><span
        class="team-short">{data.homeCode} vs. {data.awayCode}</span
      >
    </p>
  </div>
  <div class="sports-meta">
    <p class="accent">
      {data.status === 'scheduled'
        ? `${day(data.startsAt, now)} · ${time(data.startsAt)}`
        : `${data.homeScore} : ${data.awayScore}`}
    </p>
    <p class="quiet">{data.detail.replace(' · Dallas', '')}</p>
  </div>
  {#if size === 'large'}<p class="quiet">{data.league} · Mock game</p>{/if}
</div>
