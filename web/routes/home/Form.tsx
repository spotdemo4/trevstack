import type { Component } from "solid-js";
import { type AddRequest, AddRequestSchema } from "$connect/number/v1/add_pb";
import { useAppForm } from "$lib/form/hook";
import { createSchema } from "$lib/schema";

type Request = Omit<AddRequest, "$typeName">;
const schema = createSchema(AddRequestSchema);

const AddForm: Component<{
	onSubmit: (value: Request) => void;
}> = (props) => {
	const form = useAppForm(() => ({
		defaultValues: {
			name: "",
		} as Request,
		validators: {
			onChange: schema,
		},
		onSubmit: async ({ value }) => {
			props.onSubmit(value);
			form.reset();
		},
	}));

	return (
		<form.AppForm>
			<form.Form class="max-w-sm flex-col">
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
