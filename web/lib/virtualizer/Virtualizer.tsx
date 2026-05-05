import { createVirtualizer, type VirtualizerOptions } from "@tanstack/solid-virtual";
import type { JSX } from "solid-js";
import { type Component, Index } from "solid-js";
import { twMerge } from "tailwind-merge";

type VirtualizerProps = {
  class?: string;
  count: number;
  overscan?: number;
  size?: number;
  onChange?: VirtualizerOptions<HTMLDivElement, Element>["onChange"];
  children: (index: number) => JSX.Element;
};

export const Virtualizer: Component<VirtualizerProps> = (props) => {
  // oxlint-disable-next-line no-unassigned-vars
  let parentRef!: HTMLDivElement;

  const virtualizer = createVirtualizer({
    count: props.count,
    overscan: props.overscan ?? 5,
    estimateSize: () => props.size ?? 35,
    getScrollElement: () => parentRef,
    onChange: props?.onChange,
  });

  return (
    <>
      {/* The scrollable element for your list */}
      <div ref={parentRef} class={twMerge("overflow-auto", props.class)}>
        {/* The large inner element to hold all of the items */}
        <div
          style={{
            height: `${virtualizer.getTotalSize()}px`,
            width: "100%",
            position: "relative",
          }}
        >
          {/* Only the visible items in the virtualizer, manually positioned to be in view */}
          <Index each={virtualizer.getVirtualItems()}>
            {(virtualItem) => (
              <div
                style={{
                  position: "absolute",
                  top: 0,
                  left: 0,
                  width: "100%",
                  height: `${virtualItem().size}px`,
                  transform: `translateY(${virtualItem().start}px)`,
                }}
              >
                {props.children(virtualItem().index)}
              </div>
            )}
          </Index>
        </div>
      </div>
    </>
  );
};
