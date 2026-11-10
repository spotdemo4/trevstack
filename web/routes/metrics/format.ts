/** Locale-formatted number with at most two fraction digits. */
export const formatNumber = (n: bigint | number): string =>
  typeof n === "bigint"
    ? n.toLocaleString()
    : n.toLocaleString(undefined, { maximumFractionDigits: 2 });

type Counts = { count: number; streams: number; users?: number; distance?: number };

/** Multi-line tooltip text for an aggregate of positions. */
export const formatCounts = (title: string, counts: Counts): string =>
  [
    title,
    `Positions: ${formatNumber(counts.count)}`,
    `Streams: ${formatNumber(counts.streams)}`,
    ...(counts.users === undefined ? [] : [`Users: ${formatNumber(counts.users)}`]),
    ...(counts.distance === undefined ? [] : [`Distance: ${formatNumber(counts.distance)}`]),
  ].join("\n");
