import { ListStreamsRequestSchema } from "$gen/cursor/v1/list_pb";
import { sourceItems, sourceLabel } from "$lib/cursor-pad";
import { createStreamingStore } from "$lib/effect";
import { DateField } from "$lib/form/date-field";
import { Form } from "$lib/form/form";
import { useForm } from "$lib/form/hook";
import { protoSchema } from "$lib/form/proto-schema";
import { SelectField } from "$lib/form/select-field";
import { SubmitButton } from "$lib/form/submit-button";
import { TextField } from "$lib/form/text-field";
import { formatDuration, formatTimestamp } from "$lib/format";
import { CursorClient } from "$lib/rpc";
import { Table, type TableScrollMode } from "$lib/table";
import type { MessageInitShape } from "@bufbuild/protobuf";
import { type Component, createSignal } from "solid-js";

import { FilteredTable } from "./filtered-table";

/** The filters and results of the streams table, which outlive its tab. */
export function createStreams() {
  const [request, setRequest] = createSignal<MessageInitShape<typeof ListStreamsRequestSchema>>({});
  const stream = createStreamingStore(
    request,
    (req) => CursorClient.listStreams(req),
    (resp) => resp.stream!,
  );

  const form = useForm(protoSchema(ListStreamsRequestSchema), () => ({
    validation: {
      onChange: true,
    },
    onSubmit: async ({ value }) => {
      setRequest(value);
    },
  }));

  return { stream, form };
}

export type StreamsState = ReturnType<typeof createStreams>;

export const Streams: Component<{ state: StreamsState }> = (props) => {
  const form = props.state.form;
  const stream = props.state.stream;

  const Filters: Component = () => (
    <Form form={form} class="justify-center">
      <TextField field={form.field("username")} label="User" />
      <SelectField field={form.field("source")} label="Source" items={sourceItems} />
      <DateField field={form.field("start")} label="Start" />
      <DateField field={form.field("end")} label="End" />
      <SubmitButton form={form} label="Filter" />
    </Form>
  );

  const StreamsTable: Component<{ scrollMode?: TableScrollMode }> = (tableProps) => (
    <Table.Table
      columns={["5rem", "minmax(6rem, 2fr)", "5rem", "14rem", "6rem", "6rem", "6rem"]}
      scrollMode={tableProps.scrollMode}
    >
      <Table.Header>
        <th>ID</th>
        <th>User</th>
        <th>Source</th>
        <th>Start</th>
        <th>Duration</th>
        <th>Positions</th>
        <th>Distance</th>
      </Table.Header>
      <Table.Body items={stream.items} loading={stream.loading()}>
        {(summary) => (
          <>
            <td class="font-mono tabular-nums">{summary.id.toString()}</td>
            <td class="truncate font-medium">{summary.username}</td>
            <td>{sourceLabel(summary.source)}</td>
            <td class="text-sm text-foreground-subtle tabular-nums">
              {formatTimestamp(summary.start)}
            </td>
            <td class="tabular-nums">{formatDuration(summary.start, summary.end)}</td>
            <td class="font-mono tabular-nums">{summary.count.toLocaleString()}</td>
            <td class="font-mono tabular-nums">{summary.distance.toFixed(2)}</td>
          </>
        )}
      </Table.Body>
    </Table.Table>
  );

  return <FilteredTable filters={Filters} table={StreamsTable} />;
};
