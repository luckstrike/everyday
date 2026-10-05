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
