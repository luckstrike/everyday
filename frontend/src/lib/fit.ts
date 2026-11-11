export function fittingItemCount(
  availableSpace: number,
  itemSizes: readonly number[],
  gap: number,
): number {
  let usedSpace = 0;
  let count = 0;

  for (const size of itemSizes) {
    const nextSpace = usedSpace + (count > 0 ? gap : 0) + size;
    if (size <= 0 || nextSpace > availableSpace) break;
    usedSpace = nextSpace;
    count += 1;
  }

  return count;
}
