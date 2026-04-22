import { createAsync } from "@solidjs/router";
import { type Component, Show, Suspense } from "solid-js";
import { NumberClient } from "$lib/transport";
import NumbersTable from "./NumbersTable";

const App: Component = () => {
	const numbers = createAsync(async () => {
		const [response, err] = await NumberClient.list({});
		if (err) {
			console.error("Failed to fetch numbers:", err);
			return;
		}

		return response;
	});

	return (
		<div class="flex h-full flex-col items-center justify-center gap-4">
			<h1 class="font-bold text-2xl">Numbers</h1>
			<Suspense fallback={<span>Loading account stats...</span>}>
				<Show when={numbers()} fallback={<span>No numbers found.</span>} keyed>
					{(resp) => (
						<NumbersTable
							count={resp.totalCount}
							cursor={resp.nextCursor}
							items={resp.items}
						/>
					)}
				</Show>
			</Suspense>
		</div>
	);
};

export default App;
