<script lang="ts">
  import type { DashboardState, Size } from '../../lib/types.ts';
  import { day, money } from '../../lib/format.ts';
  let {
    data,
    size,
    now,
  }: { data: DashboardState['finance']; size: Size; now: number } = $props();
  const bills = $derived(
    [...data.bills].sort((a, b) => Date.parse(a.dueAt) - Date.parse(b.dueAt)),
  );
</script>

<div class="widget-body finance">
  <p class="eyebrow">Next bill</p>
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
      </div>{/if}{:else}<p class="widget-title">All clear</p>
    <p class="quiet">No upcoming bills.</p>{/if}
</div>

<style>
  .bill-summary .widget-title {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  @container (max-height: 120px) {
    .finance .eyebrow,
    .payment-method {
      display: none;
    }
  }
</style>
