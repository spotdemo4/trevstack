import { NumberInput } from "$lib/input";
import { Field } from "@ark-ui/solid/field";
import { createMemo, For, type Component } from "solid-js";

import { useFieldContext } from "./context";

type NumberFieldProps = {
  label?: string;
};

export const NumberField: Component<NumberFieldProps> = (props) => {
  const field = useFieldContext<number>();
  const name = field().name;
  const errors = createMemo(() => [
    ...new Set(field().state.meta.errors.map((err) => err.message as string)),
  ]);

  return (
    <Field.Root
      invalid={!(field().state.meta.isValid || !field().state.meta.isBlurred)}
      class="flex flex-col gap-1.5"
    >
      <NumberInput
        name={name}
        label={props.label}
        value={field().state.value != undefined ? String(field().state.value) : ""}
        onBlur={field().handleBlur}
        onValueChange={(c) => {
          // remove undefined numbers from form state
          if (Number.isNaN(c.valueAsNumber) && field().state.value !== undefined) {
            field().form.deleteField(name);
          }

          // only handle change if the value is a valid number
          if (!Number.isNaN(c.valueAsNumber)) {
            field().handleChange(c.valueAsNumber);
          } else {
            // still validate
            void field().form.validateField(name, "change");
          }
        }}
      />
      <For each={errors()}>
        {(err) => <Field.ErrorText class="text-xs text-ctp-red">{err}</Field.ErrorText>}
      </For>
    </Field.Root>
  );
};
