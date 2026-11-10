import { TimeInterval } from "$gen/cursor/v1/metrics_pb";
import { createEffectResource, toastFailure } from "$lib/effect";
import { DateField } from "$lib/form/date-field";
import { Form } from "$lib/form/form";
import { useForm } from "$lib/form/hook";
import { ResetButton } from "$lib/form/reset-button";
import { SubmitButton } from "$lib/form/submit-button";
import { CursorClient } from "$lib/rpc";
import type { Timestamp } from "@bufbuild/protobuf/wkt";
import { Errored, isPending, Loading, type Component, createSignal } from "solid-js";

import { ChartCard } from "./chart-card";
import { CountInput, IntervalSelect } from "./controls";
import { HeatmapChart } from "./heatmap-chart";
import { SummaryCards } from "./summary-cards";
import { TimeSeriesChart } from "./time-series-chart";
import { TopUsersChart } from "./top-users-chart";

type MetricsRange = {
  start?: Timestamp;
  end?: Timestamp;
};

const Metrics: Component = () => {
  // Shared by every request; set by the filter form.
  const [range, setRange] = createSignal<MetricsRange>({});

  const form = useForm(() => ({
    defaultValues: {} as MetricsRange,
    onSubmit: async ({ value }) => {
      setRange(value);
    },
  }));

  const summary = createEffectResource(range, (req) =>
    toastFailure("Failed to load summary")(CursorClient.summary(req)),
  );

  const [interval, setInterval] = createSignal<TimeInterval>(TimeInterval.MINUTE);
  const timeSeries = createEffectResource(
    () => ({ ...range(), interval: interval() }),
    (req) => toastFailure("Failed to load activity over time")(CursorClient.timeSeries(req)),
  );

  // Rows follow columns, so the cells are about square on the 16:9 pad.
  const [columns, setColumns] = createSignal(32);
  const heatmap = createEffectResource(
    () => ({
      ...range(),
      columns: columns(),
      rows: Math.min(100, Math.max(1, Math.round((columns() * 9) / 16))),
    }),
    (req) => toastFailure("Failed to load heatmap")(CursorClient.heatmap(req)),
  );

  const [limit, setLimit] = createSignal(8);
  const topUsers = createEffectResource(
    () => ({ ...range(), limit: limit() }),
    (req) => toastFailure("Failed to load top users")(CursorClient.topUsers(req)),
  );

  return (
    <div class="mx-auto flex max-w-7xl flex-col gap-4 p-4">
      <Form
        form={form}
        class="flex flex-row flex-wrap items-end justify-center gap-2 md:justify-start"
      >
        <DateField field={form.field("start")} label="Start" class="bg-surface" />
        <DateField field={form.field("end")} label="End" class="bg-surface" />
        <SubmitButton form={form} />
        <ResetButton form={form} />
      </Form>

      <Errored fallback={<div class="text-sm text-danger">Failed to load summary.</div>}>
        <Loading fallback={<SummaryCards />}>
          <SummaryCards data={summary()} loading={isPending(summary)} />
        </Loading>
      </Errored>

      <ChartCard
        title="Activity over time"
        height={280}
        errorLabel="Failed to load activity."
        controls={<IntervalSelect value={interval()} onChange={setInterval} />}
      >
        <TimeSeriesChart points={timeSeries()?.points ?? []} height={280} />
      </ChartCard>

      <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <ChartCard
          title="Heatmap"
          height={260}
          errorLabel="Failed to load heatmap."
          controls={<CountInput value={columns()} onChange={setColumns} label="Columns" />}
        >
          <HeatmapChart heatmap={heatmap()} emptyHeight={260} />
        </ChartCard>

        <ChartCard
          title="Top users"
          height={260}
          errorLabel="Failed to load top users."
          controls={<CountInput value={limit()} onChange={setLimit} label="Users" />}
        >
          <TopUsersChart users={topUsers()?.users ?? []} minHeight={260} />
        </ChartCard>
      </div>
    </div>
  );
};

export default Metrics;
