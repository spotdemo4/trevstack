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
		<div ref={parentRef} class="h-96 w-full overflow-auto">
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
		<Show when={props.item} fallback={<div>Loading...</div>} keyed>
			{(item) => (
				<div class="flex w-full flex-row items-center justify-between rounded-md bg-ctp-surface1 px-4 py-2">
					{timestampDate(item.timestamp!).toLocaleString()}
				</div>
			)}
		</Show>
	);
};

export default NumbersTable;
