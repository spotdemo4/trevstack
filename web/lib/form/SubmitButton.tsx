import { Button as ButtonPrimative } from "@kobalte/core/button";
import { useFormContext } from "./context";

export function SubmitButton(props: { label?: string }) {
	const form = useFormContext();

	return (
		<form.Subscribe
			selector={(state) => ({
				canSubmit: state.canSubmit,
				isSubmitting: state.isSubmitting,
				isTouched: state.isTouched,
			})}
			children={(state) => {
				return (
					<ButtonPrimative
						type="submit"
						disabled={!state().canSubmit || !state().isTouched}
						class="mt-2 inline-flex items-center justify-center rounded-md bg-ctp-sky px-4 py-2 text-sm font-semibold text-ctp-base shadow-sm transition-colors hover:cursor-pointer hover:bg-ctp-sapphire focus:outline-none focus:ring-2 focus:ring-ctp-sky/50 focus:ring-offset-2 focus:ring-offset-ctp-mantle disabled:cursor-not-allowed disabled:bg-ctp-surface1 disabled:text-ctp-overlay0 disabled:shadow-none"
					>
						{state().isSubmitting ? "..." : (props.label ?? "Submit")}
					</ButtonPrimative>
				);
			}}
		/>
	);
}
