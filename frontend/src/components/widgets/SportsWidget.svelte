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
  <p class="eyebrow">
    Courtside <span class:live={data.status === 'live'}
      >{data.status === 'live' ? '● Live' : data.league}</span
    >
  </p>
  {#if size === 'small'}
    <p class="widget-title">
      {data.home} <span class="quiet">vs</span>
      {data.away}
    </p>
    <p class="quiet">
      {data.status === 'scheduled'
        ? `${day(data.startsAt, now)} · ${time(data.startsAt)}`
        : `${data.homeScore} : ${data.awayScore} · ${data.detail}`}
    </p>
  {:else}
    <div class="teams">
      <div>
        <span class="team-mark">CHI</span><span>{data.home}</span><strong
          >{data.status === 'scheduled' ? '—' : data.homeScore}</strong
        >
      </div>
      <div>
        <span class="team-mark away">MIL</span><span>{data.away}</span><strong
          >{data.status === 'scheduled' ? '—' : data.awayScore}</strong
        >
      </div>
    </div>
    <p class="quiet sports-detail">
      {data.status === 'scheduled'
        ? `${day(data.startsAt, now)} · ${time(data.startsAt)}`
        : data.detail}
    </p>
    {#if size === 'large'}<div class="section-note">
        <p class="eyebrow">
          {data.status === 'scheduled' ? 'Next tip-off' : 'Game status'}
        </p>
        <p class="widget-title">
          {data.status === 'scheduled'
            ? data.detail
            : data.status === 'final'
              ? 'That’s a Chicago win.'
              : 'A close one in Chicago.'}
        </p>
        <p class="quiet">Mock game · Scores are simulated</p>
      </div>{/if}
  {/if}
</div>
