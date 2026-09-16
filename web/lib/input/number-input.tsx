import type { JSX } from "@solidjs/web";
import { type Component, omit } from "solid-js";
import { twMerge } from "tailwind-merge";

type NumberValueChangeDetails = {
  value: string;
  valueAsNumber: number;
};

type NumberInputProps = Omit<
  JSX.InputHTMLAttributes<HTMLInputElement>,
  "class" | "onInput" | "value"
> & {
  class?: string;
  value?: string;
  invalid?: boolean;
  onValueChange?: (details: NumberValueChangeDetails) => void;
};

export const NumberInput: Component<NumberInputProps> = (props) => {
  const inputProps = omit(props, "onValueChange", "class", "invalid");
  let inputRef: HTMLInputElement | undefined;

  const emitValue = () => {
    if (!inputRef) return;
    props.onValueChange?.({
      value: inputRef.value,
      valueAsNumber: inputRef.valueAsNumber,
    });
  };

  return (
    <input
      {...inputProps}
      ref={(node) => (inputRef = node)}
      type="number"
      inputmode="numeric"
      aria-invalid={props.invalid ? "true" : undefined}
      onInput={emitValue}
      class={twMerge(
        "h-9.5 w-full rounded-md border border-ctp-surface1 bg-ctp-base px-3 py-2 text-sm text-ctp-text transition-colors placeholder:text-ctp-overlay0 hover:border-ctp-surface2 focus:border-ctp-sky focus:ring-2 focus:ring-ctp-sky/40 focus:outline-none aria-invalid:border-ctp-red aria-invalid:focus:ring-ctp-red/40",
        props.class,
      )}
    />
  );
};
