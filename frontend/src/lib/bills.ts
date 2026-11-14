// The order bills are shown in: missed bills first, the longest overdue at the top; then
// bills still to pay, soonest due; then bills already paid, in the same order. A display
// with little room spends it on what needs doing. The backend sends them in this order;
// the widget applies it again so it does not depend on that.
type Orderable = { dueAt: string; paid: boolean; overdue: boolean };

const rank = (bill: Orderable) => (bill.overdue ? 0 : bill.paid ? 2 : 1);

export function orderBills<T extends Orderable>(bills: readonly T[]): T[] {
  return [...bills].sort(
    (a, b) => rank(a) - rank(b) || Date.parse(a.dueAt) - Date.parse(b.dueAt),
  );
}

export const overdueCount = (bills: readonly { overdue: boolean }[]) =>
  bills.filter((bill) => bill.overdue).length;

// How late a missed bill is, in words short enough for a narrow card.
export const daysLateText = (days: number) =>
  days === 1 ? '1 day overdue' : `${days} days overdue`;
