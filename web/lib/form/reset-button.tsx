import type { Component } from "solid-js";
import { twMerge } from "tailwind-merge";

import type { FormController } from "./hook";

type ResetButtonProps = {
  class?: string;
  form: FormController<any>;
  label?: string;
};

export const ResetButton: Component<ResetButtonProps> = (props) => (
  <button
    type="button"
    disabled={props.form.isDefaultValue()}
    class={twMerge(
      "inline-flex h-9.5 items-center justify-center rounded-md bg-fill px-4 py-2 text-sm font-semibold text-foreground shadow-sm transition-colors hover:cursor-pointer hover:bg-fill-strong focus:ring-2 focus:ring-foreground-faint/50 focus:ring-offset-2 focus:ring-offset-surface focus:outline-none disabled:cursor-not-allowed disabled:bg-fill disabled:text-foreground-faint disabled:shadow-none",
      props.class,
    )}
    onClick={() => props.form.reset()}
  >
    {props.label ?? "Reset"}
  </button>
);
