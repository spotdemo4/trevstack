import { DateInput } from "$lib/input";
import { Field } from "@ark-ui/solid/field";
import { type Timestamp, timestampDate, timestampFromDate } from "@bufbuild/protobuf/wkt";
import { CalendarDate, parseDate, fromDate, toCalendarDate } from "@internationalized/date";
import { createMemo, For, type Component } from "solid-js";

import { useFieldContext } from "./context";

function timestampToDateString(ts: Timestamp | undefined): string {
  if (!ts) return "";
  const jsDate = timestampDate(ts);
  const zonedDateTime = fromDate(jsDate, "UTC");
  const calendarDate = toCalendarDate(zonedDateTime);
  return calendarDate.toString();
}

function calendarDateToTimestamp(value: CalendarDate): Timestamp {
  const jsDate = value.toDate("UTC");
  return timestampFromDate(jsDate);
}

type DateFieldProps = {
  class?: string;
  label?: string;
};

export const DateField: Component<DateFieldProps> = (props) => {
  const field = useFieldContext<Timestamp | undefined>();
  const name = field().name;
  const errors = createMemo(() => [
    ...new Set(field().state.meta.errors.map((err) => err.message as string)),
  ]);

  const value = createMemo(() => {
    const str = timestampToDateString(field().state.value);
    return str ? [parseDate(str)] : [];
  });

  return (
    <Field.Root
      invalid={!(field().state.meta.isValid || !field().state.meta.isBlurred)}
      class="flex flex-col gap-1.5"
    >
      <DateInput
        name={name}
        label={props.label}
        class={props.class}
        value={value()}
        onBlur={field().handleBlur}
        onValueChange={(details) => {
          const first = details.value[0];
          if (first instanceof CalendarDate) {
            field().handleChange(calendarDateToTimestamp(first));
          } else if (!first) {
            field().handleChange(undefined);
          }
        }}
      />
      <For each={errors()}>
        {(err) => <Field.ErrorText class="text-xs text-ctp-red">{err}</Field.ErrorText>}
      </For>
    </Field.Root>
  );
};
