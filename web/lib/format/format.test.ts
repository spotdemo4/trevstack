import { timestampFromMs } from "@bufbuild/protobuf/wkt";
import { expect, test } from "vitest";

import { formatCoordinate, formatDuration } from "./format.ts";

test("durations use a unit that suits their length", () => {
  const at = (ms: number) => timestampFromMs(ms);
  expect(formatDuration(at(0), at(850))).toBe("850 ms");
  expect(formatDuration(at(0), at(12_340))).toBe("12.3 s");
  expect(formatDuration(at(0), at(245_000))).toBe("4m 05s");
  expect(formatDuration(at(0), at(3_900_000))).toBe("1h 05m");
  expect(formatDuration(undefined, at(1))).toBe("");
});

test("coordinates have three decimal places", () => {
  expect(formatCoordinate(0.5)).toBe("0.500");
  expect(formatCoordinate(1 / 3)).toBe("0.333");
});
