export function fittingRowCount(
  availableHeight: number,
  rowHeights: readonly number[],
  gap: number,
): number {
  let usedHeight = 0;
  let count = 0;

  for (const height of rowHeights) {
    const nextHeight = usedHeight + (count > 0 ? gap : 0) + height;
    if (height <= 0 || nextHeight > availableHeight) break;
    usedHeight = nextHeight;
    count += 1;
  }

  return count;
}
