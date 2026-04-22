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
		>
			{resolved()}
		</form>
	);
}
