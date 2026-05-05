import { DateInput, useDateInput } from "@ark-ui/solid/date-input";
import { DatePicker, useDatePicker } from "@ark-ui/solid/date-picker";
import { Field } from "@ark-ui/solid/field";
import { type Timestamp, timestampDate, timestampFromDate } from "@bufbuild/protobuf/wkt";
import { CalendarDate, parseDate, fromDate, toCalendarDate } from "@internationalized/date";
import { mergeProps } from "@zag-js/solid";
import { CalendarIcon, ChevronLeftIcon, ChevronRightIcon } from "lucide-solid";
import { createMemo, For, Index, Show } from "solid-js";
import { Portal } from "solid-js/web";

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

export function DateField(props: { label?: string }) {
  const field = useFieldContext<Timestamp | undefined>();
  const name = field().name;
  const errors = createMemo(() => [
    ...new Set(field().state.meta.errors.map((err) => err.message as string)),
  ]);

  const value = createMemo(() => {
    const str = timestampToDateString(field().state.value);
    return str ? [parseDate(str)] : [];
  });

  const handleValueChange = (newValue: readonly unknown[]) => {
    const first = newValue[0];
    if (first instanceof CalendarDate) {
      field().handleChange(calendarDateToTimestamp(first));
    } else if (!first) {
      field().form.resetField(name);
    }
  };

  const datePicker = useDatePicker(() => ({
    value: value(),
    onValueChange: (details) => handleValueChange(details.value),
  }));

  const dateInput = useDateInput(() => ({
    value: value(),
    onValueChange: (details) => handleValueChange(details.value),
  }));

  return (
    <Field.Root
      invalid={!(field().state.meta.isValid || !field().state.meta.isBlurred)}
      class="flex flex-col gap-1.5"
    >
      <DateInput.RootProvider value={dateInput} class="flex flex-col gap-1.5">
        <Show when={props.label}>
          <DateInput.Label class="text-sm font-medium text-ctp-subtext1 data-invalid:text-ctp-red">
            {props.label}
          </DateInput.Label>
        </Show>
        <DateInput.Control class="flex items-center gap-2 rounded-md border border-ctp-surface1 bg-ctp-base px-3 py-2 text-sm text-ctp-text transition-colors focus-within:border-ctp-sky focus-within:ring-2 focus-within:ring-ctp-sky/40 hover:border-ctp-surface2 data-invalid:border-ctp-red data-invalid:focus-within:ring-ctp-red/40">
          <DatePicker.RootProvider value={datePicker} class="flex flex-1 items-center">
            <DatePicker.Control class="flex flex-1 items-center gap-2">
              <DateInput.SegmentGroup
                onFocusOut={field().handleBlur}
                class="flex flex-1 items-center gap-0.5"
              >
                <Index each={dateInput().getSegments({})}>
                  {(segment) => (
                    <span
                      {...mergeProps(() => dateInput().getSegmentProps({ segment: segment() }))}
                      class="rounded px-0.5 focus:bg-ctp-surface1 focus:outline-none"
                    >
                      {segment().text}
                    </span>
                  )}
                </Index>
              </DateInput.SegmentGroup>
              <DatePicker.Trigger
                aria-label="Open calendar"
                class="ml-auto cursor-pointer rounded p-0.5 transition-colors hover:bg-ctp-surface1"
              >
                <CalendarIcon size={16} />
              </DatePicker.Trigger>
            </DatePicker.Control>
            <Portal>
              <DatePicker.Positioner>
                <DatePicker.Content class="z-50 w-72 rounded-md border border-ctp-surface1 bg-ctp-base p-3 text-ctp-text shadow-lg">
                  <DatePicker.View view="day">
                    <DatePicker.Context>
                      {(datePicker) => (
                        <>
                          <DatePicker.ViewControl class="mb-2 flex items-center justify-between">
                            <DatePicker.PrevTrigger
                              aria-label="Previous"
                              class="cursor-pointer rounded p-1 transition-colors hover:bg-ctp-surface1"
                            >
                              <ChevronLeftIcon size={16} />
                            </DatePicker.PrevTrigger>
                            <DatePicker.ViewTrigger class="cursor-pointer rounded px-2 py-1 text-sm font-medium transition-colors hover:bg-ctp-surface1">
                              <DatePicker.RangeText />
                            </DatePicker.ViewTrigger>
                            <DatePicker.NextTrigger
                              aria-label="Next"
                              class="cursor-pointer rounded p-1 transition-colors hover:bg-ctp-surface1"
                            >
                              <ChevronRightIcon size={16} />
                            </DatePicker.NextTrigger>
                          </DatePicker.ViewControl>
                          <DatePicker.Table class="w-full table-fixed border-collapse">
                            <DatePicker.TableHead>
                              <DatePicker.TableRow>
                                <Index each={datePicker().weekDays}>
                                  {(weekDay) => (
                                    <DatePicker.TableHeader class="py-1 text-center text-xs font-normal text-ctp-subtext0">
                                      {weekDay().short}
                                    </DatePicker.TableHeader>
                                  )}
                                </Index>
                              </DatePicker.TableRow>
                            </DatePicker.TableHead>
                            <DatePicker.TableBody>
                              <Index each={datePicker().weeks}>
                                {(week) => (
                                  <DatePicker.TableRow>
                                    <Index each={week()}>
                                      {(day) => (
                                        <DatePicker.TableCell
                                          value={day()}
                                          class="p-0.5 text-center"
                                        >
                                          <DatePicker.TableCellTrigger class="inline-flex aspect-square w-full items-center justify-center rounded text-sm tabular-nums transition-colors hover:cursor-pointer hover:bg-ctp-surface1 data-outside-range:text-ctp-overlay0 data-selected:bg-ctp-sky data-selected:text-ctp-base">
                                            {day().day}
                                          </DatePicker.TableCellTrigger>
                                        </DatePicker.TableCell>
                                      )}
                                    </Index>
                                  </DatePicker.TableRow>
                                )}
                              </Index>
                            </DatePicker.TableBody>
                          </DatePicker.Table>
                        </>
                      )}
                    </DatePicker.Context>
                  </DatePicker.View>
                  <DatePicker.View view="month">
                    <DatePicker.Context>
                      {(datePicker) => (
                        <>
                          <DatePicker.ViewControl class="mb-2 flex items-center justify-between">
                            <DatePicker.PrevTrigger
                              aria-label="Previous"
                              class="cursor-pointer rounded p-1 transition-colors hover:bg-ctp-surface1"
                            >
                              <ChevronLeftIcon size={16} />
                            </DatePicker.PrevTrigger>
                            <DatePicker.ViewTrigger class="cursor-pointer rounded px-2 py-1 text-sm font-medium transition-colors hover:bg-ctp-surface1">
                              <DatePicker.RangeText />
                            </DatePicker.ViewTrigger>
                            <DatePicker.NextTrigger
                              aria-label="Next"
                              class="cursor-pointer rounded p-1 transition-colors hover:bg-ctp-surface1"
                            >
                              <ChevronRightIcon size={16} />
                            </DatePicker.NextTrigger>
                          </DatePicker.ViewControl>
                          <DatePicker.Table class="w-full table-fixed border-collapse">
                            <DatePicker.TableBody>
                              <Index
                                each={datePicker().getMonthsGrid({ columns: 4, format: "short" })}
                              >
                                {(months) => (
                                  <DatePicker.TableRow>
                                    <Index each={months()}>
                                      {(month) => (
                                        <DatePicker.TableCell
                                          value={month().value}
                                          class="p-0.5 text-center"
                                        >
                                          <DatePicker.TableCellTrigger class="inline-flex w-full items-center justify-center rounded px-2 py-2 text-sm transition-colors hover:cursor-pointer hover:bg-ctp-surface1 data-selected:bg-ctp-sky data-selected:text-ctp-base">
                                            {month().label}
                                          </DatePicker.TableCellTrigger>
                                        </DatePicker.TableCell>
                                      )}
                                    </Index>
                                  </DatePicker.TableRow>
                                )}
                              </Index>
                            </DatePicker.TableBody>
                          </DatePicker.Table>
                        </>
                      )}
                    </DatePicker.Context>
                  </DatePicker.View>
                  <DatePicker.View view="year">
                    <DatePicker.Context>
                      {(datePicker) => (
                        <>
                          <DatePicker.ViewControl class="mb-2 flex items-center justify-between">
                            <DatePicker.PrevTrigger
                              aria-label="Previous"
                              class="cursor-pointer rounded p-1 transition-colors hover:bg-ctp-surface1"
                            >
                              <ChevronLeftIcon size={16} />
                            </DatePicker.PrevTrigger>
                            <DatePicker.ViewTrigger class="cursor-pointer rounded px-2 py-1 text-sm font-medium transition-colors hover:bg-ctp-surface1">
                              <DatePicker.RangeText />
                            </DatePicker.ViewTrigger>
                            <DatePicker.NextTrigger
                              aria-label="Next"
                              class="cursor-pointer rounded p-1 transition-colors hover:bg-ctp-surface1"
                            >
                              <ChevronRightIcon size={16} />
                            </DatePicker.NextTrigger>
                          </DatePicker.ViewControl>
                          <DatePicker.Table class="w-full table-fixed border-collapse">
                            <DatePicker.TableBody>
                              <Index each={datePicker().getYearsGrid({ columns: 4 })}>
                                {(years) => (
                                  <DatePicker.TableRow>
                                    <Index each={years()}>
                                      {(year) => (
                                        <DatePicker.TableCell
                                          value={year().value}
                                          class="p-0.5 text-center"
                                        >
                                          <DatePicker.TableCellTrigger class="inline-flex w-full items-center justify-center rounded px-2 py-2 text-sm tabular-nums transition-colors hover:cursor-pointer hover:bg-ctp-surface1 data-selected:bg-ctp-sky data-selected:text-ctp-base">
                                            {year().label}
                                          </DatePicker.TableCellTrigger>
                                        </DatePicker.TableCell>
                                      )}
                                    </Index>
                                  </DatePicker.TableRow>
                                )}
                              </Index>
                            </DatePicker.TableBody>
                          </DatePicker.Table>
                        </>
                      )}
                    </DatePicker.Context>
                  </DatePicker.View>
                </DatePicker.Content>
              </DatePicker.Positioner>
            </Portal>
          </DatePicker.RootProvider>
        </DateInput.Control>
        <DateInput.HiddenInput name={name} />
      </DateInput.RootProvider>
      <For each={errors()}>
        {(err) => <Field.ErrorText class="text-xs text-ctp-red">{err}</Field.ErrorText>}
      </For>
    </Field.Root>
  );
}
