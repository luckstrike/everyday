// The rules for a bill being typed into the settings panel. They mirror the checks
// the backend makes, so a mistake is caught while the person is still looking at the
// form. The backend still has the last word, and its message is shown if it differs.
export interface BillEntry {
  name: string;
  amount: number;
  day: number;
  autopay: boolean;
}

// What is on screen while a bill is being edited. The amount stays text so
// half-typed values like "84." are not rewritten under the person's fingers.
export interface BillDraft {
  name: string;
  amount: string;
  day: number;
  autopay: boolean;
}

export const emptyDraft = (): BillDraft => ({
  name: '',
  amount: '',
  day: 1,
  autopay: false,
});

export const draftFrom = (entry: BillEntry): BillDraft => ({
  name: entry.name,
  amount: String(entry.amount),
  day: entry.day,
  autopay: entry.autopay,
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

export function entryFrom(draft: BillDraft): BillEntry {
  return {
    name: draft.name.trim(),
    amount: parseAmount(draft.amount) ?? 0,
    day: draft.day,
    autopay: draft.autopay,
  };
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
    ? `The ${ordinal(day)} of every month, or the last day of a shorter month.`
    : `The ${ordinal(day)} of every month.`;
}
