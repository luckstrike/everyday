<script lang="ts">
  import type { Snippet } from 'svelte';

  let { title, actions }: { title: string; actions?: Snippet } = $props();
</script>

<header class="widget-header">
  <h2 {title}>{title}</h2>
  {#if actions}<div class="widget-actions">{@render actions()}</div>{/if}
</header>

<style>
  .widget-header {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-shrink: 0;
    min-width: 0;
    min-height: 20px;
  }
  h2 {
    flex: 1 1 0;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 16px;
    line-height: 20px;
    font-weight: 400;
    color: var(--muted);
  }
  .widget-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    height: 20px;
    gap: 8px;
    flex-shrink: 0;
  }
  @container (max-width: 340px) {
    h2 {
      font-size: 14px;
    }
  }
  @container (max-height: 150px) {
    h2 {
      font-size: 14px;
    }
  }
</style>
