import { children, type JSX } from "solid-js";
import { useFormContext } from "./context";

export function Form(props: { children?: JSX.Element }) {
	const form = useFormContext();
	const resolved = children(() => props.children);

	return (
		<form
			onSubmit={async (e) => {
				e.preventDefault();
				e.stopPropagation();
				form.handleSubmit();
			}}
			class="flex w-full max-w-sm flex-col gap-4 rounded-xl border border-ctp-surface0 bg-ctp-mantle p-6 shadow-ctp-crust/40 shadow-lg"
		>
			{resolved()}
		</form>
	);
}
