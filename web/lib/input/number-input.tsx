import type { JSX } from "@solidjs/web";
import { type Component, createEffect, omit } from "solid-js";
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
  const inputProps = omit(props, "onValueChange", "class", "invalid", "value");
  let inputRef: HTMLInputElement | undefined;

  // The value is written only when it differs from the number the input holds, so it doesn't
  // rewrite what is being typed: "0." or "-" hold no number yet, which a field reports as unset,
  // and "0.50" holds 0.5.
  createEffect(
    () => props.value ?? "",
    (value) => {
      if (!inputRef) return;
      const current = inputRef.valueAsNumber;
      const same = Number.isNaN(current) ? value === "" : value !== "" && Number(value) === current;
      if (!same) inputRef.value = value;
    },
  );

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
      inputmode={props.inputmode ?? "numeric"}
      aria-invalid={props.invalid ? "true" : undefined}
      onInput={emitValue}
      class={twMerge(
        "h-9.5 w-full rounded-md border border-line bg-background px-3 py-2 text-sm text-foreground transition-colors placeholder:text-foreground-faint hover:border-line-strong focus:border-ring focus:ring-2 focus:ring-ring/40 focus:outline-none aria-invalid:border-danger aria-invalid:focus:ring-danger/40",
        props.class,
      )}
    />
  );
};
