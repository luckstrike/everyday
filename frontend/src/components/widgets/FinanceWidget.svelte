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
      <p class="widget-title">{bills[0].name}</p>
      <p class="bill-amount">{money(bills[0].amount)}</p>
    </div>
    <p class="quiet">
      {day(bills[0].dueAt, now)} · {bills[0].autopay
        ? 'Autopay on'
        : 'Manual payment'}
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
