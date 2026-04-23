import { children, type JSX } from "solid-js";
import { twMerge } from "tailwind-merge";
import { useFormContext } from "./context";

export function Form(props: { children?: JSX.Element; class?: string }) {
	const form = useFormContext();
	const resolved = children(() => props.children);

	return (
		<form
			onSubmit={async (e) => {
				e.preventDefault();
				e.stopPropagation();
				form.handleSubmit();
			}}
			class={twMerge("flex w-full flex-col gap-4", props.class)}
		>
			{resolved()}
		</form>
	);
}
