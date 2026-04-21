import { type Component, createSignal } from "solid-js";
import { GreetClient } from "$lib/transport";
import logo from "./logo.svg";

const App: Component = () => {
	const [name, setName] = createSignal("");
	const [greeting, setGreeting] = createSignal("");

	return (
		<div class="flex flex-col items-center justify-center gap-4 h-screen">
			<img src={logo} class="h-28" alt="logo" />
			<form
				onSubmit={async (e) => {
					e.preventDefault();
					const response = await GreetClient.greet({
						name: name(),
					});
					setGreeting(response.greeting);
				}}
			>
				<input
					class="border border-gray-300 rounded px-2 py-1"
					value={name()}
					onInput={(e) => setName(e.currentTarget.value)}
				/>
				<button type="submit">Send</button>
				{greeting()}
			</form>
		</div>
	);
};

export default App;
