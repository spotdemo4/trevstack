import { type Accessor, createSignal, onCleanup, onMount } from "solid-js";

export type ChartMargin = {
  top: number;
  right: number;
  bottom: number;
  left: number;
};

export function useChartSize(
  ref: () => HTMLElement | undefined,
  fallbackHeight: number,
): { width: Accessor<number>; height: Accessor<number> } {
  const [width, setWidth] = createSignal(0);
  const [height] = createSignal(fallbackHeight);

  onMount(() => {
    const el = ref();
    if (!el) return;
    const ro = new ResizeObserver((entries) => {
      const entry = entries[0];
      if (entry) setWidth(entry.contentRect.width);
    });
    ro.observe(el);
    onCleanup(() => ro.disconnect());
  });

  return { width, height };
}

export function getChartInnerSize(
  width: number,
  height: number,
  margin: ChartMargin,
): { innerW: number; innerH: number } {
  return {
    innerW: Math.max(0, width - margin.left - margin.right),
    innerH: Math.max(0, height - margin.top - margin.bottom),
  };
}
