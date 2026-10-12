<script lang="ts">
  import Icon from './Icon.svelte';
  import { catalog } from '../lib/catalog.ts';
  import type { Placement, Size, WidgetId } from '../lib/types.ts';
  let {
    items,
    selected,
    size,
    existing,
    custom,
    canFit,
    fits,
    onchoose,
    onsize,
    onpreview,
    onclose,
    onback,
    onremove,
  }: {
    items: Placement[];
    selected: WidgetId | null;
    size: Size | null;
    existing: boolean;
    custom: boolean;
    canFit: boolean;
    fits: Record<Size, boolean>;
    onchoose: (id: WidgetId) => void;
    onsize: (size: Size | null) => void;
    onpreview: () => void;
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
  <div class="picker-intro">
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
    <p class="quiet">
      {selected
        ? 'Choose a size, then a position on the board.'
        : 'Select a widget to preview its position, or edit one already on your board.'}
    </p>
  </div>
  {#if selected}
    <div
      class="size-options"
      class:has-current={custom}
      aria-label="Widget size"
    >
      {#if custom}<button
          class:selected={size === null}
          aria-pressed={size === null}
          onclick={() => onsize(null)}
          >Current<small>Keep your size</small></button
        >{/if}
      {#each ['small', 'medium', 'large'] as value}<button
          class:selected={size === value}
          aria-pressed={size === value}
          onclick={() => onsize(value as Size)}
          >{value}<small
            >{fits[value as Size] ? 'Fits' : 'No room'}{selected === 'sports' &&
            value === 'small'
              ? ' · Strip'
              : ''}</small
          ></button
        >{/each}
    </div>
    {#if !canFit}<p class="placement-message" role="status">
        {Object.values(fits).some(Boolean)
          ? 'This size won’t fit. Choose a smaller size, or resize another widget.'
          : 'No open area fits this widget. Go back to edit another widget and make room, or cancel to choose another board.'}
      </p>{/if}
    {#if canFit}<button class="primary" onclick={onpreview}
        >Choose position<Icon name="arrow" size={18} /></button
      >{/if}
    <div class="picker-actions">
      <button onclick={onback}>Back to widgets</button>{#if existing}<button
          class="danger"
          onclick={onremove}>Remove widget</button
        >{/if}
    </div>
  {:else}
    <div class="catalog">
      {#each catalog as widget}{@const present = items.some(
          (p) => p.id === widget.id,
        )}<button
          class="catalog-row"
          class:on-board={present}
          onclick={() => onchoose(widget.id)}
          ><Icon name={widget.id} /><span
            ><strong>{widget.title}</strong><small>{widget.description}</small
            ></span
          ><span class:select-label={!present}
            >{present ? 'Edit' : 'Select'}</span
          ></button
        >{/each}
    </div>
    <p class="picker-footnote">Changes take effect when you save.</p>
  {/if}
</aside>
