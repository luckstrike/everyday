<script lang="ts">
  import type { DashboardState, Size, SourceStatus } from '../../lib/types.ts';
  import { day, money } from '../../lib/format.ts';
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
  const bills = $derived(
    [...data.bills].sort((a, b) => Date.parse(a.dueAt) - Date.parse(b.dueAt)),
  );
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
  <ol class="bills" class:large={size === 'large'} bind:this={billList}>
    {#each bills as bill, i}<li
        class="bill"
        style:visibility={i < visibleCount ? 'visible' : 'hidden'}
        aria-hidden={i >= visibleCount}
      >
        <div class="bill-main">
          <p class="bill-name" title={bill.name}>{bill.name}</p>
          <p class="bill-due quiet" class:accent={i === 0}>
            {day(bill.dueAt, now)} · {bill.autopay
              ? 'Autopay on'
              : 'Manual payment'}
          </p>
        </div>
        <p class="bill-amount">{money(bill.amount)}</p>
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
    font-size: 18px;
    line-height: 1.25;
    font-variant-numeric: tabular-nums;
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
