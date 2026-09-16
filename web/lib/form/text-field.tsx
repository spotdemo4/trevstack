import { TextInput } from "$lib/input";
import type { JSX } from "@solidjs/web";
import { For, Show, type Component, createUniqueId, omit } from "solid-js";

import type { FieldController } from "./hook";

type TextFieldProps = Omit<
  JSX.InputHTMLAttributes<HTMLInputElement>,
  "class" | "id" | "name" | "value" | "onInput" | "onBlur"
> & {
  class?: string;
  field: FieldController<string | undefined>;
  label?: string;
};

export const TextField: Component<TextFieldProps> = (props) => {
  const id = createUniqueId();
  const errorId = `${id}-errors`;
  const inputProps = omit(props, "class", "field", "label");

  return (
    <div class="flex flex-col gap-1.5">
      <Show when={props.label}>
        <label
          for={id}
          class="text-sm font-medium text-ctp-subtext1 aria-invalid:text-ctp-red"
          aria-invalid={props.field.invalid() ? "true" : undefined}
        >
          {props.label}
        </label>
      </Show>
      <TextInput
        {...inputProps}
        id={id}
        name={props.field.name}
        value={props.field.value() ?? ""}
        aria-invalid={props.field.invalid() ? "true" : undefined}
        aria-describedby={props.field.invalid() ? errorId : undefined}
        onInput={(event) => props.field.handleChange(event.currentTarget.value)}
        onBlur={props.field.handleBlur}
        class={props.class}
      />
      <Show when={props.field.invalid()}>
        <div id={errorId}>
          <For each={props.field.errors()}>
            {(error) => <span class="block text-xs text-ctp-red">{error}</span>}
          </For>
        </div>
      </Show>
    </div>
  );
};
