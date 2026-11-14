// Bills still to pay come first, soonest due first, then the ones already paid in the
// same order. This is the order the backend sends them in; the widget applies it again
// so it does not depend on that.
export function orderBills<T extends { dueAt: string; paid: boolean }>(
  bills: readonly T[],
): T[] {
  return [...bills].sort(
    (a, b) =>
      Number(a.paid) - Number(b.paid) ||
      Date.parse(a.dueAt) - Date.parse(b.dueAt),
  );
}
