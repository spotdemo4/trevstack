import { children, createMemo, For, type JSX } from "solid-js";
import { twMerge } from "tailwind-merge";
import { useFormContext } from "./context";

export function Form(props: { children?: JSX.Element; class?: string }) {
	const form = useFormContext();
	const child = children(() => props.children);

	return (
		<form.Subscribe
			selector={(state) => ({
				errors: state.errors,
			})}
		>
			{(state) => {
				const errors = createMemo(() => findErrors(state().errors));

				return (
					<form
						onSubmit={async (e) => {
							e.preventDefault();
							e.stopPropagation();

							if (document.activeElement instanceof HTMLElement) {
								document.activeElement.blur();
							}

							await form.handleSubmit();

							// Manually trigger validation after submission to update error messages
							form.validate("change");
						}}
						class={twMerge("flex w-full flex-col gap-4", props.class)}
					>
						{child()}
						<For each={errors()}>
							{(err) => <span class="text-ctp-red text-xs">{err}</span>}
						</For>
					</form>
				);
			}}
		</form.Subscribe>
	);
}

type Error = {
	message?: string;
	path?: string[];
};

type FieldErrors = {
	""?: Error[];
};

function findErrors(f: FieldErrors[]): string[] {
	const messages = new Set<string>();

	for (const field of f) {
		if (!field[""]) {
			continue;
		}

		for (const error of field[""]) {
			if (error.message && !error.path) {
				messages.add(error.message);
			}
		}
	}

	return Array.from(messages);
}
