import { ListRequestSchema, type ListRequest, type ListResponse } from "$connect/number/v1/list_pb";
import Splitter from "$lib/splitter";
import { NumberClient } from "$lib/transport";
import { create } from "@bufbuild/protobuf";
import { type Component, createMemo, createResource, createSignal } from "solid-js";

import Form from "./Form";
import Table from "./Table";

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
          <Form onSubmit={(req) => setRequest(req)} />
        </Splitter.Panel>

        <Splitter.ResizeTrigger id="a:b" />

        <Splitter.Panel id="b">
          <Table
            count={createMemo(() => response()?.totalCount)}
            items={() => response()?.items ?? []}
            onScroll={(_, end) => {
              if (response()?.items[end]) return;
              refetch();
            }}
          />
        </Splitter.Panel>
      </Splitter.Root>
    </div>
  );
};

export default Numbers;
