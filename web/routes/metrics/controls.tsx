import { TimeInterval } from "$gen/number/v1/metrics_pb";
import { NumberInput, SelectInput } from "$lib/input";
import type { Component } from "solid-js";

const intervalItems = [
  { value: TimeInterval.HOUR, label: "Hour" },
  { value: TimeInterval.DAY, label: "Day" },
  { value: TimeInterval.WEEK, label: "Week" },
  { value: TimeInterval.MONTH, label: "Month" },
].map((item) => ({ label: item.label, value: String(item.value) }));

export const IntervalSelect: Component<{
  value: TimeInterval;
  onChange: (value: TimeInterval) => void;
}> = (props) => (
  <div class="w-36">
    <SelectInput
      items={intervalItems}
      value={[String(props.value)]}
      onValueChange={(details) => {
        const first = details.value[0];
        if (first !== undefined) props.onChange(Number(first) as TimeInterval);
      }}
    />
  </div>
);

const MIN_COUNT = 1;
const MAX_COUNT = 100;

/** Number input that only reports whole values within 1-100. */
export const CountInput: Component<{
  value: number;
  onChange: (value: number) => void;
}> = (props) => (
  <div class="w-32">
    <NumberInput
      min={MIN_COUNT}
      max={MAX_COUNT}
      value={String(props.value)}
      onValueChange={({ valueAsNumber: n }) => {
        if (Number.isInteger(n) && n >= MIN_COUNT && n <= MAX_COUNT) props.onChange(n);
      }}
    />
  </div>
);
