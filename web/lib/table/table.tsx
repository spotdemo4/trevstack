import { Skeleton } from "$lib/skeleton";
import type { JSX } from "@solidjs/web";
import {
  type Component,
  createContext,
  createEffect,
  createMemo,
  createSignal,
  For,
  onCleanup,
  onSettled,
  Show,
  useContext,
} from "solid-js";
import { twMerge } from "tailwind-merge";

import styles from "./table.module.css";

const ROW_HEIGHT = 35;
const OVERSCAN = 5;

export type TableScrollMode = "element" | "window";

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
  scrollMode?: TableScrollMode;
  onScroll?: (start: number, end: number) => void;
  children?: JSX.Element;
};

type TableContextValue = {
  ref: () => HTMLDivElement | undefined;
  columns: () => string[];
  scrollMode: () => TableScrollMode;
  onScroll?: (start: number, end: number) => void;
};

type FixedRowVirtualizer = {
  items: () => number[];
  totalSize: () => number;
};

const TableContext = createContext<TableContextValue>();

const useTableContext = (componentName: string) => {
  const context = useContext(TableContext);
  if (!context) {
    throw new Error(`${componentName} must be used within Table.Table`);
  }

  return context;
};

const createFixedRowVirtualizer = (options: {
  count: () => number;
  getScrollElement: () => HTMLDivElement | undefined;
  getBodyElement: () => HTMLTableSectionElement | undefined;
  scrollMode: () => TableScrollMode;
  onRangeChange?: (start: number, end: number) => void;
}): FixedRowVirtualizer => {
  const [viewportStart, setViewportStart] = createSignal(0);
  const [viewportEnd, setViewportEnd] = createSignal(0);
  let resizeObserver: ResizeObserver | undefined;
  let animationFrame: number | undefined;

  const measure = () => {
    const body = options.getBodyElement();
    if (!body?.isConnected) return;

    const bodyRect = body.getBoundingClientRect();
    let viewportTop: number;
    let viewportBottom: number;

    if (options.scrollMode() === "window") {
      const visualViewport = window.visualViewport;
      viewportTop = visualViewport?.offsetTop ?? 0;
      viewportBottom = viewportTop + (visualViewport?.height ?? window.innerHeight);
    } else {
      const element = options.getScrollElement();
      if (!element?.isConnected) return;

      const elementRect = element.getBoundingClientRect();
      viewportTop = elementRect.top;
      viewportBottom = elementRect.bottom;
    }

    setViewportStart(viewportTop - bodyRect.top);
    setViewportEnd(viewportBottom - bodyRect.top);
  };

  const scheduleMeasure = () => {
    if (animationFrame !== undefined) return;

    animationFrame = requestAnimationFrame(() => {
      animationFrame = undefined;
      measure();
    });
  };

  onSettled(() => {
    const body = options.getBodyElement();
    if (!body) return;

    resizeObserver = new ResizeObserver(scheduleMeasure);
    resizeObserver.observe(body);

    const scrollMode = options.scrollMode();
    const element = options.getScrollElement();
    const visualViewport = window.visualViewport;

    if (scrollMode === "window") {
      window.addEventListener("scroll", scheduleMeasure, { passive: true });
      window.addEventListener("resize", scheduleMeasure);
      visualViewport?.addEventListener("scroll", scheduleMeasure, { passive: true });
      visualViewport?.addEventListener("resize", scheduleMeasure);
    } else if (element) {
      element.addEventListener("scroll", scheduleMeasure, { passive: true });
      resizeObserver.observe(element);
    }

    measure();

    return () => {
      if (scrollMode === "window") {
        window.removeEventListener("scroll", scheduleMeasure);
        window.removeEventListener("resize", scheduleMeasure);
        visualViewport?.removeEventListener("scroll", scheduleMeasure);
        visualViewport?.removeEventListener("resize", scheduleMeasure);
      } else {
        element?.removeEventListener("scroll", scheduleMeasure);
      }

      resizeObserver?.disconnect();
      if (animationFrame !== undefined) cancelAnimationFrame(animationFrame);
    };
  });

  const range = createMemo(() => {
    const count = Math.max(0, options.count());
    if (count === 0) return { start: 0, end: -1 };

    const totalSize = count * ROW_HEIGHT;
    const visibleStart = Math.max(0, viewportStart());
    const visibleEnd = Math.min(totalSize, viewportEnd());
    if (visibleEnd <= visibleStart || visibleEnd <= 0 || visibleStart >= totalSize) {
      return { start: 0, end: -1 };
    }

    const firstVisible = Math.min(count - 1, Math.floor(visibleStart / ROW_HEIGHT));
    const lastVisible = Math.min(
      count - 1,
      Math.max(firstVisible, Math.ceil(visibleEnd / ROW_HEIGHT) - 1),
    );

    return {
      start: Math.max(0, firstVisible - OVERSCAN),
      end: Math.min(count - 1, lastVisible + OVERSCAN),
    };
  });

  const items = createMemo(() => {
    const { start, end } = range();
    const visibleItems: number[] = [];

    for (let index = start; index <= end; index += 1) {
      visibleItems.push(index);
    }

    return visibleItems;
  });

  createEffect(
    () => range(),
    ({ start, end }) => {
      if (end >= start) options.onRangeChange?.(start, end);
    },
  );

  return {
    items,
    totalSize: () => Math.max(0, options.count()) * ROW_HEIGHT,
  };
};

const Table: Component<TableProps> = (props) => {
  let parentRef: HTMLDivElement | undefined;
  const scrollMode = () => props.scrollMode ?? "element";

  return (
    <TableContext
      value={{
        ref: () => parentRef,
        columns: () => props.columns,
        scrollMode,
        onScroll: props.onScroll,
      }}
    >
      <div
        ref={(node) => (parentRef = node)}
        class={twMerge(
          scrollMode() === "element" ? "h-full overflow-auto" : "overflow-visible",
          "bg-ctp-base",
          props.class,
        )}
      >
        <table class="block w-full border-separate border-spacing-0 text-ctp-text [&_td]:truncate [&_td]:px-3 [&_td]:py-2 [&_th]:px-3 [&_th]:py-2">
          {props.children}
        </table>
      </div>
    </TableContext>
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
  let bodyRef: HTMLTableSectionElement | undefined;
  let lastStart = -1;
  let lastEnd = -1;
  let onScrollTimeout: ReturnType<typeof setTimeout> | undefined;

  const notifyRange = (start: number, end: number) => {
    if (start === lastStart && end === lastEnd) return;
    lastStart = start;
    lastEnd = end;

    if (!table.onScroll) return;
    if (onScrollTimeout) clearTimeout(onScrollTimeout);
    onScrollTimeout = setTimeout(() => {
      onScrollTimeout = undefined;
      table.onScroll?.(start, end);
    }, 100);
  };

  const virtualizer = createFixedRowVirtualizer({
    count: () => props.items.length,
    getScrollElement: table.ref,
    getBodyElement: () => bodyRef,
    scrollMode: table.scrollMode,
    onRangeChange: notifyRange,
  });

  onCleanup(() => {
    if (onScrollTimeout) clearTimeout(onScrollTimeout);
  });

  return (
    <tbody
      ref={(node) => (bodyRef = node)}
      style={{
        display: "block",
        height: `${virtualizer.totalSize()}px`,
        position: "relative",
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
        <For each={virtualizer.items()}>
          {(index) => (
            <tr
              data-index={index}
              class={twMerge(
                "border-b border-ctp-surface0/60 text-sm transition-colors",
                "hover:bg-ctp-surface0/40",
                index % 2 === 0 ? "bg-ctp-base" : "bg-ctp-mantle/40",
                "[&>td]:flex [&>td]:items-center",
                styles.fadeIn,
                props.class,
              )}
              style={{
                display: "grid",
                "grid-template-columns": table.columns().join(" "),
                position: "absolute",
                height: `${ROW_HEIGHT}px`,
                transform: `translateY(${index * ROW_HEIGHT}px)`,
                width: "100%",
              }}
            >
              <Show
                when={props.items[index]}
                fallback={
                  <For each={table.columns()}>
                    {() => (
                      <td>
                        <Skeleton class="w-full" />
                      </td>
                    )}
                  </For>
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
