import { ListRequestSchema, type ListRequest } from "$connect/number/v1/list_pb";
import { NumberClient } from "$lib/connect";
import { Drawer } from "$lib/drawer";
import { createStreamingStore } from "$lib/effect";
import { DateField } from "$lib/form/date-field";
import { Form } from "$lib/form/form";
import { useForm } from "$lib/form/hook";
import { NumberField } from "$lib/form/number-field";
import { SubmitButton } from "$lib/form/submit-button";
import { TextField } from "$lib/form/text-field";
import { SlidersHorizontal } from "$lib/icon";
import { createMediaQuery } from "$lib/media-query";
import { Splitter } from "$lib/splitter";
import { Table } from "$lib/table";
import { create } from "@bufbuild/protobuf";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import { createStandardSchema } from "@bufbuild/protovalidate";
import { type Component, createSignal, Show } from "solid-js";

export const Numbers: Component = () => {
  const isDesktop = createMediaQuery("(min-width: 900px)");

  const [request, setRequest] = createSignal<ListRequest>(create(ListRequestSchema));
  const stream = createStreamingStore(
    request,
    (req) => NumberClient.list(req),
    (resp) => resp.item!,
  );

  const form = useForm(() => ({
    defaultValues: { ...create(ListRequestSchema) },
    validators: {
      onChange: createStandardSchema(ListRequestSchema),
    },
    onSubmit: async ({ value }) => {
      setRequest(value);
    },
  }));

  const FormContent: Component = () => (
    <Form form={form} class="justify-center">
      <TextField field={form.field("name")} label="Name" />
      <NumberField field={form.field("minimum")} label="Minimum" />
      <NumberField field={form.field("maximum")} label="Maximum" />
      <DateField field={form.field("start")} label="Start" />
      <DateField field={form.field("end")} label="End" />
      <SubmitButton form={form} label="Filter" />
    </Form>
  );

  const TableContent: Component = () => (
    <Table.Table columns={["12rem", "2fr", "1fr"]}>
      <Table.Header>
        <th>Timestamp</th>
        <th>Name</th>
        <th>Number</th>
      </Table.Header>
      <Table.Body items={stream.items} loading={stream.loading()}>
        {(item) => (
          <>
            <td class="text-sm text-ctp-subtext0 tabular-nums">
              {timestampDate(item.timestamp!).toLocaleString()}
            </td>
            <td class="truncate font-medium">{item.name}</td>
            <td class="font-mono tabular-nums">{item.number}</td>
          </>
        )}
      </Table.Body>
    </Table.Table>
  );

  return (
    <div class="h-body">
      <Show
        when={isDesktop()}
        fallback={
          <div class="relative h-full">
            <div class="h-full min-h-0">
              <TableContent />
            </div>
            <Drawer.Root>
              <Drawer.Trigger class="absolute right-4 bottom-4 z-10 m-0 shadow-lg">
                <SlidersHorizontal size={16} /> Filters
              </Drawer.Trigger>
              <Drawer.Content>
                <div class="flex items-center justify-between">
                  <Drawer.Title>Filters</Drawer.Title>
                  <Drawer.CloseTrigger />
                </div>
                <FormContent />
              </Drawer.Content>
            </Drawer.Root>
          </div>
        }
      >
        <Splitter.Root
          class="h-full"
          defaultSize={[15, 50]}
          panels={[{ id: "a", minSize: 20, maxSize: 50 }, { id: "b" }]}
        >
          <Splitter.Panel id="a" class="bg-ctp-mantle p-4">
            <FormContent />
          </Splitter.Panel>
          <Splitter.ResizeTrigger id="a:b" />
          <Splitter.Panel id="b" class="min-h-0">
            <TableContent />
          </Splitter.Panel>
        </Splitter.Root>
      </Show>
    </div>
  );
};
