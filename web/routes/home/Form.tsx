import { create } from "@bufbuild/protobuf";
import { createStandardSchema } from "@bufbuild/protovalidate";
import type { Component } from "solid-js";
import { type AddRequest, AddRequestSchema } from "$connect/number/v1/add_pb";
import { useForm } from "$lib/form/hook";

const AddForm: Component<{
	onSubmit: (value: AddRequest) => void;
}> = (props) => {
	const form = useForm(() => ({
		defaultValues: create(AddRequestSchema),
		validators: {
			onMount: createStandardSchema(AddRequestSchema),
			onChange: createStandardSchema(AddRequestSchema),
		},
		onSubmit: async ({ value }) => {
			props.onSubmit(value);
			form.reset();
		},
	}));

	return (
		<form.AppForm>
			<form.Form class="max-w-sm flex-col">
				<form.AppField name="name">
					{(field) => <field.TextField label="Name" />}
				</form.AppField>
				<form.AppField name="number">
					{(field) => <field.NumberField label="Number" />}
				</form.AppField>
				<form.SubmitButton />
			</form.Form>
		</form.AppForm>
	);
};

export default AddForm;
