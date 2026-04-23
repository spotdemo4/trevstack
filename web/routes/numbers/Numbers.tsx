import { createAsync } from "@solidjs/router";
import { LoaderCircle } from "lucide-solid";
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
		<div class="flex h-body flex-col items-center gap-4 pt-4">
			<h1 class="font-bold text-2xl">Numbers</h1>
			<Suspense fallback={<LoaderCircle class="animate-spin" />}>
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
