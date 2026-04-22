import { timestampDate } from "@bufbuild/protobuf/wkt";
import { createVirtualizer, type Virtualizer } from "@tanstack/solid-virtual";
import { type Component, createSignal, Show } from "solid-js";
import type { Item } from "$connect/number/v1/list_pb";
import { NumberClient } from "$lib/transport";

const NumbersTable: Component<{
	count: bigint;
	cursor: bigint;
	items: Item[];
}> = (props) => {
	let cursor = props.cursor;
	const [items, setItems] = createSignal(props.items);

	let isUpdating = false;
	const onChange = async (instance: Virtualizer<HTMLDivElement, Element>) => {
		if (!instance.range || isUpdating) {
			return;
		}
		isUpdating = true;

		while (instance.range.endIndex > items().length - 1) {
			const [response, err] = await NumberClient.list({
				cursor: cursor,
			});
			if (err) {
				console.error("Failed to fetch numbers:", err);
			} else {
				setItems((prev) => [...prev, ...(response.items ?? [])]);
				cursor = response.nextCursor;
			}
		}

		isUpdating = false;
	};

	let parentRef!: HTMLDivElement;
	const rowVirtualizer = createVirtualizer({
		count: Number(props.count),
		getScrollElement: () => parentRef,
		estimateSize: () => 35,
		overscan: 5,
		onChange: onChange,
	});

	return (
		<div ref={parentRef} class="h-full w-full overflow-auto">
			<div class="sticky top-0 z-10 flex w-full flex-row items-center gap-4 border-ctp-surface0 border-b bg-ctp-base px-4 py-2 font-semibold text-ctp-subtext1 text-xs uppercase tracking-wide">
				<span class="w-48 shrink-0">Timestamp</span>
				<span class="flex-1">Name</span>
				<span class="shrink-0">Number</span>
			</div>
			{/* The large inner element to hold all of the items */}
			<div
				style={{
					height: `${rowVirtualizer.getTotalSize()}px`,
					width: "100%",
					position: "relative",
				}}
			>
				{/* Only the visible items in the virtualizer, manually positioned to be in view */}
				{rowVirtualizer.getVirtualItems().map((virtualItem) => (
					<div
						// key={virtualItem.key}
						style={{
							position: "absolute",
							top: 0,
							left: 0,
							width: "100%",
							height: `${virtualItem.size}px`,
							transform: `translateY(${virtualItem.start}px)`,
						}}
					>
						<NumbersTableRow item={items()[virtualItem.index]} />
					</div>
				))}
			</div>
		</div>
	);
};

const NumbersTableRow: Component<{ item?: Item }> = (props) => {
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

export default NumbersTable;
