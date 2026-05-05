import type { Component } from "solid-js";
import { twMerge } from "tailwind-merge";

import { useFormContext } from "./context";

type ResetButtonProps = {
  class?: string;
  label?: string;
};

export const ResetButton: Component<ResetButtonProps> = (props) => {
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
            class={twMerge(
              "inline-flex items-center justify-center rounded-md bg-ctp-sky px-4 py-2 text-sm font-semibold text-ctp-base shadow-sm transition-colors hover:cursor-pointer hover:bg-ctp-sapphire focus:ring-2 focus:ring-ctp-sky/50 focus:ring-offset-2 focus:ring-offset-ctp-mantle focus:outline-none disabled:cursor-not-allowed disabled:bg-ctp-surface1 disabled:text-ctp-overlay0 disabled:shadow-none",
              props.class,
            )}
            onClick={() => {
              form.reset();
            }}
          >
            {props.label ?? "Reset"}
          </button>
        );
      }}
    />
  );
};
