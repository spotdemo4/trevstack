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
            class="bg-ctp-sky text-ctp-base hover:bg-ctp-sapphire focus:ring-ctp-sky/50 focus:ring-offset-ctp-mantle disabled:bg-ctp-surface1 disabled:text-ctp-overlay0 mt-2 inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-semibold shadow-sm transition-colors hover:cursor-pointer focus:ring-2 focus:ring-offset-2 focus:outline-none disabled:cursor-not-allowed disabled:shadow-none"
          >
            {props.label ?? "Reset"}
          </button>
        );
      }}
    />
  );
}
