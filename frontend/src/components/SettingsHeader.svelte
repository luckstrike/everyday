<script lang="ts">
  import type { Snippet } from 'svelte';
  // The top of every settings section: what this is on the left, its main action on
  // the right, and a rule underneath before the content begins.
  let {
    title,
    description = '',
    actions,
  }: { title: string; description?: string; actions?: Snippet } = $props();
</script>

<header class="settings-header">
  <div class="titles">
    <h3>{title}</h3>
    {#if description}<p class="quiet">{description}</p>{/if}
  </div>
  <div class="actions">
    {#if actions}{@render actions()}{/if}
  </div>
</header>

<style>
  .settings-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 24px;
    padding-bottom: 20px;
    border-bottom: 1px solid var(--line);
  }
  .titles {
    min-width: 0;
  }
  h3 {
    margin: 0;
    font-size: 24px;
    font-weight: 400;
    line-height: 1.25;
  }
  p {
    margin-top: 4px;
    line-height: 1.4;
  }
  .actions {
    flex-shrink: 0;
  }
  .actions:empty {
    display: none;
  }
  /* On a short screen, such as a landscape LCD, height is what runs out first: keep
     the title and the rule, and drop the description. */
  @media (max-height: 560px) {
    .settings-header {
      padding-bottom: 12px;
    }
    h3 {
      font-size: 20px;
    }
    p {
      display: none;
    }
  }
</style>
