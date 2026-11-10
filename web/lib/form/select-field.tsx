import { SelectInput } from "$lib/input";
import type { JSX } from "@solidjs/web";
import { For, Show, createUniqueId } from "solid-js";

import type { FieldController } from "./hook";

type SelectFieldProps<T> = {
  class?: string;
  field: FieldController<T | undefined>;
  label?: string;
  items: { label: string; value: T }[];
  /** The label of the first option, which leaves the field unset. */
  emptyLabel?: string;
};

/** A select of `items`, any value, such as an enum's, plus an option that leaves the field unset. */
export function SelectField<T>(props: SelectFieldProps<T>): JSX.Element {
  const id = createUniqueId();
  const errorId = `${id}-errors`;
  const selected = () => props.items.findIndex((item) => item.value === props.field.value());
  const options = () => [
    { label: props.emptyLabel ?? "Any", value: "" },
    ...props.items.map((item, index) => ({ label: item.label, value: String(index) })),
  ];

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
      <SelectInput
        id={id}
        name={props.field.name}
        class={props.class}
        invalid={props.field.invalid()}
        aria-describedby={props.field.invalid() ? errorId : undefined}
        items={options()}
        value={[selected() === -1 ? "" : String(selected())]}
        onBlur={props.field.handleBlur}
        onValueChange={(details) => {
          const index = details.value[0];
          props.field.handleChange(index ? props.items[Number(index)]?.value : undefined);
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
}
