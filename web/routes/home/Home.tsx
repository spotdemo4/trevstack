import { AddRequestSchema } from "$connect/number/v1/add_pb";
import { Card } from "$lib/card";
import { NumberClient } from "$lib/connect";
import { useForm } from "$lib/form/hook";
import { toaster } from "$lib/toast";
import { create } from "@bufbuild/protobuf";
import { createStandardSchema } from "@bufbuild/protovalidate";
import { Effect } from "effect";
import type { Component } from "solid-js";

const Home: Component = () => {
  const form = useForm(() => ({
    defaultValues: create(AddRequestSchema),
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
        <form.AppForm>
          <form.Form class="max-w-sm flex-col">
            <form.AppField name="name">{(field) => <field.TextField label="Name" />}</form.AppField>
            <form.AppField name="number">
              {(field) => <field.NumberField label="Number" />}
            </form.AppField>
            <form.SubmitButton />
          </form.Form>
        </form.AppForm>
      </Card>
    </div>
  );
};

export default Home;
