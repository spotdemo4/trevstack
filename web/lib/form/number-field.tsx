import { NumberInput } from "$lib/input";
import { For, Show, type Component, createUniqueId } from "solid-js";

import type { FieldController } from "./hook";

type NumberFieldProps = {
  class?: string;
  field: FieldController<number | undefined>;
  label?: string;
};

export const NumberField: Component<NumberFieldProps> = (props) => {
  const id = createUniqueId();
  const errorId = `${id}-errors`;

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
      <NumberInput
        id={id}
        name={props.field.name}
        invalid={props.field.invalid()}
        aria-describedby={props.field.invalid() ? errorId : undefined}
        class={props.class}
        value={props.field.value() != undefined ? String(props.field.value()) : ""}
        onBlur={props.field.handleBlur}
        onValueChange={(details) => {
          props.field.handleChange(
            Number.isNaN(details.valueAsNumber) ? undefined : details.valueAsNumber,
          );
        }}
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
