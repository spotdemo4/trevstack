import type { Component } from "solid-js";
import logo from "./logo.svg";

const App: Component = () => {
	return (
		<div class="flex flex-col items-center justify-center gap-4 h-screen">
			<img src={logo} class="h-28 animate-spin" alt="logo" />
			<p>
				Edit <code>src/App.tsx</code> and save to reload.
			</p>
			<a
				class="text-blue-500 hover:underline"
				href="https://github.com/solidjs/solid"
				target="_blank"
				rel="noopener noreferrer"
			>
				Learn Solid
			</a>
		</div>
	);
};

export default App;
