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

<style>
  @container (max-height: 64px) {
    .sports {
      padding: 8px 16px;
    }
    .sports .eyebrow,
    .sports-meta .quiet {
      display: none;
    }
    .sports .widget-title,
    .sports-meta > p:first-child {
      font-size: 16px;
    }
    :global(.editing) .sports-meta {
      padding-right: 32px;
    }
  }
  @container (max-width: 340px) and (max-height: 150px) {
    .sports {
      gap: 8px;
    }
    .sports .eyebrow {
      font-size: 12px;
      margin-bottom: 4px;
    }
    .sports-meta > p:first-child {
      font-size: 14px;
    }
    .sports-meta .quiet {
      display: none;
    }
  }
</style>
