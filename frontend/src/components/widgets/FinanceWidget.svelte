<script lang="ts">
  import type { DashboardState, Size, SourceStatus } from '../../lib/types.ts';
  import Icon from '../Icon.svelte';
  import { day, money } from '../../lib/format.ts';
  import { daysLateText, orderBills, overdueCount } from '../../lib/bills.ts';
  import { sourceNotice } from '../../lib/sources.ts';
  import { fittingItemCount } from '../../lib/fit.ts';
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
  // Bills are typed into a file rather than connected, so a missing file says so.
  const notice = $derived(
    sourceNotice(status, 'Bills', {
      unconfigured: {
        title: 'No bills yet',
        detail: 'Add some to your bills file.',
      },
      unavailableDetail: 'Check your bills file.',
    }),
  );
  // What was missed first, then what is left to pay, then what is already paid.
  const bills = $derived(orderBills(data.bills));
  const overdue = $derived(overdueCount(bills));
  let billList = $state<HTMLOListElement>();
  let visibleCount = $state(0);
  const message = $derived(
    notice ??
      (bills.length === 0
        ? { title: 'All clear', detail: 'No upcoming bills.' }
        : null),
  );

  // Show only the cards that fit whole, as Up next does.
  $effect(() => {
    const list = billList;
    if (!list || bills.length === 0) return;
    const rows = Array.from(list.children);
    const measure = () => {
      const gap = Number.parseFloat(getComputedStyle(list).rowGap) || 0;
      visibleCount = fittingItemCount(
        list.getBoundingClientRect().height,
        rows.map((row) => row.getBoundingClientRect().height),
        gap,
      );
    };
    const observer = new ResizeObserver(measure);
    observer.observe(list);
    for (const row of rows) observer.observe(row);
    measure();

    return () => observer.disconnect();
  });
</script>

<WidgetFrame title="Bills" {message}>
  {#snippet actions()}
    {#if !notice && overdue > 0}<span class="overdue-chip"
        ><Icon name="alert" size={12} />{overdue} overdue</span
      >{/if}
  {/snippet}
  <ol class="bills" class:large={size === 'large'} bind:this={billList}>
    {#each bills as bill, i}<li
        class="bill"
        class:paid={bill.paid}
        class:overdue={bill.overdue}
        style:visibility={i < visibleCount ? 'visible' : 'hidden'}
        aria-hidden={i >= visibleCount}
      >
        <div class="bill-main">
          <p class="bill-name" title={bill.name}>{bill.name}</p>
          <p
            class="bill-due quiet"
            class:accent={i === 0 && !bill.paid && !bill.overdue}
          >
            {#if bill.overdue}<span class="overdue-word"
                >{daysLateText(bill.daysLate)}</span
              >{:else}{day(bill.dueAt, now)}{/if}{#if bill.paid}{' · '}<span
                class="paid-word">Paid</span
              >{/if}
          </p>
        </div>
        <p class="bill-amount">
          {#if bill.paid}<Icon
              name="check"
              size={16}
            />{:else if bill.overdue}<Icon name="alert" size={16} />{/if}{money(
            bill.amount,
          )}
        </p>
      </li>{/each}
  </ol>
</WidgetFrame>

<style>
  .bills {
    list-style: none;
    margin: 0;
    padding: 0;
    flex: 1 1 0;
    min-height: 0;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  /* Each bill is a card, like an event in Up next: the widget's outline, one radius
     step smaller, with every other card filled with the raised surface colour. */
  .bill {
    flex-shrink: 0;
    min-width: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px;
    border: 1px solid var(--line);
    border-radius: 12px;
  }
  .bill:nth-child(odd) {
    background: var(--surface);
  }
  .bill-main {
    flex: 1 1 0;
    min-width: 0;
  }
  .bill-name {
    font-size: 18px;
    line-height: 1.25;
    font-weight: 400;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .bill .bill-due {
    font-size: 14px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .bill-amount {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 18px;
    line-height: 1.25;
    font-variant-numeric: tabular-nums;
  }
  /* A paid bill recedes, since there is nothing left to do about it: its name and
     amount go quiet and a tick in the accent colour says it is done. The "Paid" word
     and the tick carry the meaning, so it never depends on colour alone. */
  .bill.paid .bill-name,
  .bill.paid .bill-amount {
    color: var(--muted);
  }
  .paid-word {
    color: var(--accent);
  }
  .bill-amount :global(svg) {
    color: var(--accent);
  }
  /* A missed bill gets the warm colour on its border, the days it is late, and an icon
     by its amount, so it reads without relying on colour alone. */
  .bill.overdue {
    border-color: var(--warn-line);
  }
  .overdue-word,
  .bill.overdue .bill-amount :global(svg) {
    color: var(--warn);
  }
  .overdue-chip {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 13px;
    line-height: 20px;
    color: var(--warn);
  }
  .bills.large {
    gap: 10px;
  }
  .large .bill {
    padding: 16px 14px;
  }
  .large .bill-name,
  .large .bill-amount {
    font-size: 22px;
  }
  .large .bill .bill-due {
    font-size: 16px;
  }
</style>
