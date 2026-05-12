import {
  DateInput as ArkDateInput,
  type DateInputValueChangeDetails,
  useDateInput,
} from "@ark-ui/solid/date-input";
import { DatePicker, useDatePicker } from "@ark-ui/solid/date-picker";
import type { DateValue } from "@internationalized/date";
import { mergeProps } from "@zag-js/solid";
import { CalendarIcon, ChevronLeftIcon, ChevronRightIcon } from "lucide-solid";
import { type Component, type JSX, Index, Show } from "solid-js";
import { Portal } from "solid-js/web";
import { twMerge } from "tailwind-merge";

type DateInputProps = {
  class?: string;
  label?: string;
  name?: string;
  value?: DateValue[];
  onValueChange?: (details: DateInputValueChangeDetails) => void;
  onBlur?: JSX.FocusEventHandler<HTMLDivElement, FocusEvent>;
};

const iconTriggerClass =
  "inline-flex h-7 w-7 items-center justify-center rounded-md border border-ctp-surface1 bg-ctp-surface0 text-ctp-subtext0 shadow-sm transition-colors hover:cursor-pointer hover:border-ctp-surface2 hover:bg-ctp-surface1 hover:text-ctp-text focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ctp-sky/40 active:bg-ctp-surface2";

export const DateInput: Component<DateInputProps> = (props) => {
  const datePicker = useDatePicker(() => ({
    value: props.value,
    onValueChange: (details) => props.onValueChange?.(details),
  }));

  const dateInput = useDateInput(() => ({
    value: props.value,
    onValueChange: (details) => props.onValueChange?.(details),
  }));

  return (
    <ArkDateInput.RootProvider value={dateInput} class="flex min-w-42 flex-col gap-1.5">
      <Show when={props.label}>
        <ArkDateInput.Label class="text-sm font-medium text-ctp-subtext1 data-invalid:text-ctp-red">
          {props.label}
        </ArkDateInput.Label>
      </Show>
      <ArkDateInput.Control
        class={twMerge(
          "flex items-center gap-2 rounded-md border border-ctp-surface1 bg-ctp-base p-1 pl-3 text-sm text-ctp-text transition-colors focus-within:border-ctp-sky focus-within:ring-2 focus-within:ring-ctp-sky/40 hover:border-ctp-surface2 data-invalid:border-ctp-red data-invalid:focus-within:ring-ctp-red/40",
          props.class,
        )}
      >
        <DatePicker.RootProvider value={datePicker} class="flex flex-1 items-center">
          <DatePicker.Control class="flex flex-1 items-center gap-2">
            <ArkDateInput.SegmentGroup
              onFocusOut={props.onBlur}
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
            </ArkDateInput.SegmentGroup>
            <DatePicker.Trigger
              aria-label="Open calendar"
              class={twMerge(iconTriggerClass, "ml-auto")}
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
                                      <DatePicker.TableCell value={day()} class="p-0.5 text-center">
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
      </ArkDateInput.Control>
      <ArkDateInput.HiddenInput name={props.name} />
    </ArkDateInput.RootProvider>
  );
};
