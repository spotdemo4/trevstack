import { SignupRequestSchema } from "$gen/auth/v1/signup_pb";
import { ListRequestSchema } from "$gen/cursor/v1/list_pb";
import { HeatmapRequestSchema, TimeSeriesRequestSchema } from "$gen/cursor/v1/metrics_pb";
import { SendRequestSchema } from "$gen/cursor/v1/send_pb";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { expect, test } from "vitest";

import { protoSchema } from "./proto-schema.ts";

function validate(schema: Parameters<typeof protoSchema>[0], value: object) {
  return protoSchema(schema)["~standard"].validate(value);
}

test("valid values become the request message", () => {
  expect(validate(SendRequestSchema, { x: 0.25, y: 1 })).toEqual({
    value: expect.objectContaining({ $typeName: "cursor.v1.SendRequest", x: 0.25, y: 1 }),
  });
});

test("violations are reported at their field", () => {
  expect(validate(ListRequestSchema, { username: "a".repeat(51) })).toEqual({
    issues: [{ message: "must be at most 50 characters", path: ["username"] }],
  });
  expect(validate(SendRequestSchema, { x: 1.5 })).toEqual({
    issues: [
      { message: "must be greater than or equal to 0 and less than or equal to 1", path: ["x"] },
      { message: "value is required", path: ["y"] },
    ],
  });
  expect(validate(TimeSeriesRequestSchema, {})).toEqual({
    issues: [{ message: "value is required", path: ["interval"] }],
  });
});

test("custom rules keep their messages", () => {
  expect(validate(SignupRequestSchema, { username: "-alice", password: "password" })).toEqual({
    issues: [
      {
        message:
          "value must start with a letter or digit and contain only letters, digits, underscores, and hyphens",
        path: ["username"],
      },
    ],
  });
});

test("passwords are limited in bytes, not characters", () => {
  expect(validate(SignupRequestSchema, { username: "alice", password: "é".repeat(36) })).toEqual({
    value: expect.anything(),
  });
  expect(validate(SignupRequestSchema, { username: "alice", password: "é".repeat(37) })).toEqual({
    issues: [{ message: "must be at most 72 bytes", path: ["password"] }],
  });
});

test("numbers must fit their integer field before the rules apply", () => {
  expect(validate(HeatmapRequestSchema, { columns: 1.5, rows: 2 })).toEqual({
    issues: [{ message: "must be a whole number", path: ["columns"] }],
  });
  expect(validate(HeatmapRequestSchema, { columns: -1, rows: 2 })).toEqual({
    issues: [{ message: "must not be negative", path: ["columns"] }],
  });
  expect(validate(ListRequestSchema, { streamId: 1.5 })).toEqual({
    issues: [{ message: "must be a whole number", path: ["streamId"] }],
  });
});

test("numbers for 64-bit integer fields become bigints", () => {
  expect(validate(ListRequestSchema, { streamId: 7 })).toEqual({
    value: expect.objectContaining({ streamId: 7n }),
  });
  expect(validate(ListRequestSchema, { streamId: 0 })).toEqual({
    issues: [{ message: "must be greater than or equal to 1", path: ["streamId"] }],
  });
});

test("rules of the whole message have no field", () => {
  const start = timestampFromDate(new Date("2026-02-01T00:00:00Z"));
  const end = timestampFromDate(new Date("2026-01-01T00:00:00Z"));
  expect(validate(ListRequestSchema, { minX: 0.5, maxX: 0.5, start, end })).toEqual({
    issues: [{ message: "min x must be less than max x" }, { message: "end must be after start" }],
  });
});
