import type { AttentionEvent, DashboardState } from './types.ts';

// Compare local calendar dates, not elapsed hours, so midnight and DST behave naturally.
export function daysUntil(dueAt: string, now: number): number {
  const date = new Date(dueAt),
    today = new Date(now);
  const ordinal = (d: Date) =>
    Date.UTC(d.getFullYear(), d.getMonth(), d.getDate());
  return Math.round((ordinal(date) - ordinal(today)) / 86400000);
}

export function billEvents(
  bills: DashboardState['finance']['bills'],
  now: number,
): AttentionEvent[] {
  const start = new Date(now),
    end = new Date(now);
  start.setHours(0, 0, 0, 0);
  end.setHours(24, 0, 0, 0);
  return bills.flatMap((bill) => {
    const days = daysUntil(bill.dueAt, now);
    if (!Number.isFinite(days) || days > 7) return [];
    const kind =
      days < 0
        ? 'overdue'
        : days === 0
          ? 'today'
          : days === 1
            ? 'tomorrow'
            : days <= 3
              ? 'soon'
              : 'week';
    return [
      {
        id: `bill:${bill.name}:${bill.dueAt}`,
        widget: 'finance',
        kind,
        startsAt: start.toISOString(),
        expiresAt: end.toISOString(),
        decay: false,
      },
    ];
  });
}
