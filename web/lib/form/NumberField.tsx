import { NumberField as NumberFieldPrimative } from "@kobalte/core/number-field";
import { ChevronDown, ChevronUp } from "lucide-solid";
import { Show } from "solid-js";
import { useFieldContext } from "./context";

export function NumberField(props: { label?: string }) {
	const field = useFieldContext<number>();
	const name = field().name;

	return (
		<NumberFieldPrimative
			name={name}
			rawValue={field().state.value ?? undefined}
			onRawValueChange={(c) => {
				if (Number.isNaN(c)) {
					field().form.deleteField(name);
				} else {
					field().handleChange(c);
				}
			}}
			onBlur={field().handleBlur}
			validationState={
				field().state.meta.isTouched && !field().state.meta.isValid
					? "invalid"
					: "valid"
			}
			class="flex flex-col gap-1.5"
		>
			<Show when={props.label}>
				<NumberFieldPrimative.Label class="font-medium text-ctp-subtext1 text-sm data-invalid:text-ctp-red">
					{props.label}
				</NumberFieldPrimative.Label>
			</Show>
			<div class="relative">
				<NumberFieldPrimative.Input class="w-full rounded-md border border-ctp-surface1 bg-ctp-base px-3 py-2 text-ctp-text text-sm transition-colors placeholder:text-ctp-overlay0 hover:border-ctp-surface2 focus:border-ctp-sky focus:outline-none focus:ring-2 focus:ring-ctp-sky/40 data-invalid:border-ctp-red data-invalid:focus:ring-ctp-red/40" />
				<NumberFieldPrimative.IncrementTrigger
					aria-label="Increment"
					class="absolute top-1 right-1 cursor-pointer rounded-t-sm transition-colors hover:bg-ctp-surface1"
				>
					<ChevronUp size={15} />
				</NumberFieldPrimative.IncrementTrigger>
				<NumberFieldPrimative.DecrementTrigger
					aria-label="Decrement"
					class="absolute right-1 bottom-1 cursor-pointer rounded-b-sm transition-colors hover:bg-ctp-surface1"
				>
					<ChevronDown size={15} />
				</NumberFieldPrimative.DecrementTrigger>
			</div>
			<NumberFieldPrimative.ErrorMessage class="text-ctp-red text-xs">
				{field()
					.state.meta.errors.map((err) => err.message)
					.join(",")}
			</NumberFieldPrimative.ErrorMessage>
		</NumberFieldPrimative>
	);
}
