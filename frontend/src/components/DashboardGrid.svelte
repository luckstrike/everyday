<script lang="ts">
  import Widget from './Widget.svelte';
  import { canPlace, widgetSize } from '../lib/layout.ts';
  import type {
    DashboardState,
    Placement,
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
  }: {
    items: Placement[];
    dashboard: DashboardState;
    now: number;
    editing: boolean;
    definitions: WidgetDefinition[];
    onchange: (items: Placement[]) => void;
  } = $props();
  let grid: HTMLDivElement;
  let active = $state<WidgetId | null>(null);
  let feedback = $state('');
  const rows = $derived(Math.max(6, ...items.map((item) => item.y + item.h)));
  function change(item: Placement, patch: Partial<Placement>) {
    const candidate = { ...item, ...patch };
    if (canPlace(candidate, items, definitions)) {
      onchange(items.map((p) => (p.id === item.id ? candidate : p)));
      feedback = '';
    } else feedback = 'That space is occupied or outside the grid.';
  }
  function pointer(event: PointerEvent, item: Placement, resize = false) {
    if (event.button !== 0) return;
    event.preventDefault();
    const button = event.currentTarget as HTMLButtonElement;
    button.setPointerCapture(event.pointerId);
    const rect = grid.getBoundingClientRect(),
      startX = event.clientX,
      startY = event.clientY;
    const gap = parseFloat(getComputedStyle(grid).columnGap);
    const column = (rect.width + gap) / 12,
      row = (rect.height + gap) / rows;
    const origin = { ...item };
    active = item.id;
    const move = (e: PointerEvent) => {
      const dx = Math.round((e.clientX - startX) / column),
        dy = Math.round((e.clientY - startY) / row);
      change(
        origin,
        resize
          ? { w: origin.w + dx, h: origin.h + dy }
          : { x: origin.x + dx, y: origin.y + dy },
      );
    };
    const end = () => {
      active = null;
      button.removeEventListener('pointermove', move);
      button.removeEventListener('pointerup', end);
      button.removeEventListener('pointercancel', end);
    };
    button.addEventListener('pointermove', move);
    button.addEventListener('pointerup', end);
    button.addEventListener('pointercancel', end);
  }
  function keyboard(event: KeyboardEvent, item: Placement, resize = false) {
    const delta: Record<string, [number, number]> = {
      ArrowLeft: [-1, 0],
      ArrowRight: [1, 0],
      ArrowUp: [0, -1],
      ArrowDown: [0, 1],
    };
    const d = delta[event.key];
    if (!d) return;
    event.preventDefault();
    change(
      item,
      resize
        ? { w: item.w + d[0], h: item.h + d[1] }
        : { x: item.x + d[0], y: item.y + d[1] },
    );
  }
</script>

<div
  class="dashboard-grid"
  class:editing
  bind:this={grid}
  style={`--rows:${rows}`}
>
  {#each items as item (item.id)}
    <section
      class="widget"
      class:clock-panel={item.id === 'clock'}
      class:moving={active === item.id}
      class:large={widgetSize(item) === 'large'}
      class:small={widgetSize(item) === 'small'}
      style={`grid-column:${item.x + 1}/span ${item.w};grid-row:${item.y + 1}/span ${item.h};`}
      aria-label={definitions.find((d) => d.id === item.id)?.title}
    >
      {#if editing}<div class="edit-handles">
          <button
            class="move-handle"
            aria-label={`Move ${item.id}, use arrow keys`}
            onpointerdown={(e) => pointer(e, item)}
            onkeydown={(e) => keyboard(e, item)}
            >⠿ {item.id} <small>{item.w} × {item.h}</small></button
          ><button
            aria-label={`Remove ${item.id}`}
            onclick={() => onchange(items.filter((p) => p.id !== item.id))}
            >×</button
          >
        </div>{/if}
      <Widget id={item.id} {dashboard} {now} size={widgetSize(item)} />
      {#if editing}<button
          class="resize-handle"
          aria-label={`Resize ${item.id}, use arrow keys`}
          onpointerdown={(e) => pointer(e, item, true)}
          onkeydown={(e) => keyboard(e, item, true)}>↘</button
        >{/if}
    </section>
  {/each}
  {#if items.length === 0}<p class="empty-grid">
      This screen is empty. Open Edit layout to add a widget.
    </p>{/if}
</div>
{#if editing}<p class="edit-help" role="status">
    {feedback ||
      'Drag handles to move or resize. Focus a handle and use arrow keys for precise adjustments. Occupied cells are protected.'}
  </p>{/if}
