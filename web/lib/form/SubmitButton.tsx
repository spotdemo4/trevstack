import { Button as ButtonPrimative } from "@kobalte/core/button";
import { useFormContext } from "./context";

export function SubmitButton(props: { label?: string }) {
	const form = useFormContext();

	return (
		<form.Subscribe
			selector={(state) => ({
				canSubmit: state.canSubmit,
				isSubmitting: state.isSubmitting,
			})}
			children={(state) => {
				return (
					<ButtonPrimative type="submit" disabled={!state().canSubmit}>
						{state().isSubmitting ? "..." : (props.label ?? "Submit")}
					</ButtonPrimative>
				);
			}}
		/>
	);
}
