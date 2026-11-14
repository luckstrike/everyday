// The rules for a bill being typed into the settings panel. They mirror the checks
// the backend makes, so a mistake is caught while the person is still looking at the
// form. The backend still has the last word, and its message is shown if it differs.
export interface BillEntry {
  name: string;
  amount: number;
  day: number;
  // The due date (YYYY-MM-DD) this bill was last marked paid for. It only counts
  // while it is the due date the bill is currently counting toward, so a tick clears
  // itself when the next month starts.
  paidFor?: string;
}

// A bill as the backend lists it for editing: what is saved, plus the due date it is
// counting toward now and whether that one is paid.
export interface ListedBill extends BillEntry {
  due: string; // YYYY-MM-DD
  paid: boolean;
}

// What is saved for a bill: only what belongs in the file, not the dates worked out
// for display.
export function toEntry(bill: ListedBill): BillEntry {
  const entry: BillEntry = {
    name: bill.name,
    amount: bill.amount,
    day: bill.day,
  };
  if (bill.paidFor) entry.paidFor = bill.paidFor;
  return entry;
}

// The same bill, marked paid or not paid for the due date it is counting toward.
// Marking it paid records that date; unmarking forgets it.
export function withPaid(bill: ListedBill, paid: boolean): BillEntry {
  const entry = toEntry(bill);
  if (paid) entry.paidFor = bill.due;
  else delete entry.paidFor;
  return entry;
}

// What is on screen while a bill is being edited. The amount stays text so
// half-typed values like "84." are not rewritten under the person's fingers.
export interface BillDraft {
  name: string;
  amount: string;
  day: number;
}

export const emptyDraft = (): BillDraft => ({
  name: '',
  amount: '',
  day: 1,
});

export const draftFrom = (entry: BillEntry): BillDraft => ({
  name: entry.name,
  amount: entry.amount.toFixed(2),
  day: entry.day,
});

// Reads an amount the way people write one: "1850", "1,850.50", "$84.5". Returns
// null when it is not a usable amount.
export function parseAmount(text: string): number | null {
  const cleaned = text.trim().replace(/^\$/, '').replace(/,/g, '');
  if (!/^(\d+\.?\d*|\.\d+)$/.test(cleaned)) return null;
  const amount = Number(cleaned);
  return Number.isFinite(amount) ? Math.round(amount * 100) / 100 : null;
}

// A message to show when the draft cannot be saved, or null when it can.
export function validateDraft(draft: BillDraft): string | null {
  if (draft.name.trim() === '') return 'Give the bill a name.';
  if (parseAmount(draft.amount) === null)
    return 'Enter the amount as a number, like 84.50.';
  if (!Number.isInteger(draft.day) || draft.day < 1 || draft.day > 31)
    return 'Pick a day from 1 to 31.';
  return null;
}

// The bill the draft describes. When it is an edit of an existing bill, that bill's
// paid mark carries over, so changing a name or amount does not un-pay it.
export function entryFrom(draft: BillDraft, existing?: BillEntry): BillEntry {
  const entry: BillEntry = {
    name: draft.name.trim(),
    amount: parseAmount(draft.amount) ?? 0,
    day: draft.day,
  };
  if (existing?.paidFor) entry.paidFor = existing.paidFor;
  return entry;
}

// "1st", "2nd", "3rd", "11th", "22nd": for reading a due day aloud.
export function ordinal(n: number): string {
  const lastTwo = n % 100;
  if (lastTwo >= 11 && lastTwo <= 13) return `${n}th`;
  return `${n}${({ 1: 'st', 2: 'nd', 3: 'rd' } as Record<number, string>)[n % 10] ?? 'th'}`;
}

// Says when a bill falls due, including what happens at the end of a short month.
export function dueNote(day: number): string {
  return day > 28
    ? `${ordinal(day)} of every month, or the last day of a shorter month.`
    : `${ordinal(day)} of every month.`;
}

// A bill with where it sits in the saved list, which is what editing refers to.
export interface PlacedBill<T extends BillEntry = BillEntry> {
  bill: T;
  index: number;
  // Its place in the date order shown on screen, for alternating row fills.
  position: number;
}

// Bills grouped by the day of the month they fall due, earliest day first. Bills
// that share a day sit together, in name order. The saved list is left as it is:
// each bill keeps its index in it, so an edit changes the right one.
export function groupByDay<T extends BillEntry>(
  entries: readonly T[],
): { day: number; bills: PlacedBill<T>[] }[] {
  const ordered = entries
    .map((bill, index) => ({ bill, index }))
    .sort(
      (a, b) =>
        a.bill.day - b.bill.day ||
        a.bill.name.localeCompare(b.bill.name, undefined, {
          sensitivity: 'base',
        }) ||
        a.index - b.index,
    );
  const groups: { day: number; bills: PlacedBill<T>[] }[] = [];
  ordered.forEach(({ bill, index }, position) => {
    const last = groups[groups.length - 1];
    const placed = { bill, index, position };
    if (last && last.day === bill.day) last.bills.push(placed);
    else groups.push({ day: bill.day, bills: [placed] });
  });
  return groups;
}
