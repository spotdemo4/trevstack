import type { Component } from "solid-js";
import AddForm from "./AddForm";

const App: Component = () => {
	return (
		<div class="flex h-full flex-col items-center justify-center gap-4">
			<h1 class="font-bold text-2xl">Add Numbers</h1>
			<AddForm />
		</div>
	);
};

export default App;
