import { debounce } from "@solid-primitives/scheduled";
import { createVirtualizer } from "@tanstack/solid-virtual";
import { CircleSlash2, LoaderCircle } from "lucide-solid";
import type { Accessor, JSX } from "solid-js";
import {
  type Component,
  createContext,
  createEffect,
  createMemo,
  createSignal,
  Index,
  Match,
  onCleanup,
  Show,
  Switch,
  useContext,
} from "solid-js";
import { twMerge } from "tailwind-merge";

type HeaderProps = {
  class?: string;
  children?: JSX.Element;
};

type BodyProps<T> = {
  class?: string;
  count: Accessor<bigint | undefined>;
  items: Accessor<T[]>;
  children: (item: T) => JSX.Element;
};

type RowsProps<T> = {
  class?: string;
  count: bigint;
  items: Accessor<T[]>;
  children: (item: T) => JSX.Element;
};

type TableProps = {
  class?: string;
  /**
   * CSS `grid-template-columns` value applied to header and body rows so that
   * every `<th>` and `<td>` auto-aligns. Example: `"200px 1fr 120px"` or
   * `"auto auto auto"`.
   */
  columns: string;
  onScroll?: (start: number, end: number) => void;
  children?: JSX.Element;
};

type TableContextValue = {
  ref: () => HTMLDivElement | undefined;
  columns: () => string;
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

const LOADING_STATE_DELAY_MS = 150;

const Table: Component<TableProps> = (props) => {
  let parentRef: HTMLDivElement | undefined;

  return (
    <TableContext.Provider
      value={{
        ref: () => parentRef,
        columns: () => props.columns,
        onScroll: props.onScroll,
      }}
    >
      <div
        ref={(node) => (parentRef = node)}
        class={twMerge(
          "h-full overflow-auto rounded-md border border-ctp-surface0 bg-ctp-base",
          props.class,
        )}
      >
        <table class="w-full border-separate border-spacing-0 text-ctp-text [&_td]:truncate [&_td]:px-3 [&_td]:py-2 [&_th]:px-3 [&_th]:py-2">
          {props.children}
        </table>
      </div>
    </TableContext.Provider>
  );
};

const Header: Component<HeaderProps> = (props) => {
  const table = useTableContext("Table.Header");

  return (
    <thead class="sticky top-0 z-10 bg-ctp-mantle/95 backdrop-blur supports-backdrop-filter:bg-ctp-mantle/75">
      <tr
        class={twMerge(
          "text-left text-xs font-semibold tracking-wider text-ctp-subtext1 uppercase",
          "[&>th]:border-b [&>th]:border-ctp-surface0",
          props.class,
        )}
        style={{ display: "grid", "grid-template-columns": table.columns() }}
      >
        {props.children}
      </tr>
    </thead>
  );
};

const Rows = <T extends unknown>(props: RowsProps<T>): JSX.Element => {
  const table = useTableContext("Table.Rows");
  const onScroll = table.onScroll ? debounce(table.onScroll, 100) : undefined;

  let start = 0;
  let end = 0;

  const virtualizer = createVirtualizer({
    count: Number(props.count),
    overscan: 5,
    estimateSize: () => 35,
    getScrollElement: () => table.ref() ?? null,
    onChange: (i) => {
      if (!i.range) return;
      if (i.range.startIndex === start && i.range.endIndex === end) return;
      start = i.range.startIndex;
      end = i.range.endIndex;
      onScroll?.(start, end);
    },
  });

  return (
    <>
      {/* The large inner element to hold all of the items */}
      <tbody
        style={{
          height: `${virtualizer.getTotalSize()}px`, //tells scrollbar how big the table is
          position: "relative", //needed for absolute positioning of rows
        }}
      >
        {/* Only the visible items in the virtualizer, manually positioned to be in view */}
        <Index each={virtualizer.getVirtualItems()}>
          {(virtualItem) => (
            <tr
              data-index={virtualItem().index}
              class={twMerge(
                "border-b border-ctp-surface0/60 text-sm transition-colors",
                "hover:bg-ctp-surface0/40",
                "odd:bg-ctp-base even:bg-ctp-mantle/40",
                "[&>td]:flex [&>td]:items-center",
                props.class,
              )}
              style={{
                display: "grid",
                "grid-template-columns": table.columns(),
                position: "absolute",
                height: `${virtualItem().size}px`,
                transform: `translateY(${virtualItem().start}px)`, //this should always be a `style` as it changes on scroll
                width: "100%",
              }}
            >
              <Show
                when={props.items()[virtualItem().index]}
                fallback={<td class="text-ctp-overlay0">Loading...</td>}
                keyed
              >
                {(item) => props.children(item as T)}
              </Show>
            </tr>
          )}
        </Index>
      </tbody>
    </>
  );
};

const Body = <T extends unknown>(props: BodyProps<T>): JSX.Element => {
  useTableContext("Table.Body");
  const count = createMemo(() => props.count());
  const [showLoadingState, setShowLoadingState] = createSignal(false);

  createEffect(() => {
    if (count() !== undefined) {
      setShowLoadingState(false);
      return;
    }

    const timeoutId = setTimeout(() => {
      setShowLoadingState(true);
    }, LOADING_STATE_DELAY_MS);

    onCleanup(() => clearTimeout(timeoutId));
  });

  return (
    <Show
      when={count()}
      fallback={
        <Switch>
          <Match when={count() === undefined && showLoadingState()}>
            <div class="flex w-full animate-fade-in items-center justify-center gap-2 py-10 text-ctp-subtext0 motion-reduce:animate-none">
              <LoaderCircle class="animate-spin" size={20} />
              <span class="text-sm">Loading…</span>
            </div>
          </Match>
          <Match when={count() === BigInt(0)}>
            <div class="flex w-full animate-fade-in flex-col items-center justify-center gap-2 py-10 text-ctp-overlay1 motion-reduce:animate-none">
              <CircleSlash2 size={24} />
              <span class="text-sm">No results</span>
            </div>
          </Match>
        </Switch>
      }
      keyed
    >
      {(count) => (
        <Rows count={count} items={props.items} class={props.class}>
          {props.children}
        </Rows>
      )}
    </Show>
  );
};

export { Table, Header, Body };
