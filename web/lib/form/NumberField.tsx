import { NumberField as NumberFieldPrimative } from "@kobalte/core/number-field";
import { ChevronDown, ChevronUp } from "lucide-solid";
import { For, Show } from "solid-js";
import { useFieldContext } from "./context";

export function NumberField(props: { label?: string }) {
	const field = useFieldContext<number>();
	const name = field().name;

	return (
		<NumberFieldPrimative
			name={name}
			rawValue={field().state.value ?? undefined}
			onRawValueChange={(c) => {
				if (Number.isNaN(c) && field().state.value !== undefined) {
					field().form.resetField(name);
					field().form.validateField(name, "change");
				} else if (!Number.isNaN(c)) {
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
				<NumberFieldPrimative.Label class="font-medium text-ctp-subtext1 text-sm data-invalid:text-ctp-red">
					{props.label}
				</NumberFieldPrimative.Label>
			</Show>
			<div class="relative">
				<NumberFieldPrimative.Input
					onBlur={field().handleBlur}
					class="w-full rounded-md border border-ctp-surface1 bg-ctp-base px-3 py-2 text-ctp-text text-sm transition-colors placeholder:text-ctp-overlay0 hover:border-ctp-surface2 focus:border-ctp-sky focus:outline-none focus:ring-2 focus:ring-ctp-sky/40 data-invalid:border-ctp-red data-invalid:focus:ring-ctp-red/40"
				/>
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
			<For each={field().state.meta.errors}>
				{(err) => (
					<NumberFieldPrimative.ErrorMessage class="text-ctp-red text-xs">
						{err.message}
					</NumberFieldPrimative.ErrorMessage>
				)}
			</For>
		</NumberFieldPrimative>
	);
}
