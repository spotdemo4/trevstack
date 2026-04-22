import { createFormHook } from "@tanstack/solid-form";
import { fieldContext, formContext } from "./context";
import { Form } from "./Form";
import { NumberField } from "./NumberField";
import { SubmitButton } from "./SubmitButton";
import { TextField } from "./TextField";

export const { useAppForm, withForm } = createFormHook({
	fieldComponents: {
		TextField,
		NumberField,
	},
	formComponents: {
		Form,
		SubmitButton,
	},
	fieldContext,
	formContext,
});
