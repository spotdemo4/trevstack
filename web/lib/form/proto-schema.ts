import {
  type DescField,
  type DescMessage,
  type MessageShape,
  type MessageValidType,
  ScalarType,
} from "@bufbuild/protobuf";
import { createStandardSchema } from "@bufbuild/protovalidate";
import type { StandardSchemaV1 } from "effect/StandardSchema";

/**
 * The integer types, which a form's numbers must fit, and their ranges. A number input only holds
 * safe integers exactly, which bounds the 64-bit types.
 */
const INTEGER_RANGES = new Map<ScalarType, readonly [number, number]>([
  [ScalarType.INT32, [-(2 ** 31), 2 ** 31 - 1]],
  [ScalarType.SINT32, [-(2 ** 31), 2 ** 31 - 1]],
  [ScalarType.SFIXED32, [-(2 ** 31), 2 ** 31 - 1]],
  [ScalarType.UINT32, [0, 2 ** 32 - 1]],
  [ScalarType.FIXED32, [0, 2 ** 32 - 1]],
  [ScalarType.INT64, [Number.MIN_SAFE_INTEGER, Number.MAX_SAFE_INTEGER]],
  [ScalarType.SINT64, [Number.MIN_SAFE_INTEGER, Number.MAX_SAFE_INTEGER]],
  [ScalarType.SFIXED64, [Number.MIN_SAFE_INTEGER, Number.MAX_SAFE_INTEGER]],
  [ScalarType.UINT64, [0, Number.MAX_SAFE_INTEGER]],
  [ScalarType.FIXED64, [0, Number.MAX_SAFE_INTEGER]],
]);

/** The 64-bit integer types, whose fields hold a bigint rather than a form's number. */
const BIGINT_TYPES = new Set([
  ScalarType.INT64,
  ScalarType.SINT64,
  ScalarType.SFIXED64,
  ScalarType.UINT64,
  ScalarType.FIXED64,
]);

/**
 * A Standard Schema for a form that fills in a request message of type `schema`. It checks the
 * values against the `buf.validate` rules in the message's proto, which the client and the server
 * check the request against too, so the form reports the mistakes they would reject.
 *
 * It is protovalidate's Standard Schema, which only checks the rules, plus a check that numbers
 * fit their integer field: a number input can hold 1.5 or -1, which protovalidate accepts and the
 * request then fails to encode. Numbers for 64-bit integer fields become bigints. Use `createStandardSchema` directly once
 * https://github.com/bufbuild/protovalidate-es/issues/218 is fixed.
 */
export function protoSchema<Desc extends DescMessage>(
  schema: Desc,
): StandardSchemaV1<MessageShape<Desc>, MessageValidType<Desc>> {
  const standard = createStandardSchema(schema)["~standard"];
  return {
    "~standard": {
      ...standard,
      validate(value) {
        if (typeof value !== "object" || value === null) return standard.validate(value);
        const values: Record<string, unknown> = { ...value };
        const issues: StandardSchemaV1.Issue[] = [];
        for (const field of schema.fields) {
          const fieldValue = values[field.localName];
          const message = typeProblem(field, fieldValue);
          if (message !== undefined) {
            issues.push({ message, path: [field.localName] });
            delete values[field.localName];
          } else if (
            field.fieldKind === "scalar" &&
            BIGINT_TYPES.has(field.scalar) &&
            typeof fieldValue === "number"
          ) {
            values[field.localName] = BigInt(fieldValue);
          }
        }
        if (issues.length === 0) return standard.validate(values);

        // A field with a type problem was left out of the message, so its rules, such as being
        // required, do not also report on it.
        const fields = new Set(issues.map((issue) => issue.path?.[0]));
        const merge = (result: StandardSchemaV1.Result<unknown>) => ({
          issues: [
            ...issues,
            ...(result.issues ?? []).filter((issue) => !fields.has(issue.path?.[0])),
          ],
        });
        const result = standard.validate(values);
        return result instanceof Promise ? result.then(merge) : merge(result);
      },
    },
  };
}

/** Why `value` does not fit the integer `field`, or undefined when it does or is not a number. */
function typeProblem(field: DescField, value: unknown): string | undefined {
  if (field.fieldKind !== "scalar" || typeof value !== "number") return undefined;
  const range = INTEGER_RANGES.get(field.scalar);
  if (range === undefined) return undefined;
  const [minimum, maximum] = range;
  if (!Number.isInteger(value)) return "must be a whole number";
  if (value < minimum)
    return minimum === 0 ? "must not be negative" : `must be at least ${minimum}`;
  if (value > maximum) return `must be at most ${maximum}`;
  return undefined;
}
