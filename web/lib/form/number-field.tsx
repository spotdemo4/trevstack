import { NumberInput } from "$lib/input";
import { For, Show, type Component, createUniqueId } from "solid-js";

import type { FieldController } from "./hook";

type NumberFieldProps = {
  class?: string;
  /** A field of a 64-bit integer holds a bigint, which `protoSchema` makes of the number. */
  field: FieldController<number | undefined> | FieldController<bigint | undefined>;
  label?: string;
  min?: number;
  max?: number;
  /** The granularity of the value; a fractional step lets the input hold decimals. */
  step?: number;
};

export const NumberField: Component<NumberFieldProps> = (props) => {
  const id = createUniqueId();
  const errorId = `${id}-errors`;

  return (
    <div class="flex flex-col gap-1.5">
      <Show when={props.label}>
        <label
          for={id}
          class="text-sm font-medium text-foreground-muted aria-invalid:text-danger"
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
        min={props.min}
        max={props.max}
        step={props.step}
        inputmode={
          props.step !== undefined && !Number.isInteger(props.step) ? "decimal" : "numeric"
        }
        value={props.field.value() != undefined ? String(props.field.value()) : ""}
        onBlur={props.field.handleBlur}
        onValueChange={(details) => {
          const handleChange = props.field.handleChange as (value: number | undefined) => void;
          handleChange(Number.isNaN(details.valueAsNumber) ? undefined : details.valueAsNumber);
        }}
      />
      <Show when={props.field.invalid()}>
        <div id={errorId}>
          <For each={props.field.errors()}>
            {(error) => <span class="block text-xs text-danger">{error}</span>}
          </For>
        </div>
      </Show>
    </div>
  );
};
