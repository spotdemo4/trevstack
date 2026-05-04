import { ListRequestSchema, type ListRequest, type ListResponse } from "$connect/number/v1/list_pb";
import Splitter from "$lib/splitter";
import Table from "$lib/table";
import { NumberClient } from "$lib/transport";
import { create } from "@bufbuild/protobuf";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import { type Component, createResource, createSignal } from "solid-js";

import Form from "./Form";

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

  return (
    <div class="h-body">
      <Splitter.Root
        panels={[
          {
            id: "a",
            minSize: 15,
            maxSize: 50,
          },
          { id: "b" },
        ]}
        defaultSize={[15, 50]}
      >
        <Splitter.Panel id="a" class="bg-ctp-mantle p-4">
          <Form onSubmit={setRequest} />
        </Splitter.Panel>

        <Splitter.ResizeTrigger id="a:b" />

        <Splitter.Panel id="b">
          <Table.Table
            onScroll={async (_, end) => {
              while (!response()?.items[end]) {
                await refetch();
              }
            }}
          >
            <Table.Header>
              <span class="w-40">Timestamp</span>
              <span>Name</span>
              <span>Number</span>
            </Table.Header>
            <Table.Rows count={() => response()?.totalCount} items={() => response()?.items ?? []}>
              {(item) => (
                <>
                  <span class="text-sm text-ctp-subtext0 tabular-nums">
                    {timestampDate(item.timestamp!).toLocaleString()}
                  </span>
                  <span class="truncate font-medium">{item.name}</span>
                  <span class="font-mono tabular-nums">{item.number}</span>
                </>
              )}
            </Table.Rows>
          </Table.Table>
        </Splitter.Panel>
      </Splitter.Root>
    </div>
  );
};

export default Numbers;
