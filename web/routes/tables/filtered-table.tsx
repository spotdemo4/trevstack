import { Drawer } from "$lib/drawer";
import { SlidersHorizontal } from "$lib/icon";
import { createMediaQuery } from "$lib/media-query";
import { Splitter } from "$lib/splitter";
import type { TableScrollMode } from "$lib/table";
import { Portal } from "@solidjs/web";
import { type Component, Show } from "solid-js";

type FilteredTableProps = {
  /** The filter form, beside the table on wide screens and in a drawer on narrow ones. */
  filters: Component;
  table: Component<{ scrollMode?: TableScrollMode }>;
};

/**
 * A table with its filters. On wide screens it fills its parent's height, with the filters in a
 * resizable panel and the table scrolling on its own; on narrow ones the page scrolls the table,
 * and a floating button opens the filters.
 */
export const FilteredTable: Component<FilteredTableProps> = (props) => {
  const isDesktop = createMediaQuery("(min-width: 900px)");

  return (
    <Show
      when={isDesktop()}
      fallback={
        <div class="pb-[calc(3rem+env(safe-area-inset-bottom))]">
          <props.table scrollMode="window" />
          <Drawer.Root>
            {/* Portaled out of the tab panel, whose entry animation's transform would
                otherwise make it the button's containing block for a moment. */}
            <Portal>
              <Drawer.Trigger class="fixed right-[calc(1rem+env(safe-area-inset-right))] bottom-[calc(2rem+env(safe-area-inset-bottom))] z-20 m-0 shadow-lg">
                <SlidersHorizontal size={16} /> Filters
              </Drawer.Trigger>
            </Portal>
            <Drawer.Content>
              <div class="flex items-center justify-between">
                <Drawer.Title>Filters</Drawer.Title>
                <Drawer.CloseTrigger />
              </div>
              <props.filters />
            </Drawer.Content>
          </Drawer.Root>
        </div>
      }
    >
      <Splitter.Root
        class="h-full bg-surface"
        defaultSize={[15, 50]}
        panels={[{ id: "a", minSize: 20, maxSize: 50 }, { id: "b" }]}
      >
        <Splitter.Panel id="a" class="overflow-y-auto bg-surface p-4">
          <props.filters />
        </Splitter.Panel>
        <Splitter.ResizeTrigger id="a:b" />
        <Splitter.Panel id="b" class="min-h-0">
          <props.table />
        </Splitter.Panel>
      </Splitter.Root>
    </Show>
  );
};
