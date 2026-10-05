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
  const bill = $derived(bills[0]);
</script>

<div class="widget-body finance">
  <p class="eyebrow">On the horizon <span>Bills</span></p>
  {#if bill}
    <p class="widget-title">{bill.name}</p>
    <p class="bill-amount">{money(bill.amount)}</p>
    <p class="quiet">
      Due {day(bill.dueAt, now).toLowerCase()}{#if size !== 'small'}
        <span class="autopay"
          >{bill.autopay ? 'Autopay on' : 'Manual payment'}</span
        >{/if}
    </p>
  {:else}<p class="widget-title">All clear</p>
    <p class="quiet">No upcoming bills.</p>{/if}
  {#if size === 'large'}
    <div class="bill-list">
      {#each bills.slice(1) as other}<div>
          <span
            >{other.name}<small
              >{day(other.dueAt, now)} · {other.autopay
                ? 'Autopay'
                : 'Manual'}</small
            ></span
          ><strong>{money(other.amount)}</strong>
        </div>{/each}
    </div>
  {/if}
</div>
