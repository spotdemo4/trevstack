import { timestampDate } from "@bufbuild/protobuf/wkt";
import { createVirtualizer } from "@tanstack/solid-virtual";
import type { Component } from "solid-js";
import type { Item } from "$connect/number/v1/list_pb";

const NumbersTable: Component<{ items: Item[] }> = (props) => {
	// const [sorting, setSorting] = createSignal<SortingState>([]);

	let ref!: HTMLDivElement;

	const rowVirtualizer = createVirtualizer({
		count: 10000,
		getScrollElement: () => ref,
		estimateSize: () => 35,
	});

	return (
		<>
			{/* The scrollable element for your list */}
			<div
				ref={ref}
				style={{
					height: `400px`,
					overflow: "auto", // Make it scroll!
				}}
			>
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
							Row {virtualItem.index}
						</div>
					))}
				</div>
			</div>
		</>
	);
};

export default NumbersTable;
