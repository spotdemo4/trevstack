import { children, type JSX, Show } from "solid-js";
import { twMerge } from "tailwind-merge";
import { useFormContext } from "./context";

export function Form(props: { children?: JSX.Element; class?: string }) {
	const form = useFormContext();
	const resolved = children(() => props.children);

	return (
		<form.Subscribe
			selector={(state) => ({
				errors: state.errors,
			})}
		>
			{(state) => (
				<form
					onSubmit={async (e) => {
						e.preventDefault();
						e.stopPropagation();

						if (document.activeElement instanceof HTMLElement) {
							document.activeElement.blur();
						}

						await form.handleSubmit();
						form.validate("change");
					}}
					class={twMerge("flex w-full flex-col gap-4", props.class)}
				>
					{resolved()}
					<Show when={state().errors.length > 0}>
						<span class="text-ctp-red text-xs">
							{findError(state().errors)}
						</span>
					</Show>
				</form>
			)}
		</form.Subscribe>
	);
}

// Recursively find the first error message in the errors object without a path
function findError(errors: unknown): string {
	if (typeof errors === "object" && errors !== null) {
		if (
			"message" in errors &&
			typeof errors.message === "string" &&
			!("path" in errors)
		) {
			return errors.message;
		}

		for (const value of Object.values(errors)) {
			const message = findError(value);
			if (message !== "") {
				return message;
			}
		}
	} else if (Array.isArray(errors)) {
		for (const err of errors) {
			const message = findError(err);
			if (message !== "") {
				return message;
			}
		}
	}

	return "";
}
