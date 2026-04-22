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
			class="flex flex-col gap-1.5"
		>
			<Show when={props.label}>
				<TextFieldPrimative.Label class="font-medium text-ctp-subtext1 text-sm data-invalid:text-ctp-red">
					{props.label}
				</TextFieldPrimative.Label>
			</Show>
			<TextFieldPrimative.Input class="rounded-md border border-ctp-surface1 bg-ctp-base px-3 py-2 text-ctp-text text-sm transition-colors placeholder:text-ctp-overlay0 hover:border-ctp-surface2 focus:border-ctp-sky focus:outline-none focus:ring-2 focus:ring-ctp-sky/40 data-invalid:border-ctp-red data-invalid:focus:ring-ctp-red/40" />
			<TextFieldPrimative.ErrorMessage class="text-ctp-red text-xs">
				{field()
					.state.meta.errors.map((err) => err.message)
					.join(",")}
			</TextFieldPrimative.ErrorMessage>
		</TextFieldPrimative>
	);
}
