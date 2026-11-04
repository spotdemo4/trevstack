import { Card } from "$lib/card";
import { errorDescription } from "$lib/effect";
import { Form } from "$lib/form/form";
import { useForm } from "$lib/form/hook";
import { NumberField } from "$lib/form/number-field";
import { SubmitButton } from "$lib/form/submit-button";
import { TextField } from "$lib/form/text-field";
import { NumberClient } from "$lib/rpc";
import { toaster } from "$lib/toast";
import { Effect, Schema } from "effect";
import type { Component } from "solid-js";

// The same rules as the server's, so most mistakes are caught before a request.
const addSchema = Schema.toStandardSchemaV1(
  Schema.Struct({
    name: Schema.String.check(
      Schema.isMinLength(1, { message: "Name is required" }),
      Schema.isMaxLength(50, { message: "Name must be at most 50 characters" }),
    ).annotateKey({ messageMissingKey: "Name is required" }),
    number: Schema.Number.check(
      Schema.isInt({ message: "Number must be a whole number" }),
      Schema.isBetween(
        { minimum: 1, maximum: 1_000_000 },
        { message: "Number must be between 1 and 1000000" },
      ),
    ).annotateKey({ messageMissingKey: "Number is required" }),
  }),
);

export const Home: Component = () => {
  const form = useForm(addSchema, () => ({
    validation: {
      onMount: true,
      onChange: true,
    },
    onSubmit: ({ value }) =>
      NumberClient.add({ name: value.name, number: value.number }).pipe(
        Effect.match({
          onSuccess: (response) => {
            toaster.success({
              title: "Number added",
              description: `new sum: ${response.sum}`,
            });
            form.reset();
          },
          onFailure: (err) => {
            toaster.error({
              title: "Failed to add number",
              description: errorDescription(err),
            });
          },
        }),
        Effect.runPromise,
      ),
  }));

  return (
    <div class="flex h-body flex-col items-center justify-center gap-4">
      <h1 class="text-2xl font-bold">Add Numbers</h1>
      <Card class="w-full max-w-sm">
        <Form form={form} class="max-w-sm flex-col">
          <TextField field={form.field("name")} label="Name" />
          <NumberField field={form.field("number")} label="Number" />
          <SubmitButton form={form} />
        </Form>
      </Card>
    </div>
  );
};
