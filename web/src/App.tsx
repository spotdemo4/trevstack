import { Toast } from "@kobalte/core";
import type { Component } from "solid-js";
import { Portal } from "solid-js/web";
import AddForm from "./AddForm";
import logo from "./logo.svg";

const App: Component = () => {
	return (
		<>
			<div class="flex h-screen flex-col items-center justify-center gap-4 text-ctp-text">
				<img src={logo} class="h-28" alt="logo" />
				<AddForm />
			</div>
			<Portal>
				<Toast.Region>
					<Toast.List class="fixed top-4 right-4 z-50 flex w-96 max-w-[calc(100vw-2rem)] flex-col gap-2 outline-none" />
				</Toast.Region>
			</Portal>
		</>
	);
};

export default App;
