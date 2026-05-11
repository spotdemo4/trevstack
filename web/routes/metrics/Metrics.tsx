import {
  type DistributionResponse,
  type SummaryResponse,
  TimeInterval,
  type TimeSeriesResponse,
  type TopNamesResponse,
} from "$connect/number/v1/metrics_pb";
import { Card } from "$lib/card";
import { useForm } from "$lib/form/hook";
import { NumberInput, SelectInput } from "$lib/input";
import { NumberClient } from "$lib/transport";
import { type Timestamp } from "@bufbuild/protobuf/wkt";
import { type Component, createMemo, createResource, createSignal } from "solid-js";

import DistributionChart from "./DistributionChart";
import SummaryCards from "./SummaryCards";
import TimeSeriesChart from "./TimeSeriesChart";
import TopNamesChart from "./TopNamesChart";

const intervalOptions: { value: TimeInterval; label: string }[] = [
  { value: TimeInterval.HOUR, label: "Hour" },
  { value: TimeInterval.DAY, label: "Day" },
  { value: TimeInterval.WEEK, label: "Week" },
  { value: TimeInterval.MONTH, label: "Month" },
];

type MetricsRange = {
  start?: Timestamp;
  end?: Timestamp;
};

const Metrics: Component = () => {
  const [rangeFilter, setRangeFilter] = createSignal<MetricsRange>({});
  const [interval, setInterval] = createSignal<TimeInterval>(TimeInterval.DAY);
  const [bucketCount, setBucketCount] = createSignal(10);
  const [limit, setLimit] = createSignal(8);

  const form = useForm(() => ({
    defaultValues: {} as MetricsRange,
    onSubmit: async ({ value }) => setRangeFilter(value),
  }));

  const range = createMemo(() => rangeFilter());

  const [summary] = createResource<SummaryResponse | undefined, ReturnType<typeof range>>(
    range,
    async (req) => {
      const [resp, err] = await NumberClient.summary(req);
      if (err) {
        console.error("summary failed", err);
        return undefined;
      }
      return resp;
    },
  );

  const timeSeriesArgs = createMemo(() => ({ ...range(), interval: interval() }));
  const [timeSeries] = createResource<
    TimeSeriesResponse | undefined,
    ReturnType<typeof timeSeriesArgs>
  >(timeSeriesArgs, async (req) => {
    const [resp, err] = await NumberClient.timeSeries(req);
    if (err) {
      console.error("timeSeries failed", err);
      return undefined;
    }
    return resp;
  });

  const distributionArgs = createMemo(() => ({ ...range(), bucketCount: bucketCount() }));
  const [distribution] = createResource<
    DistributionResponse | undefined,
    ReturnType<typeof distributionArgs>
  >(distributionArgs, async (req) => {
    const [resp, err] = await NumberClient.distribution(req);
    if (err) {
      console.error("distribution failed", err);
      return undefined;
    }
    return resp;
  });

  const topNamesArgs = createMemo(() => ({ ...range(), limit: limit() }));
  const [topNames] = createResource<TopNamesResponse | undefined, ReturnType<typeof topNamesArgs>>(
    topNamesArgs,
    async (req) => {
      const [resp, err] = await NumberClient.topNames(req);
      if (err) {
        console.error("topNames failed", err);
        return undefined;
      }
      return resp;
    },
  );

  return (
    <div class="mx-auto flex max-w-7xl flex-col gap-4 p-4">
      <form.AppForm>
        <form.Form class="flex flex-row items-center gap-2">
          <form.AppField name="start">
            {(field) => <field.DateField class="h-9 bg-ctp-mantle" />}
          </form.AppField>
          <form.AppField name="end">
            {(field) => <field.DateField class="h-9 bg-ctp-mantle" />}
          </form.AppField>
          <form.SubmitButton class="h-9" />
          <form.ResetButton class="h-9" />
        </form.Form>
      </form.AppForm>

      <SummaryCards data={summary()} loading={summary.loading} />

      <Card class="p-4">
        <div class="mb-3 flex items-center justify-between gap-2">
          <h2 class="font-semibold text-ctp-text">Activity over time</h2>
          <div class="w-36">
            <SelectInput
              items={intervalOptions.map((opt) => ({
                label: opt.label,
                value: String(opt.value),
              }))}
              value={[String(interval())]}
              placeholder="Select interval"
              onValueChange={(details) => {
                const first = details.value[0];
                if (typeof first === "string") {
                  setInterval(Number(first) as TimeInterval);
                }
              }}
            />
          </div>
        </div>
        <TimeSeriesChart points={timeSeries()?.points ?? []} />
      </Card>

      <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card class="p-4">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h2 class="font-semibold text-ctp-text">Number distribution</h2>
            <div class="w-32">
              <NumberInput
                min={1}
                max={100}
                value={String(bucketCount())}
                onValueChange={(c) => {
                  if (
                    !Number.isNaN(c.valueAsNumber) &&
                    c.valueAsNumber >= 1 &&
                    c.valueAsNumber <= 100
                  ) {
                    setBucketCount(c.valueAsNumber);
                  }
                }}
              />
            </div>
          </div>
          <DistributionChart buckets={distribution()?.buckets ?? []} />
        </Card>

        <Card class="p-4">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h2 class="font-semibold text-ctp-text">Top names</h2>
            <div class="w-32">
              <NumberInput
                min={1}
                max={100}
                value={String(limit())}
                onValueChange={(c) => {
                  if (
                    !Number.isNaN(c.valueAsNumber) &&
                    c.valueAsNumber >= 1 &&
                    c.valueAsNumber <= 100
                  ) {
                    setLimit(c.valueAsNumber);
                  }
                }}
              />
            </div>
          </div>
          <TopNamesChart names={topNames()?.names ?? []} />
        </Card>
      </div>
    </div>
  );
};

export default Metrics;
