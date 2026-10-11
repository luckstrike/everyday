<script lang="ts">
  import { onDestroy } from 'svelte';
  import type { Candidate } from '../lib/priority.ts';
  import Widget from './Widget.svelte';
  import Icon from './Icon.svelte';
  import { canPlace, GRID, overlaps } from '../lib/boards.ts';
  import type {
    DashboardState,
    Placement,
    Size,
    WidgetDefinition,
    WidgetId,
  } from '../lib/types.ts';
  let {
    items,
    dashboard,
    now,
    editing,
    definitions,
    onchange,
    attention,
    previews = [],
    pendingId = null,
    onplace,
    onselect,
  }: {
    items: Placement[];
    dashboard: DashboardState;
    now: number;
    editing: boolean;
    definitions: WidgetDefinition[];
    onchange: (items: Placement[]) => void;
    attention?: Candidate;
    previews?: Placement[];
    pendingId?: WidgetId | null;
    onplace: (item: Placement) => void;
    onselect: (item: Placement) => void;
  } = $props();
  let grid: HTMLDivElement;
  let gridWidth = $state(1280),
    gridHeight = $state(800);
  let feedback = $state('');
  let cleanup = () => {};
  let dragged = false;
  const hints = $derived.by(() => {
    const result: Placement[] = [];
    for (const item of previews)
      if (!result.some((other) => overlaps(item, other))) result.push(item);
    return result;
  });
  function sizeFor(item: Placement): Size {
    const width = (gridWidth * item.w) / GRID;
    const height = (gridHeight * item.h) / GRID;
    return width > 500 && height > 360
      ? 'large'
      : width > 260 && height > 180
        ? 'medium'
        : 'small';
  }
  function change(item: Placement, patch: Partial<Placement>) {
    const candidate = { ...item, ...patch };
    if (canPlace(candidate, items, definitions)) {
      onchange(items.map((p) => (p.id === item.id ? candidate : p)));
      feedback = '';
    } else feedback = 'That position is occupied or outside the board.';
  }
  function pointer(event: PointerEvent, item: Placement) {
    if (event.button !== 0) return;
    cleanup();
    const button = event.currentTarget as HTMLButtonElement;
    button.setPointerCapture(event.pointerId);
    const rect = grid.getBoundingClientRect(),
      startX = event.clientX,
      startY = event.clientY;
    const original = items.map((p) => ({ ...p }));
    dragged = false;
    const move = (e: PointerEvent) => {
      const dx = Math.round((e.clientX - startX) / (rect.width / GRID)),
        dy = Math.round((e.clientY - startY) / (rect.height / GRID));
      if (dx || dy) {
        dragged = true;
        change(item, { x: item.x + dx, y: item.y + dy });
      }
    };
    const end = (e: PointerEvent) => {
      if (e.type === 'pointercancel' || e.type === 'lostpointercapture')
        onchange(original);
      cleanup();
    };
    cleanup = () => {
      button.removeEventListener('pointermove', move);
      button.removeEventListener('pointerup', end);
      button.removeEventListener('pointercancel', end);
      button.removeEventListener('lostpointercapture', end);
      cleanup = () => {};
    };
    button.addEventListener('pointermove', move);
    button.addEventListener('pointerup', end);
    button.addEventListener('pointercancel', end);
    button.addEventListener('lostpointercapture', end);
  }
  function keyboard(event: KeyboardEvent, item: Placement) {
    const delta: Record<string, [number, number]> = {
      ArrowLeft: [-1, 0],
      ArrowRight: [1, 0],
      ArrowUp: [0, -1],
      ArrowDown: [0, 1],
    };
    const d = delta[event.key];
    if (d) {
      event.preventDefault();
      change(item, { x: item.x + d[0], y: item.y + d[1] });
    }
  }
  const position = (item: Placement) =>
    `grid-column:${item.x + 1}/span ${item.w};grid-row:${item.y + 1}/span ${item.h};`;
  onDestroy(() => cleanup());
</script>

<div
  class="dashboard-grid"
  class:editing
  class:placing={!!pendingId}
  bind:this={grid}
  bind:clientWidth={gridWidth}
  bind:clientHeight={gridHeight}
>
  {#each items as item (item.id)}<section
      class="widget"
      class:pending={pendingId === item.id}
      class:emphasized={!editing && attention?.widget === item.id}
      class:small={sizeFor(item) === 'small'}
      style={position(item)}
      aria-label={definitions.find((d) => d.id === item.id)?.title}
    >
      <Widget id={item.id} {dashboard} {now} size={sizeFor(item)} />
      {#if editing}<button
          class="move-handle icon-button"
          data-widget-id={item.id}
          aria-label={`Edit ${item.id}, drag or use arrow keys to move`}
          onpointerdown={(e) => pointer(e, item)}
          onkeydown={(e) => keyboard(e, item)}
          onclick={() => {
            if (!dragged) onselect(item);
            dragged = false;
          }}><Icon name="grip" size={18} /></button
        >{/if}
    </section>{/each}
  {#each hints as item}<button
      class="placement-preview"
      style={position(item)}
      aria-label={`Place ${item.id} at column ${item.x + 1}, row ${item.y + 1}`}
      onclick={() => onplace(item)}
      ><Icon name="plus" /><span>Place here</span></button
    >{/each}
  {#if !items.length && !pendingId}<div class="empty-board">
      <p>A little room for your day.</p>
      <span>Tap Add widget to start this board.</span>
    </div>{/if}
</div>
<span class="sr-only" role="status"
  >{feedback}{pendingId && previews.length
    ? ` ${hints.length} available placement areas.`
    : ''}</span
>
