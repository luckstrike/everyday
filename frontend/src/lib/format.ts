export const money = (value: number) =>
  new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: value % 1 ? 2 : 0,
  }).format(value);
export const time = (value: string | number) =>
  new Date(value).toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
  });
export function day(value: string, now: number): string {
  const target = new Date(value),
    today = new Date(now);
  const dateUTC = (d: Date) =>
    Date.UTC(d.getFullYear(), d.getMonth(), d.getDate());
  const days = Math.round((dateUTC(target) - dateUTC(today)) / 86400000);
  return days === 0
    ? 'Today'
    : days === 1
      ? 'Tomorrow'
      : target.toLocaleDateString('en-US', {
          weekday: 'short',
          month: 'short',
          day: 'numeric',
        });
}
// When an event happens, e.g. "Tomorrow · 6:00 PM" or "Today · All day". An all-day
// event that is already under way (a multi-day trip) reads as today, not as its first day.
export function when(
  event: { startsAt: string; allDay?: boolean },
  now: number,
): string {
  if (!event.allDay)
    return `${day(event.startsAt, now)} · ${time(event.startsAt)}`;
  const label =
    Date.parse(event.startsAt) <= now ? 'Today' : day(event.startsAt, now);
  return `${label} · All day`;
}
