import type { Component } from "solid-js";
import type { AddRequest } from "$connect/number/v1/add_pb";
import { Card } from "$lib/card";
import { toast } from "$lib/toast";
import { NumberClient } from "$lib/transport";
import Form from "./Form";

type Request = Omit<AddRequest, "$typeName">;

const App: Component = () => {
	const onSubmit = async (value: Request) => {
		const [response, err] = await NumberClient.add({
			name: value.name,
			number: value.number,
		});
		if (err) {
			toast.error(`Failed to add: ${err.message}`);
			return;
		}

		toast.success(`Total: ${response.sum}`);
	};

	return (
		<div class="flex h-body flex-col items-center justify-center gap-4">
			<h1 class="font-bold text-2xl">Add Numbers</h1>
			<Card class="w-full max-w-sm">
				<Form onSubmit={onSubmit} />
			</Card>
		</div>
	);
};

export default App;
