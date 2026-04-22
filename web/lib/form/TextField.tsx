import { TextField as TextFieldPrimative } from "@kobalte/core/text-field";
import { Show } from "solid-js";
import { useFieldContext } from "./context";

export function TextField(props: { label?: string }) {
	const field = useFieldContext<string>();

	return (
		<TextFieldPrimative
			name={field().name}
			value={field().state.value}
			onChange={field().handleChange}
			onBlur={field().handleBlur}
			validationState={
				field().state.meta.isTouched && !field().state.meta.isValid
					? "invalid"
					: "valid"
			}
		>
			<Show when={props.label}>
				<TextFieldPrimative.Label>{props.label}</TextFieldPrimative.Label>
			</Show>
			<TextFieldPrimative.Input />
			<TextFieldPrimative.ErrorMessage>
				{field()
					.state.meta.errors.map((err) => err.message)
					.join(",")}
			</TextFieldPrimative.ErrorMessage>
		</TextFieldPrimative>
	);
}
