import { TextField as TextFieldPrimative } from "@kobalte/core/text-field";
import { For, Show } from "solid-js";
import { useFieldContext } from "./context";

export function TextField(props: { label?: string }) {
	const field = useFieldContext<string>();
	const name = field().name;

	return (
		<TextFieldPrimative
			name={name}
			value={field().state.value ?? ""}
			onChange={(c) => {
				if (c === "" && field().state.value !== undefined) {
					field().form.resetField(name);
					field().form.validateField(name, "change");
				} else {
					field().handleChange(c);
				}
			}}
			validationState={
				field().state.meta.isValid || !field().state.meta.isBlurred
					? "valid"
					: "invalid"
			}
			class="flex flex-col gap-1.5"
		>
			<Show when={props.label}>
				<TextFieldPrimative.Label class="font-medium text-ctp-subtext1 text-sm data-invalid:text-ctp-red">
					{props.label}
				</TextFieldPrimative.Label>
			</Show>
			<TextFieldPrimative.Input
				onBlur={field().handleBlur}
				class="rounded-md border border-ctp-surface1 bg-ctp-base px-3 py-2 text-ctp-text text-sm transition-colors placeholder:text-ctp-overlay0 hover:border-ctp-surface2 focus:border-ctp-sky focus:outline-none focus:ring-2 focus:ring-ctp-sky/40 data-invalid:border-ctp-red data-invalid:focus:ring-ctp-red/40"
			/>
			<For each={field().state.meta.errors}>
				{(err) => (
					<TextFieldPrimative.ErrorMessage class="text-ctp-red text-xs">
						{err.message}
					</TextFieldPrimative.ErrorMessage>
				)}
			</For>
		</TextFieldPrimative>
	);
}
