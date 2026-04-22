import type { Component } from "solid-js";
import { GreetRequestSchema } from "$connect/greet/v1/greet_pb";
import { useAppForm } from "$lib/form/hook";
import { createSchema } from "$lib/schema";
import { toast } from "$lib/toast";
import { GreetClient } from "$lib/transport";

const schema = createSchema(GreetRequestSchema);

const GreetForm: Component = () => {
	const form = useAppForm(() => ({
		defaultValues: {
			name: "",
		},
		validators: {
			onChangeAsync: schema,
		},
		onSubmit: async ({ value }) => {
			const [response, err] = await GreetClient.greet({ name: value.name });
			if (err) {
				toast.error(`Failed to greet: ${err.message}`);
				return;
			}

			toast.success(response.greeting);
		},
	}));

	return (
		<form.AppForm>
			<form.Form>
				<form.AppField
					name="name"
					children={(field) => <field.TextField label="Name" />}
				/>
				<form.SubmitButton />
			</form.Form>
		</form.AppForm>
	);
};

export default GreetForm;
