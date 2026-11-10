import { format } from "d3-format";

/** Locale-formatted number with at most two fraction digits. */
export const formatNumber = (n: bigint | number): string =>
  typeof n === "bigint"
    ? n.toLocaleString()
    : n.toLocaleString(undefined, { maximumFractionDigits: 2 });

const compact = (n: number): string =>
  format("~s")(Math.abs(n) < 1000 ? n : Math.round(n / 1000) * 1000).replace("G", "B");

/** Short range label such as `1k-2k`, for axis ticks. */
export const formatRangeShort = (lower: number, upper: number): string =>
  lower === upper ? compact(lower) : `${compact(lower)}-${compact(upper)}`;

/** Full range label such as `1000-2000`. */
export const formatRange = (lower: number, upper: number): string =>
  lower === upper ? `${lower}` : `${lower}-${upper}`;

type Stats = { sum: number; count: number; average: number };

/** Multi-line tooltip text for an aggregate of numbers. */
export const formatStats = (title: string, stats: Stats): string =>
  [
    title,
    `Total value: ${formatNumber(stats.sum)}`,
    `Count: ${formatNumber(stats.count)}`,
    `Average: ${formatNumber(stats.average)}`,
  ].join("\n");
