<script lang="ts">
  import Icon from './Icon.svelte';
  import { catalog } from '../lib/catalog.ts';
  import type { Placement, Size, WidgetId } from '../lib/types.ts';
  let {
    items,
    selected,
    size,
    existing,
    fits,
    onchoose,
    onsize,
    onclose,
    onback,
    onremove,
  }: {
    items: Placement[];
    selected: WidgetId | null;
    size: Size;
    existing: boolean;
    fits: Record<Size, boolean>;
    onchoose: (id: WidgetId) => void;
    onsize: (size: Size) => void;
    onclose: () => void;
    onback: () => void;
    onremove: () => void;
  } = $props();
</script>

<aside
  class="widget-picker popover"
  aria-label={selected ? 'Widget placement' : 'Add a widget'}
  tabindex="-1"
>
  <div class="popover-heading">
    <h2>
      {selected
        ? `${existing ? 'Edit' : 'Place'} ${catalog.find((w) => w.id === selected)?.title}`
        : 'Add a widget'}
    </h2>
    <button class="icon-button" aria-label="Close picker" onclick={onclose}
      ><Icon name="close" /></button
    >
  </div>
  {#if selected}
    <p class="quiet">Tap a highlighted position to place it.</p>
    <div class="size-options" aria-label="Widget size">
      {#each ['small', 'medium', 'large'] as value}<button
          class:selected={size === value}
          aria-pressed={size === value}
          onclick={() => onsize(value as Size)}
          >{value}<small>{fits[value as Size] ? 'Fits' : 'No room'}</small
          ></button
        >{/each}
    </div>
    {#if !fits[size]}<p class="placement-message" role="status">
        {Object.values(fits).some(Boolean)
          ? 'This size won’t fit. Choose a smaller size, or resize another widget.'
          : 'This board is full. Remove or resize a widget, or cancel and create another board.'}
      </p>{/if}
    <div class="picker-actions">
      <button onclick={onback}>Back to widgets</button>{#if existing}<button
          class="danger"
          onclick={onremove}>Remove widget</button
        >{/if}
    </div>
  {:else}
    <p class="quiet">Tap a widget to see where it fits.</p>
    <div class="catalog">
      {#each catalog as widget}{@const present = items.some(
          (p) => p.id === widget.id,
        )}<button
          class="catalog-row"
          disabled={present}
          onclick={() => onchoose(widget.id)}
          ><Icon name={widget.id} /><span
            ><strong>{widget.title}</strong><small>{widget.description}</small
            ></span
          ><span class:select-label={!present}
            >{present ? 'On board' : 'Select'}</span
          ></button
        >{/each}
    </div>
    <p class="picker-footnote">Changes take effect when you save.</p>
  {/if}
</aside>
