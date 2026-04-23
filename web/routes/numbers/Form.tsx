import type { Component } from "solid-js";
import {
	type ListRequest,
	ListRequestSchema,
} from "$connect/number/v1/list_pb";
import { useAppForm } from "$lib/form/hook";
import { createSchema } from "$lib/schema";

type Request = Omit<ListRequest, "$typeName">;
const requestSchema = createSchema(ListRequestSchema);

const Form: Component<{
	onSubmit: (value: Request) => void;
}> = (props) => {
	const form = useAppForm(() => ({
		defaultValues: {} as Request,
		validators: {
			onChange: requestSchema,
		},
		onSubmit: async ({ value }) => {
			props.onSubmit(value);
		},
	}));

	return (
		<form.AppForm>
			<form.Form class="justify-center">
				<form.AppField
					name="name"
					children={(field) => <field.TextField label="Name" />}
				/>
				<form.AppField
					name="min"
					children={(field) => <field.NumberField label="Min" />}
				/>
				<form.AppField
					name="max"
					children={(field) => <field.NumberField label="Max" />}
				/>
				<form.SubmitButton />
			</form.Form>
		</form.AppForm>
	);
};

export default Form;
