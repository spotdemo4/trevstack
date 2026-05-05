import { ListRequestSchema, type ListRequest, type ListResponse } from "$connect/number/v1/list_pb";
import { useForm } from "$lib/form/hook";
import Splitter from "$lib/splitter";
import Table from "$lib/table";
import { NumberClient } from "$lib/transport";
import { create } from "@bufbuild/protobuf";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import { createStandardSchema } from "@bufbuild/protovalidate";
import { type Component, createResource, createSignal } from "solid-js";

const Numbers: Component = () => {
  const [request, setRequest] = createSignal<ListRequest>(create(ListRequestSchema));
  const [response, { refetch }] = createResource<ListResponse | undefined, ListRequest, boolean>(
    request,
    async (req, info) => {
      if (info.refetching) {
        req = { ...req, cursor: info.value?.nextCursor };
      }

      const [resp, err] = await NumberClient.list(req);
      if (err) {
        console.error("Failed to fetch numbers:", err);
        return info.value;
      }

      if (info.refetching) {
        resp.items.unshift(...(info.value?.items ?? []));
      }

      return resp;
    },
  );

  const form = useForm(() => ({
    defaultValues: create(ListRequestSchema),
    validators: {
      onChange: createStandardSchema(ListRequestSchema),
    },
    onSubmit: async ({ value }) => setRequest(value),
  }));

  return (
    <div class="h-body">
      <Splitter.Root
        defaultSize={[15, 50]}
        panels={[{ id: "a", minSize: 15, maxSize: 50 }, { id: "b" }]}
      >
        <Splitter.Panel id="a" class="bg-ctp-mantle p-4">
          <form.AppForm>
            <form.Form class="justify-center">
              <form.AppField name="name">
                {(field) => <field.TextField label="Name" />}
              </form.AppField>
              <form.AppField name="min">
                {(field) => <field.NumberField label="Minimum" />}
              </form.AppField>
              <form.AppField name="max">
                {(field) => <field.NumberField label="Maximum" />}
              </form.AppField>
              <form.AppField name="start">
                {(field) => <field.DateField label="Start" />}
              </form.AppField>
              <form.AppField name="end">{(field) => <field.DateField label="End" />}</form.AppField>
              <form.SubmitButton />
            </form.Form>
          </form.AppForm>
        </Splitter.Panel>
        <Splitter.ResizeTrigger id="a:b" />
        <Splitter.Panel id="b">
          <Table.Table
            columns={["12rem", "2fr", "1fr"]}
            onScroll={async (_, end) => {
              while (!response()?.items[end]) {
                await refetch();
              }
            }}
          >
            <Table.Header>
              <th>Timestamp</th>
              <th>Name</th>
              <th>Number</th>
            </Table.Header>
            <Table.Body count={() => response()?.totalCount} items={() => response()?.items ?? []}>
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
        </Splitter.Panel>
      </Splitter.Root>
    </div>
  );
};

export default Numbers;
