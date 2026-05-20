import { Skeleton } from "$lib/skeleton";
import { debounce } from "@solid-primitives/scheduled";
import { createVirtualizer } from "@tanstack/solid-virtual";
import type { JSX } from "solid-js";
import { type Component, createContext, For, Index, Show, useContext } from "solid-js";
import { twMerge } from "tailwind-merge";

import styles from "./table.module.css";

type HeaderProps = {
  class?: string;
  children?: JSX.Element;
};

type BodyProps<T> = {
  class?: string;
  items: T[];
  loading?: boolean;
  emptyMessage?: JSX.Element;
  children: (item: T) => JSX.Element;
};

type TableProps = {
  class?: string;
  /**
   * Column track sizes applied to header and body rows so that every `<th>`
   * and `<td>` auto-aligns. Each entry is a `grid-template-columns` track
   * value. Example: `["200px", "1fr", "120px"]`.
   */
  columns: string[];
  onScroll?: (start: number, end: number) => void;
  children?: JSX.Element;
};

type TableContextValue = {
  ref: () => HTMLDivElement | undefined;
  columns: () => string[];
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
        class={twMerge("h-full overflow-auto bg-ctp-base", props.class)}
      >
        <table class="block w-full border-separate border-spacing-0 text-ctp-text [&_td]:truncate [&_td]:px-3 [&_td]:py-2 [&_th]:px-3 [&_th]:py-2">
          {props.children}
        </table>
      </div>
    </TableContext.Provider>
  );
};

const Header: Component<HeaderProps> = (props) => {
  const table = useTableContext("Table.Header");

  return (
    <thead class="sticky top-0 z-10 block bg-ctp-mantle/95 backdrop-blur supports-backdrop-filter:bg-ctp-mantle/75">
      <tr
        class={twMerge(
          "text-left text-xs font-semibold tracking-wider text-ctp-subtext1 uppercase",
          "[&>th]:border-b [&>th]:border-ctp-surface0",
          props.class,
        )}
        style={{ display: "grid", "grid-template-columns": table.columns().join(" ") }}
      >
        {props.children}
      </tr>
    </thead>
  );
};

const Body = <T extends unknown>(props: BodyProps<T>): JSX.Element => {
  const table = useTableContext("Table.Rows");
  const onScroll = table.onScroll ? debounce(table.onScroll, 100) : undefined;

  let start = 0;
  let end = 0;

  const virtualizer = createVirtualizer({
    // https://github.com/TanStack/virtual/issues/661#issuecomment-1937805648
    get count() {
      return props.items.length;
    },
    // https://github.com/TanStack/virtual/issues/1011#issuecomment-3677935028
    getScrollElement: () => (table.ref()?.isConnected ? table.ref()! : null),
    overscan: 5,
    estimateSize: () => 35,
    onChange: (i) => {
      if (!i.range) return;
      if (i.range.startIndex === start && i.range.endIndex === end) return;
      start = i.range.startIndex;
      end = i.range.endIndex;
      onScroll?.(start, end);
    },
  });

  return (
    <tbody
      style={{
        display: "block",
        height: `${virtualizer.getTotalSize()}px`, //tells scrollbar how big the table is
        position: "relative", //needed for absolute positioning of rows
      }}
    >
      <Show
        when={props.loading || props.items.length > 0}
        fallback={
          <tr class={twMerge("text-sm", styles.emptyState, props.class)}>
            <td colspan={table.columns().length} class="px-3 text-center text-ctp-subtext0">
              {props.emptyMessage ?? "No items found."}
            </td>
          </tr>
        }
      >
        {/* Only the visible items in the virtualizer, manually positioned to be in view */}
        <For each={virtualizer.getVirtualItems()}>
          {(virtualItem) => (
            <tr
              data-index={virtualItem.index}
              class={twMerge(
                "border-b border-ctp-surface0/60 text-sm transition-colors",
                "hover:bg-ctp-surface0/40",
                virtualItem.index % 2 === 0 ? "bg-ctp-base" : "bg-ctp-mantle/40",
                "[&>td]:flex [&>td]:items-center",
                styles.fadeIn,
                props.class,
              )}
              style={{
                display: "grid",
                "grid-template-columns": table.columns().join(" "),
                position: "absolute",
                height: `${virtualItem.size}px`,
                transform: `translateY(${virtualItem.start}px)`, //this should always be a `style` as it changes on scroll
                width: "100%",
              }}
            >
              <Show
                when={props.items[virtualItem.index]}
                fallback={
                  <Index each={table.columns()}>
                    {() => (
                      <td>
                        <Skeleton class="w-full" />
                      </td>
                    )}
                  </Index>
                }
                keyed
              >
                {(item) => props.children(item as T)}
              </Show>
            </tr>
          )}
        </For>
      </Show>
    </tbody>
  );
};

export { Table, Header, Body };
