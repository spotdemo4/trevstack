import { SelectInput } from "$lib/input";
import { Field } from "@ark-ui/solid/field";
import { createMemo, For, type Component } from "solid-js";

import { useFieldContext } from "./context";

type SelectOption = {
  label: string;
  value: string;
  disabled?: boolean;
};

type SelectFieldProps = {
  class?: string;
  label?: string;
  options: SelectOption[];
  placeholder?: string;
};

export const SelectField: Component<SelectFieldProps> = (props) => {
  const field = useFieldContext<string>();
  const name = field().name;
  const errors = createMemo(() => [
    ...new Set(field().state.meta.errors.map((err) => err.message as string)),
  ]);

  const value = createMemo(() => (field().state.value ? [field().state.value] : []));

  return (
    <Field.Root
      invalid={!(field().state.meta.isValid || !field().state.meta.isBlurred)}
      class="flex flex-col gap-1.5"
    >
      <SelectInput
        name={name}
        label={props.label}
        class={props.class}
        items={props.options}
        placeholder={props.placeholder}
        value={value()}
        onBlur={field().handleBlur}
        onValueChange={(details) => {
          const first = details.value[0];
          if (typeof first === "string") {
            field().handleChange(first);
          } else if (!first) {
            field().form.resetField(name);
          }
        }}
      />
      <For each={errors()}>
        {(err) => <Field.ErrorText class="text-xs text-ctp-red">{err}</Field.ErrorText>}
      </For>
    </Field.Root>
  );
};
