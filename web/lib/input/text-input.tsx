import type { JSX } from "@solidjs/web";
import type { Component } from "solid-js";
import { omit } from "solid-js";
import { twMerge } from "tailwind-merge";

type TextInputProps = Omit<JSX.InputHTMLAttributes<HTMLInputElement>, "class"> & {
  class?: string;
};

export const TextInput: Component<TextInputProps> = (props) => {
  const rest = omit(props, "class");

  return (
    <input
      {...rest}
      class={twMerge(
        "h-9.5 rounded-md border border-ctp-surface1 bg-ctp-base px-3 py-2 text-sm text-ctp-text transition-colors placeholder:text-ctp-overlay0 hover:border-ctp-surface2 focus:border-ctp-sky focus:ring-2 focus:ring-ctp-sky/40 focus:outline-none aria-invalid:border-ctp-red aria-invalid:focus:ring-ctp-red/40",
        props.class,
      )}
    />
  );
};
