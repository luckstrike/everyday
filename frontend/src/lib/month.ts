export function monthGrid(year: number, month: number) {
  const first = new Date(year, month, 1);
  const count = new Date(year, month + 1, 0).getDate();
  const offset = first.getDay();
  const cells = Array.from(
    { length: Math.ceil((offset + count) / 7) * 7 },
    (_, i) => (i >= offset && i < offset + count ? i - offset + 1 : null),
  );
  const weeks: (number | null)[][] = [];
  for (let i = 0; i < cells.length; i += 7) weeks.push(cells.slice(i, i + 7));
  return {
    year,
    month,
    label: first.toLocaleDateString('en-US', {
      month: 'long',
      year: 'numeric',
    }),
    weeks,
  };
}

// The row of the month grid that contains the given day, for compact displays.
export function weekContaining(
  weeks: (number | null)[][],
  day: number,
): (number | null)[][] {
  return weeks.filter((week) => week.includes(day));
}
