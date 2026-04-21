import { Button } from "@kobalte/core/button";
import { TextField } from "@kobalte/core/text-field";
import { type Component, createSignal, Match, Switch } from "solid-js";
import { GreetClient } from "$lib/transport";
import logo from "./logo.svg";

const App: Component = () => {
	const [name, setName] = createSignal("");
	const [greeting, setGreeting] = createSignal("");
	const [error, setError] = createSignal("");

	return (
		<div class="flex flex-col items-center justify-center gap-4 h-screen bg-ctp-base text-ctp-text">
			<img src={logo} class="h-28" alt="logo" />
			<Switch>
				<Match when={!greeting()}>
					<form
						class="flex flex-col justify-center gap-4"
						onSubmit={async (e) => {
							e.preventDefault();

							const [response, error] = await GreetClient.greet({
								name: name(),
							});
							if (error) {
								setError(error.rawMessage);
								return;
							}

							setGreeting(response.greeting);
						}}
					>
						<TextField
							value={name()}
							onChange={setName}
							validationState={error() ? "invalid" : "valid"}
							class="flex flex-col gap-1"
						>
							<TextField.Label>Name</TextField.Label>
							<TextField.Input class="px-3 py-2 rounded border border-ctp-overlay0 bg-ctp-surface text-ctp-text focus:outline-none focus:ring-2 focus:ring-ctp-sky focus:ring-offset-2 focus:ring-offset-ctp-base transition-all duration-200 hover:border-ctp-sky-400" />
							<TextField.ErrorMessage>{error()}</TextField.ErrorMessage>
						</TextField>
						<Button
							class="px-4 py-2 rounded bg-ctp-sky text-ctp-base hover:bg-ctp-sky-400 cursor-pointer transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
							type="submit"
						>
							Send
						</Button>
					</form>
				</Match>
				<Match when={greeting()}>{greeting()}</Match>
			</Switch>
		</div>
	);
};

export default App;
