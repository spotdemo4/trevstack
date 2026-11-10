import type { JSX } from "@solidjs/web";
import { type BaseType, select, type Selection } from "d3-selection";
import { createEffect, createSignal, onSettled, Show } from "solid-js";
import { twMerge } from "tailwind-merge";

import { hideChartTooltip, showChartTooltip } from "./chart-tooltip";

import styles from "./chart-motion.module.css";

export type ChartMargin = {
  top: number;
  right: number;
  bottom: number;
  left: number;
};

export type ChartTooltip = {
  /** Shows `text(d)` next to the pointer while it is over an element of `selection`. */
  attach: <E extends BaseType, D, P extends BaseType, PD>(
    selection: Selection<E, D, P, PD>,
    text: (d: D) => string,
  ) => void;
};

export type ChartDrawContext<T> = {
  /** Group translated by the chart margin; draw everything inside it. */
  g: Selection<SVGGElement, unknown, null, undefined>;
  /** Width inside the margin. */
  width: number;
  /** Height inside the margin. */
  height: number;
  data: T[];
  tooltip: ChartTooltip;
};

type ChartProps<T> = {
  data: T[];
  height: number;
  margin: ChartMargin;
  draw: (ctx: ChartDrawContext<T>) => void;
  emptyLabel?: string;
  class?: string;
};

const tooltipClass =
  "pointer-events-none absolute z-10 max-w-56 rounded-md border border-line bg-background/95 px-2 py-1 text-xs font-medium whitespace-pre text-foreground opacity-0 shadow-lg transition-opacity";

/**
 * Responsive SVG chart: tracks its own width, redraws with `draw` when the size
 * or data change, and owns the tooltip and empty state.
 */
export function Chart<T>(props: ChartProps<T>): JSX.Element {
  let container!: HTMLDivElement;
  let svg!: SVGSVGElement;
  let tooltipEl!: HTMLDivElement;

  const [width, setWidth] = createSignal(0);

  onSettled(() => {
    const observer = new ResizeObserver(([entry]) => {
      if (entry) setWidth(entry.contentRect.width);
    });
    observer.observe(container);
    return () => observer.disconnect();
  });

  const tooltip: ChartTooltip = {
    attach: (selection, text) => {
      selection
        .on("pointerenter.tooltip pointermove.tooltip", (event, d) => {
          showChartTooltip({
            event: event as PointerEvent,
            container,
            tooltip: tooltipEl,
            text: text(d),
          });
        })
        .on("pointerleave.tooltip", () => hideChartTooltip(tooltipEl));
    },
  };

  createEffect(
    () => ({
      width: width(),
      height: props.height,
      margin: props.margin,
      data: props.data,
      draw: props.draw,
    }),
    ({ width, height, margin, data, draw }) => {
      const root = select(svg);
      const clear = () => {
        root.selectAll("*").remove();
        hideChartTooltip(tooltipEl);
      };
      clear();
      if (width === 0 || data.length === 0) return clear;

      draw({
        g: root.append("g").attr("transform", `translate(${margin.left},${margin.top})`),
        width: Math.max(0, width - margin.left - margin.right),
        height: Math.max(0, height - margin.top - margin.bottom),
        data,
        tooltip,
      });
      return clear;
    },
  );

  return (
    <div ref={(el) => (container = el)} class={twMerge("relative w-full", props.class)}>
      <svg
        ref={(el) => (svg = el)}
        width={width()}
        height={props.height}
        class={`${styles.ChartCanvas} block`}
      />
      <div ref={(el) => (tooltipEl = el)} class={tooltipClass} />
      <Show when={props.data.length === 0}>
        <div
          class={`${styles.EmptyState} absolute inset-0 flex items-center justify-center text-sm text-foreground-subtle`}
        >
          {props.emptyLabel ?? "No data"}
        </div>
      </Show>
    </div>
  );
}
