import { timestampDate } from "@bufbuild/protobuf/wkt";
import { createVirtualizer, type Virtualizer } from "@tanstack/solid-virtual";
import type { Accessor } from "solid-js";
import { type Component, Index, Show } from "solid-js";
import type { Item } from "$connect/number/v1/list_pb";

const Table: Component<{
	count: bigint;
	items: Accessor<Item[]>;
	onScroll: (instance: Virtualizer<HTMLDivElement, Element>) => void;
}> = (props) => {
	let parentRef!: HTMLDivElement;
	const rowVirtualizer = createVirtualizer({
		count: Number(props.count),
		getScrollElement: () => parentRef,
		estimateSize: () => 35,
		overscan: 5,
		onChange: props.onScroll,
	});

	return (
		<div ref={parentRef} class="h-full w-full overflow-auto">
			<div class="sticky top-0 z-10 flex w-full flex-row items-center gap-4 border-ctp-surface0 border-b bg-ctp-base px-4 pt-4 pb-2 font-semibold text-ctp-subtext1 text-xs uppercase tracking-wide">
				<span class="w-48 shrink-0">Timestamp</span>
				<span class="flex-1">Name</span>
				<span class="shrink-0">Number</span>
			</div>
			<div
				style={{
					height: `${rowVirtualizer.getTotalSize()}px`,
					width: "100%",
					position: "relative",
				}}
			>
				<Index each={rowVirtualizer.getVirtualItems()}>
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
							<Row item={props.items()[virtualItem().index]} />
						</div>
					)}
				</Index>
			</div>
		</div>
	);
};

const Row: Component<{ item?: Item }> = (props) => {
	return (
		<Show
			when={props.item}
			fallback={
				<div class="flex w-full animate-pulse flex-row items-center gap-4 rounded-md px-4 py-2">
					<div class="h-4 w-48 shrink-0 rounded bg-ctp-surface2" />
					<div class="h-4 flex-1 rounded bg-ctp-surface2" />
					<div class="h-4 w-12 shrink-0 rounded bg-ctp-surface2" />
				</div>
			}
			keyed
		>
			{(item) => (
				<div class="flex w-full flex-row items-center gap-4 rounded-md px-4 py-2">
					<span class="w-48 shrink-0 text-ctp-subtext0 text-sm tabular-nums">
						{/** biome-ignore lint/style/noNonNullAssertion: timestamps are good */}
						{timestampDate(item.timestamp!).toLocaleString()}
					</span>
					<span class="flex-1 truncate font-medium">{item.name}</span>
					<span class="shrink-0 font-mono tabular-nums">{item.number}</span>
				</div>
			)}
		</Show>
	);
};

export default Table;
