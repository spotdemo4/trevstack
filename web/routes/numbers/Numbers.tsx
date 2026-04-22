import { createAsync } from "@solidjs/router";
import { type Component, Suspense } from "solid-js";
import { NumberClient } from "$lib/transport";
import NumbersTable from "./NumbersTable";

const App: Component = () => {
	const numbers = createAsync(async () => {
		const [response, err] = await NumberClient.list({});
		if (err) {
			console.error("Failed to fetch numbers:", err);
			return [];
		}

		return response.items;
	});

	return (
		<div class="flex h-full flex-col items-center justify-center gap-4">
			<h1 class="font-bold text-2xl">Numbers</h1>
			<Suspense fallback={<p>Loading...</p>}>
				<NumbersTable items={numbers() ?? []} />
			</Suspense>
		</div>
	);
};

export default App;
