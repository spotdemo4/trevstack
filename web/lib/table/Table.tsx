import { Skeleton } from "$lib/skeleton";
import { Virtualizer } from "$lib/virtualizer";
import { debounce } from "@solid-primitives/scheduled";
import { CircleSlash2, LoaderCircle } from "lucide-solid";
import type { Accessor, JSX } from "solid-js";
import {
  type Component,
  createContext,
  createMemo,
  Index,
  Match,
  Show,
  Switch,
  useContext,
} from "solid-js";
import { twMerge } from "tailwind-merge";

type HeaderProps = {
  class?: string;
  children?: JSX.Element;
};

type RowsProps<T> = {
  class?: string;
  count: Accessor<bigint | undefined>;
  items: Accessor<T[]>;
  columns?: number;
  children: (item: T) => JSX.Element;
};

type TableContextValue = {
  onScroll?: (start: number, end: number) => void;
};

const TableContext = createContext<TableContextValue>();

const useTableContext = (componentName: string) => {
  const context = useContext(TableContext);
  if (!context) {
    throw new Error(`${componentName} must be used within Table.Table`);
  }

  return context;
};

const Table: Component<{
  class?: string;
  onScroll?: (start: number, end: number) => void;
  children?: JSX.Element;
}> = (props) => {
  return (
    <TableContext.Provider
      value={{
        onScroll: props.onScroll,
      }}
    >
      <div class={twMerge("flex h-full min-w-0 flex-col", props.class)}>{props.children}</div>
    </TableContext.Provider>
  );
};

const Header: Component<HeaderProps> = (props) => {
  useTableContext("Table.Header");

  return (
    <div
      class={twMerge(
        "flex w-full flex-row items-center gap-4 border-b border-ctp-surface0 bg-ctp-base px-4 pt-4 pb-2 text-xs font-semibold tracking-wide text-ctp-subtext1 uppercase *:min-w-0 [&>*:last-child]:ml-auto",
        props.class,
      )}
    >
      {props.children}
    </div>
  );
};

const Row = <T extends unknown>(props: {
  index: number;
  class?: string;
  items: Accessor<T[]>;
  columns?: number;
  children: (item: T) => JSX.Element;
}): JSX.Element => {
  const item = createMemo(() => props.items()[props.index]);
  const columns = Array.from({ length: props.columns ?? 1 }, (_, i) => i);

  return (
    <div
      class={twMerge(
        "flex w-full flex-row items-center gap-4 px-4 py-2 *:min-w-0 [&>*:last-child]:ml-auto",
        props.class,
      )}
    >
      <Show
        when={item()}
        keyed
        fallback={<Index each={columns}>{() => <Skeleton class="w-full" />}</Index>}
      >
        {(item) => props.children(item as T)}
      </Show>
    </div>
  );
};

const Rows = <T extends unknown>(props: RowsProps<T>): JSX.Element => {
  const table = useTableContext("Table.Rows");
  const count = createMemo(() => props.count());
  const onScroll = table.onScroll ? debounce(table.onScroll, 100) : undefined;

  let start = 0;
  let end = 0;

  return (
    <Show
      when={count()}
      fallback={
        <Switch>
          <Match when={count() === undefined}>
            <div class="flex w-full items-center justify-center py-10">
              <LoaderCircle class="animate-spin" size={24} />
            </div>
          </Match>
          <Match when={count() === BigInt(0)}>
            <div class="flex w-full items-center justify-center py-10">
              <CircleSlash2 size={24} />
            </div>
          </Match>
        </Switch>
      }
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
            onScroll?.(start, end);
          }}
        >
          {(index) => (
            <Row index={index} class={props.class} items={props.items} columns={props.columns}>
              {props.children}
            </Row>
          )}
        </Virtualizer>
      )}
    </Show>
  );
};

export { Table, Header, Rows };
