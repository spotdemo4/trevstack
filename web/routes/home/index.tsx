import { AddRequestSchema } from "$gen/number/v1/add_pb";
import { Card } from "$lib/card";
import { errorDescription } from "$lib/effect";
import { Form } from "$lib/form/form";
import { useForm } from "$lib/form/hook";
import { NumberField } from "$lib/form/number-field";
import { protoSchema } from "$lib/form/proto-schema";
import { SubmitButton } from "$lib/form/submit-button";
import { TextField } from "$lib/form/text-field";
import { NumberClient } from "$lib/rpc";
import { toaster } from "$lib/toast";
import { Effect } from "effect";
import type { Component } from "solid-js";

const Home: Component = () => {
  const form = useForm(protoSchema(AddRequestSchema), () => ({
    validation: {
      onMount: true,
      onChange: true,
    },
    onSubmit: ({ value }) =>
      NumberClient.add(value).pipe(
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

export default Home;
