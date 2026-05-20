import { Show, type Component } from "solid-js";
import { twMerge } from "tailwind-merge";

import styles from "./chart-motion.module.css";

type ChartFrameProps = {
  width: number;
  height: number;
  isEmpty: boolean;
  containerRef: (element: HTMLDivElement) => void;
  svgRef: (element: SVGSVGElement) => void;
  tooltipRef: (element: HTMLDivElement) => void;
  emptyLabel?: string;
  class?: string;
};

const tooltipClass =
  "pointer-events-none absolute z-10 max-w-56 rounded-md border border-ctp-surface1 bg-ctp-base/95 px-2 py-1 text-xs font-medium whitespace-pre text-ctp-text opacity-0 shadow-lg transition-opacity";

export const ChartFrame: Component<ChartFrameProps> = (props) => {
  return (
    <div ref={props.containerRef} class={twMerge("relative w-full", props.class)}>
      <svg
        ref={props.svgRef}
        width={props.width}
        height={props.height}
        class={`${styles.ChartCanvas} block`}
      />
      <div ref={props.tooltipRef} class={tooltipClass} />
      <Show when={props.isEmpty}>
        <div
          class={`${styles.EmptyState} absolute inset-0 flex items-center justify-center text-sm text-ctp-subtext0`}
        >
          {props.emptyLabel ?? "No data"}
        </div>
      </Show>
    </div>
  );
};
