import { useFormContext } from "./context";

export function ResetButton(props: { label?: string }) {
	const form = useFormContext();

	return (
		<form.Subscribe
			selector={(state) => ({
				isDefaultValue: state.isDefaultValue,
			})}
			children={(state) => {
				return (
					<button
						type="reset"
						disabled={state().isDefaultValue}
						class="mt-2 inline-flex items-center justify-center rounded-md bg-ctp-sky px-4 py-2 font-semibold text-ctp-base text-sm shadow-sm transition-colors hover:cursor-pointer hover:bg-ctp-sapphire focus:outline-none focus:ring-2 focus:ring-ctp-sky/50 focus:ring-offset-2 focus:ring-offset-ctp-mantle disabled:cursor-not-allowed disabled:bg-ctp-surface1 disabled:text-ctp-overlay0 disabled:shadow-none"
					>
						{props.label ?? "Reset"}
					</button>
				);
			}}
		/>
	);
}
