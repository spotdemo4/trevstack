import { SignupRequestSchema } from "$gen/auth/v1/signup_pb";
import { AddRequestSchema } from "$gen/number/v1/add_pb";
import { ListRequestSchema } from "$gen/number/v1/list_pb";
import { TimeSeriesRequestSchema } from "$gen/number/v1/metrics_pb";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { expect, test } from "vitest";

import { protoSchema } from "./proto-schema.ts";

function validate(schema: Parameters<typeof protoSchema>[0], value: object) {
  return protoSchema(schema)["~standard"].validate(value);
}

test("valid values become the request message", () => {
  expect(validate(AddRequestSchema, { name: "foo", number: 10 })).toEqual({
    value: expect.objectContaining({ $typeName: "number.v1.AddRequest", name: "foo", number: 10 }),
  });
});

test("violations name their field", () => {
  expect(validate(AddRequestSchema, { name: "a".repeat(51) })).toEqual({
    issues: [
      { message: "Name must be at most 50 characters", path: ["name"] },
      { message: "Number is required", path: ["number"] },
    ],
  });
  expect(validate(TimeSeriesRequestSchema, {})).toEqual({
    issues: [{ message: "Interval is required", path: ["interval"] }],
  });
});

test("custom rules keep their messages", () => {
  expect(validate(SignupRequestSchema, { username: "-alice", password: "password" })).toEqual({
    issues: [
      {
        message:
          "Username must start with a letter or digit and contain only letters, digits, underscores, and hyphens",
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
    issues: [{ message: "Password must be at most 72 bytes", path: ["password"] }],
  });
});

test("numbers must fit their integer field before the rules apply", () => {
  expect(validate(AddRequestSchema, { name: "foo", number: 1.5 })).toEqual({
    issues: [{ message: "Number must be a whole number", path: ["number"] }],
  });
  expect(validate(ListRequestSchema, { minimum: -1 })).toEqual({
    issues: [{ message: "Minimum must not be negative", path: ["minimum"] }],
  });
});

test("rules of the whole message have no field", () => {
  const start = timestampFromDate(new Date("2026-02-01T00:00:00Z"));
  const end = timestampFromDate(new Date("2026-01-01T00:00:00Z"));
  expect(validate(ListRequestSchema, { minimum: 5, maximum: 5, start, end })).toEqual({
    issues: [
      { message: "Minimum must be less than maximum" },
      { message: "End must be after start" },
    ],
  });
});
