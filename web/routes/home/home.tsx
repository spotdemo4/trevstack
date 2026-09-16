import { AddRequestSchema } from "$connect/number/v1/add_pb";
import { Card } from "$lib/card";
import { NumberClient } from "$lib/connect";
import { Form } from "$lib/form/form";
import { useForm } from "$lib/form/hook";
import { NumberField } from "$lib/form/number-field";
import { SubmitButton } from "$lib/form/submit-button";
import { TextField } from "$lib/form/text-field";
import { toaster } from "$lib/toast";
import { create } from "@bufbuild/protobuf";
import { createStandardSchema } from "@bufbuild/protovalidate";
import { Effect } from "effect";
import type { Component } from "solid-js";

export const Home: Component = () => {
  const form = useForm(() => ({
    defaultValues: { ...create(AddRequestSchema) },
    validators: {
      onMount: createStandardSchema(AddRequestSchema),
      onChange: createStandardSchema(AddRequestSchema),
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
              description: err.message,
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
