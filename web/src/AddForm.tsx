import type { Component } from "solid-js";
import { AddRequestSchema } from "$connect/number/v1/add_pb";
import { useAppForm } from "$lib/form/hook";
import { createSchema } from "$lib/schema";
import { toast } from "$lib/toast";
import { NumberClient } from "$lib/transport";

const schema = createSchema(AddRequestSchema);

const AddForm: Component = () => {
	const form = useAppForm(() => ({
		defaultValues: {
			name: "",
			number: 0,
		},
		validators: {
			onChange: schema,
		},
		onSubmit: async ({ value }) => {
			const [response, err] = await NumberClient.add({
				name: value.name,
				number: value.number,
			});
			if (err) {
				toast.error(`Failed to add: ${err.message}`);
				return;
			}

			toast.success(`Total: ${response.sum}`);
			form.reset();
		},
	}));

	return (
		<form.AppForm>
			<form.Form>
				<form.AppField
					name="name"
					children={(field) => <field.TextField label="Name" />}
				/>
				<form.AppField
					name="number"
					children={(field) => <field.NumberField label="Number" />}
				/>
				<form.SubmitButton />
			</form.Form>
		</form.AppForm>
	);
};

export default AddForm;
