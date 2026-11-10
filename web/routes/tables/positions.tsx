import { ListRequestSchema } from "$gen/cursor/v1/list_pb";
import { sourceItems, sourceLabel } from "$lib/cursor-pad";
import { createStreamingStore } from "$lib/effect";
import { DateField } from "$lib/form/date-field";
import { Form } from "$lib/form/form";
import { useForm } from "$lib/form/hook";
import { NumberField } from "$lib/form/number-field";
import { protoSchema } from "$lib/form/proto-schema";
import { SelectField } from "$lib/form/select-field";
import { SubmitButton } from "$lib/form/submit-button";
import { TextField } from "$lib/form/text-field";
import { formatCoordinate, formatTimestamp } from "$lib/format";
import { CursorClient } from "$lib/rpc";
import { Table, type TableScrollMode } from "$lib/table";
import type { MessageInitShape } from "@bufbuild/protobuf";
import { type Component, createSignal } from "solid-js";

import { FilteredTable } from "./filtered-table";

/** The filters and results of the positions table, which outlive its tab. */
export function createPositions() {
  const [request, setRequest] = createSignal<MessageInitShape<typeof ListRequestSchema>>({});
  const stream = createStreamingStore(
    request,
    (req) => CursorClient.list(req),
    (resp) => resp.cursor!,
  );

  const form = useForm(protoSchema(ListRequestSchema), () => ({
    validation: {
      onChange: true,
    },
    onSubmit: async ({ value }) => {
      setRequest(value);
    },
  }));

  return { stream, form };
}

export type PositionsState = ReturnType<typeof createPositions>;

export const Positions: Component<{ state: PositionsState }> = (props) => {
  const form = props.state.form;
  const stream = props.state.stream;

  const Filters: Component = () => (
    <Form form={form} class="justify-center">
      <TextField field={form.field("username")} label="User" />
      <SelectField field={form.field("source")} label="Source" items={sourceItems} />
      <NumberField field={form.field("streamId")} label="Stream" min={1} step={1} />
      <div class="grid grid-cols-2 gap-2">
        <NumberField field={form.field("minX")} label="Min X" min={0} max={1} step={0.01} />
        <NumberField field={form.field("maxX")} label="Max X" min={0} max={1} step={0.01} />
        <NumberField field={form.field("minY")} label="Min Y" min={0} max={1} step={0.01} />
        <NumberField field={form.field("maxY")} label="Max Y" min={0} max={1} step={0.01} />
      </div>
      <DateField field={form.field("start")} label="Start" />
      <DateField field={form.field("end")} label="End" />
      <SubmitButton form={form} label="Filter" />
    </Form>
  );

  const PositionsTable: Component<{ scrollMode?: TableScrollMode }> = (tableProps) => (
    <Table.Table
      columns={["14rem", "minmax(6rem, 2fr)", "5rem", "5rem", "5rem", "5rem"]}
      scrollMode={tableProps.scrollMode}
    >
      <Table.Header>
        <th>Timestamp</th>
        <th>User</th>
        <th>Source</th>
        <th>Stream</th>
        <th>X</th>
        <th>Y</th>
      </Table.Header>
      <Table.Body items={stream.items} loading={stream.loading()}>
        {(cursor) => (
          <>
            <td class="text-sm text-foreground-subtle tabular-nums">
              {formatTimestamp(cursor.timestamp)}
            </td>
            <td class="truncate font-medium">{cursor.username}</td>
            <td>{sourceLabel(cursor.source)}</td>
            <td class="font-mono tabular-nums">{cursor.streamId.toString()}</td>
            <td class="font-mono tabular-nums">{formatCoordinate(cursor.x)}</td>
            <td class="font-mono tabular-nums">{formatCoordinate(cursor.y)}</td>
          </>
        )}
      </Table.Body>
    </Table.Table>
  );

  return <FilteredTable filters={Filters} table={PositionsTable} />;
};
