import { TextInput } from "$lib/input";
import { Field } from "@ark-ui/solid/field";
import { createMemo, For, Show, type Component } from "solid-js";

import { useFieldContext } from "./context";

type TextFieldProps = {
  label?: string;
};

export const TextField: Component<TextFieldProps> = (props) => {
  const field = useFieldContext<string>();
  const name = field().name;
  const errors = createMemo(() => [
    ...new Set(field().state.meta.errors.map((err) => err.message as string)),
  ]);

  return (
    <Field.Root
      invalid={!(field().state.meta.isValid || !field().state.meta.isBlurred)}
      class="flex flex-col gap-1.5"
    >
      <Show when={props.label}>
        <Field.Label class="text-sm font-medium text-ctp-subtext1 data-invalid:text-ctp-red">
          {props.label}
        </Field.Label>
      </Show>
      <TextInput
        name={name}
        value={field().state.value ?? ""}
        onInput={(e) => field().handleChange(e.target.value)}
        onBlur={field().handleBlur}
      />
      <For each={errors()}>
        {(err) => <Field.ErrorText class="text-xs text-ctp-red">{err}</Field.ErrorText>}
      </For>
    </Field.Root>
  );
};
