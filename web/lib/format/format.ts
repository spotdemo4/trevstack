import { type Timestamp, timestampDate } from "@bufbuild/protobuf/wkt";

const timestampFormat = new Intl.DateTimeFormat(undefined, {
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
  second: "2-digit",
  fractionalSecondDigits: 3,
});

/** A local date and time to the millisecond. */
export const formatTimestamp = (timestamp: Timestamp | undefined): string =>
  timestamp ? timestampFormat.format(timestampDate(timestamp)) : "";

/** A position's coordinate, to three decimal places. */
export const formatCoordinate = (value: number): string => value.toFixed(3);

/** The time between two timestamps, such as `850 ms`, `12.3 s`, or `4m 05s`. */
export const formatDuration = (
  start: Timestamp | undefined,
  end: Timestamp | undefined,
): string => {
  if (!start || !end) return "";
  const ms = timestampDate(end).getTime() - timestampDate(start).getTime();
  if (ms < 1000) return `${ms} ms`;
  if (ms < 60_000) return `${(ms / 1000).toFixed(1)} s`;
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ${String(seconds % 60).padStart(2, "0")}s`;
  return `${Math.floor(minutes / 60)}h ${String(minutes % 60).padStart(2, "0")}m`;
};
