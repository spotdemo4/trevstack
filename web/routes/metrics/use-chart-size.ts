import { type Accessor, createSignal, onCleanup, onMount } from "solid-js";

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
