<script lang="ts">
  import type { DashboardState, Size, SourceStatus } from '../../lib/types.ts';
  import { day, money } from '../../lib/format.ts';
  import { sourceNotice } from '../../lib/sources.ts';
  import WidgetFrame from '../WidgetFrame.svelte';
  let {
    data,
    status,
    size,
    now,
  }: {
    data: DashboardState['finance'];
    status: SourceStatus;
    size: Size;
    now: number;
  } = $props();
  const notice = $derived(sourceNotice(status, 'Bills'));
  const bills = $derived(
    [...data.bills].sort((a, b) => Date.parse(a.dueAt) - Date.parse(b.dueAt)),
  );
  const message = $derived(
    notice ??
      (bills.length === 0
        ? { title: 'All clear', detail: 'No upcoming bills.' }
        : null),
  );
</script>

<WidgetFrame title="Bills" {message}>
  {#if bills[0]}<div class="bill-summary">
      <p class="widget-title" title={bills[0].name}>{bills[0].name}</p>
      <p class="bill-amount">{money(bills[0].amount)}</p>
    </div>
    <p class="quiet">
      {day(bills[0].dueAt, now)}<span class="payment-method">
        · {bills[0].autopay ? 'Autopay on' : 'Manual payment'}</span
      >
    </p>
    {#if size === 'large'}<div class="bill-list">
        {#each bills.slice(1, 4) as bill}<div>
            <span>{bill.name}<small>{day(bill.dueAt, now)}</small></span><span
              >{money(bill.amount)}</span
            >
          </div>{/each}
      </div>{/if}{/if}
</WidgetFrame>

<style>
  .bill-summary .widget-title {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  @container (max-height: 120px) {
    .payment-method {
      display: none;
    }
  }
</style>
