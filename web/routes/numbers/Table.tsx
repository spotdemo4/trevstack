import type { Item } from "$connect/number/v1/list_pb";
import { Virtualizer } from "$lib/virtualizer";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import { debounce } from "@solid-primitives/scheduled";
import { LoaderCircle } from "lucide-solid";
import type { Accessor } from "solid-js";
import { type Component, createMemo, Show } from "solid-js";

const Table: Component<{
  count: Accessor<bigint | undefined>;
  items: Accessor<Item[]>;
  onScroll?: (start: number, end: number) => void;
}> = (props) => {
  const count = createMemo(() => props.count());
  const onScroll = debounce(props.onScroll ?? (() => {}), 100);
  let start = 0;
  let end = 0;

  return (
    <div class="flex h-full flex-col">
      <div class="border-ctp-surface0 bg-ctp-base text-ctp-subtext1 flex w-full flex-row items-center gap-4 border-b px-4 pt-4 pb-2 text-xs font-semibold tracking-wide uppercase">
        <span class="w-48 shrink-0">Timestamp</span>
        <span class="flex-1">Name</span>
        <span class="shrink-0">Number</span>
      </div>
      <Show
        when={count()}
        fallback={<LoaderCircle class="text-ctp-subtext0 mx-auto mt-8 animate-spin" size={24} />}
        keyed
      >
        {(count) => (
          <Virtualizer
            count={Number(count)}
            overscan={5}
            onChange={(i) => {
              if (!i.range) return;
              if (i.range.startIndex === start && i.range.endIndex === end) return;

              start = i.range.startIndex;
              end = i.range.endIndex;
              onScroll(start, end);
            }}
          >
            {(index) => <Row item={props.items()[index]} />}
          </Virtualizer>
        )}
      </Show>
    </div>
  );
};

const Row: Component<{ item?: Item }> = (props) => {
  return (
    <Show
      when={props.item}
      fallback={
        <div class="flex w-full animate-pulse flex-row items-center gap-4 rounded-md px-4 py-2">
          <div class="bg-ctp-surface2 h-4 w-48 shrink-0 rounded" />
          <div class="bg-ctp-surface2 h-4 flex-1 rounded" />
          <div class="bg-ctp-surface2 h-4 w-12 shrink-0 rounded" />
        </div>
      }
      keyed
    >
      {(item) => (
        <div class="flex w-full flex-row items-center gap-4 rounded-md px-4 py-2">
          <span class="text-ctp-subtext0 w-48 shrink-0 text-sm tabular-nums">
            {timestampDate(item.timestamp!).toLocaleString()}
          </span>
          <span class="flex-1 truncate font-medium">{item.name}</span>
          <span class="shrink-0 font-mono tabular-nums">{item.number}</span>
        </div>
      )}
    </Show>
  );
};

export default Table;
