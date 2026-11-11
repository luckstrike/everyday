<script lang="ts">
  import type { Snippet } from 'svelte';
  import WidgetHeader from './WidgetHeader.svelte';

  let {
    title,
    actions,
    children,
  }: {
    title: string;
    actions?: Snippet;
    children: Snippet;
  } = $props();
</script>

<div class="widget-frame">
  <WidgetHeader {title} {actions} />
  <div class="widget-content">{@render children()}</div>
</div>

<style>
  .widget-frame {
    --widget-padding-block: var(--pad, 28px);
    --widget-padding-inline: var(--pad, 28px);
    --widget-header-gap: 14px;
    --widget-header-height: 20px;
    --widget-content-height: calc(
      100cqh - 2 * var(--widget-padding-block) - var(--widget-header-height) -
        var(--widget-header-gap)
    );
    height: 100%;
    padding: var(--widget-padding-block) var(--widget-padding-inline);
    display: flex;
    flex-direction: column;
    gap: var(--widget-header-gap);
    min-width: 0;
    overflow: hidden;
  }
  .widget-content {
    display: flex;
    flex-direction: column;
    justify-content: center;
    flex: 1 1 0;
    min-width: 0;
    min-height: 0;
    gap: 14px;
  }
  @container (max-width: 340px) {
    .widget-frame {
      --widget-padding-block: 16px;
      --widget-padding-inline: 16px;
      --widget-header-gap: 10px;
    }
  }
  @container (max-height: 150px) {
    .widget-frame {
      --widget-padding-block: 16px;
      --widget-padding-inline: 20px;
      --widget-header-gap: 8px;
    }
  }
  @container (max-height: 100px) {
    .widget-frame {
      --widget-padding-block: 10px;
      --widget-padding-inline: 16px;
      --widget-header-gap: 4px;
    }
  }
</style>
