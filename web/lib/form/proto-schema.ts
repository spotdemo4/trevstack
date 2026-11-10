import {
  create,
  type DescField,
  type DescMessage,
  type MessageInitShape,
  type MessageShape,
  type MessageValidType,
  ScalarType,
} from "@bufbuild/protobuf";
import { createValidator } from "@bufbuild/protovalidate";
import type { StandardSchemaV1 } from "effect/StandardSchema";

const validator = createValidator();

/** The 32-bit integer types, which a form's numbers must fit, and their ranges. */
const INTEGER_RANGES = new Map<ScalarType, readonly [number, number]>([
  [ScalarType.INT32, [-(2 ** 31), 2 ** 31 - 1]],
  [ScalarType.SINT32, [-(2 ** 31), 2 ** 31 - 1]],
  [ScalarType.SFIXED32, [-(2 ** 31), 2 ** 31 - 1]],
  [ScalarType.UINT32, [0, 2 ** 32 - 1]],
  [ScalarType.FIXED32, [0, 2 ** 32 - 1]],
]);

/**
 * A Standard Schema for a form that fills in a request message of type `schema`. It checks the
 * values against the `buf.validate` rules in the message's proto, which the client and the server
 * check the request against too, so the form reports the mistakes they would reject. Each message
 * starts with the field it is about, such as "Name must be at most 50 characters". Rules
 * of the whole message, such as one comparing two fields, have no field and report on the form.
 *
 * protovalidate expects values of the field's type, so the schema first checks that numbers fit
 * their integer field: a number input can hold 1.5 or -1.
 */
export function protoSchema<Desc extends DescMessage>(
  schema: Desc,
): StandardSchemaV1<MessageShape<Desc>, MessageValidType<Desc>> {
  return {
    "~standard": {
      version: 1,
      vendor: "protovalidate",
      validate(value) {
        const values: Record<string, unknown> = { ...(value as object) };
        const issues: StandardSchemaV1.Issue[] = [];
        for (const field of schema.fields) {
          const problem = typeProblem(field, values[field.localName]);
          if (problem !== undefined) {
            issues.push({ message: `${label(field)} ${problem}`, path: [field.localName] });
            delete values[field.localName];
          }
        }

        const result = validator.validate(schema, create(schema, values as MessageInitShape<Desc>));
        if (result.kind === "valid") {
          return issues.length > 0 ? { issues } : { value: result.message };
        }
        if (result.kind === "error") {
          return { issues: [...issues, { message: result.error.message }] };
        }
        for (const violation of result.violations) {
          const [field] = violation.field;
          if (field?.kind !== "field") {
            issues.push({ message: capitalize(violation.message) });
          } else if (!issues.some((issue) => issue.path?.[0] === field.localName)) {
            // A field with a type problem was left out of the message, so it is not also required.
            const message = violation.message.replace(/^value /, "");
            issues.push({ message: `${label(field)} ${message}`, path: [field.localName] });
          }
        }
        return { issues };
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

/** The name of `field` for messages, such as "Bucket count" for `bucket_count`. */
function label(field: DescField): string {
  return capitalize(field.name.replaceAll("_", " "));
}

function capitalize(text: string): string {
  return text.charAt(0).toUpperCase() + text.slice(1);
}
