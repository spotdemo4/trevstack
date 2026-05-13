import { ListRequestSchema, type ListRequest } from "$connect/number/v1/list_pb";
import { NumberClient } from "$lib/connect";
import Drawer from "$lib/drawer";
import { useForm } from "$lib/form/hook";
import Splitter from "$lib/splitter";
import { createStreamingStore } from "$lib/stream";
import Table from "$lib/table";
import { create } from "@bufbuild/protobuf";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import { createStandardSchema } from "@bufbuild/protovalidate";
import { SlidersHorizontal } from "lucide-solid";
import { type Component, createSignal, onCleanup, Show } from "solid-js";

function createMediaQuery(query: string) {
  const mql = window.matchMedia(query);
  const [matches, setMatches] = createSignal(mql.matches);
  const handler = (e: MediaQueryListEvent) => setMatches(e.matches);
  mql.addEventListener("change", handler);
  onCleanup(() => mql.removeEventListener("change", handler));
  return matches;
}

const Numbers: Component = () => {
  const isDesktop = createMediaQuery("(min-width: 768px)");

  const [request, setRequest] = createSignal<ListRequest>(create(ListRequestSchema));
  const items = createStreamingStore(
    request,
    (req) => NumberClient.list(req),
    (resp) => resp.item!,
  );

  const form = useForm(() => ({
    defaultValues: create(ListRequestSchema),
    validators: {
      onChange: createStandardSchema(ListRequestSchema),
    },
    onSubmit: async ({ value }) => setRequest(value),
  }));

  const FormContent: Component = () => (
    <form.AppForm>
      <form.Form class="justify-center">
        <form.AppField name="name">{(field) => <field.TextField label="Name" />}</form.AppField>
        <form.AppField name="min">{(field) => <field.NumberField label="Minimum" />}</form.AppField>
        <form.AppField name="max">{(field) => <field.NumberField label="Maximum" />}</form.AppField>
        <form.AppField name="start">{(field) => <field.DateField label="Start" />}</form.AppField>
        <form.AppField name="end">{(field) => <field.DateField label="End" />}</form.AppField>
        <form.SubmitButton label="Filter" />
      </form.Form>
    </form.AppForm>
  );

  const TableContent: Component = () => (
    <Table.Table columns={["12rem", "2fr", "1fr"]}>
      <Table.Header>
        <th>Timestamp</th>
        <th>Name</th>
        <th>Number</th>
      </Table.Header>
      <Table.Body items={items}>
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
          <div class="flex h-full flex-col">
            <Drawer.Root>
              <Drawer.Trigger>
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
            <div class="min-h-0 grow">
              <TableContent />
            </div>
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

export default Numbers;
