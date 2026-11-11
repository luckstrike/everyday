<script lang="ts">
  import Widget from './Widget.svelte';
  import { GRID, layout, widgetTitles } from '../lib/layout.ts';
  import type { DashboardState, Placement, Size } from '../lib/types.ts';
  let { dashboard, now }: { dashboard: DashboardState; now: number } = $props();
  let gridWidth = $state(1280),
    gridHeight = $state(800);
  // Widgets adapt to the space they receive, not to a fixed screen size.
  function sizeFor(item: Placement): Size {
    const width = (gridWidth * item.w) / GRID;
    const height = (gridHeight * item.h) / GRID;
    return width > 500 && height > 360
      ? 'large'
      : width > 260 && height > 180
        ? 'medium'
        : 'small';
  }
  const position = (item: Placement) =>
    `grid-column:${item.x + 1}/span ${item.w};grid-row:${item.y + 1}/span ${item.h};`;
</script>

<div
  class="dashboard-grid"
  bind:clientWidth={gridWidth}
  bind:clientHeight={gridHeight}
>
  {#each layout as item (item.id)}<section
      class="widget"
      class:small={sizeFor(item) === 'small'}
      style={position(item)}
      aria-label={widgetTitles[item.id]}
    >
      <Widget id={item.id} {dashboard} {now} size={sizeFor(item)} />
    </section>{/each}
</div>
