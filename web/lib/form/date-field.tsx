import { DateInput } from "$lib/input";
import { create, type MessageInitShape } from "@bufbuild/protobuf";
import {
  TimestampSchema,
  type Timestamp,
  timestampDate,
  timestampFromDate,
} from "@bufbuild/protobuf/wkt";
import { For, Show, type Component, createUniqueId } from "solid-js";

import type { FieldController } from "./hook";

type TimestampInput = MessageInitShape<typeof TimestampSchema>;

function timestampToDateString(timestamp: TimestampInput | undefined): string {
  return timestamp
    ? timestampDate(create(TimestampSchema, timestamp)).toISOString().slice(0, 10)
    : "";
}

function dateStringToTimestamp(value: string): Timestamp | undefined {
  if (!value) return undefined;
  const date = new Date(`${value}T00:00:00Z`);
  return Number.isNaN(date.getTime()) ? undefined : timestampFromDate(date);
}

type DateFieldProps = {
  class?: string;
  field: FieldController<Timestamp | undefined> | FieldController<TimestampInput | undefined>;
  label?: string;
};

export const DateField: Component<DateFieldProps> = (props) => {
  const id = createUniqueId();
  const errorId = `${id}-errors`;

  return (
    <div class="flex min-w-42 flex-col gap-1.5">
      <Show when={props.label}>
        <label
          for={id}
          class="text-sm font-medium text-ctp-subtext1 aria-invalid:text-ctp-red"
          aria-invalid={props.field.invalid() ? "true" : undefined}
        >
          {props.label}
        </label>
      </Show>
      <DateInput
        id={id}
        name={props.field.name}
        invalid={props.field.invalid()}
        aria-describedby={props.field.invalid() ? errorId : undefined}
        class={props.class}
        value={timestampToDateString(props.field.value())}
        onBlur={props.field.handleBlur}
        onValueChange={(value) => props.field.handleChange(dateStringToTimestamp(value))}
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
