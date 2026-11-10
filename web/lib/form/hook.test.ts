import { createRoot } from "solid-js";
import { expect, test } from "vitest";

import { useForm } from "./hook.tsx";

test("changes to several fields in one event are all kept and submitted", async () => {
  const submitted: object[] = [];
  const form = createRoot(() =>
    useForm({
      defaultValues: {} as { x?: number; y?: number },
      onSubmit: ({ value }) => {
        submitted.push(value);
      },
    }),
  );

  form.field("x").handleChange(0.25);
  form.field("y").handleChange(0.75);
  await form.handleSubmit();

  expect(submitted).toEqual([{ x: 0.25, y: 0.75 }]);
});
