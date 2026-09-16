import type { JSX } from "@solidjs/web";
import { type Component, omit } from "solid-js";
import { twMerge } from "tailwind-merge";

type DateInputProps = Omit<
  JSX.InputHTMLAttributes<HTMLInputElement>,
  "onInput" | "type" | "value"
> & {
  class?: string;
  invalid?: boolean;
  value?: string;
  onValueChange?: (value: string) => void;
};

export const DateInput: Component<DateInputProps> = (props) => {
  const inputProps = omit(props, "class", "invalid", "value", "onValueChange");

  return (
    <input
      {...inputProps}
      type="date"
      value={props.value ?? ""}
      aria-invalid={props.invalid ? "true" : undefined}
      class={twMerge(
        "h-9.5 min-w-42 rounded-md border border-ctp-surface1 bg-ctp-base px-3 text-sm text-ctp-text transition-colors hover:border-ctp-surface2 focus:border-ctp-sky focus:ring-2 focus:ring-ctp-sky/40 focus:outline-none aria-invalid:border-ctp-red aria-invalid:focus:ring-ctp-red/40",
        props.class,
      )}
      onInput={(event) => props.onValueChange?.(event.currentTarget.value)}
    />
  );
};
