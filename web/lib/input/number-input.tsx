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

const iconTriggerClass =
  "inline-flex h-7 w-7 touch-manipulation items-center justify-center rounded-md border border-ctp-surface1 bg-ctp-surface0 text-ctp-subtext0 shadow-sm transition-colors hover:cursor-pointer hover:border-ctp-surface2 hover:bg-ctp-surface1 hover:text-ctp-text focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ctp-sky/40 active:bg-ctp-surface2 disabled:cursor-not-allowed disabled:opacity-50";

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

  const step = (direction: -1 | 1) => {
    if (!inputRef) return;
    if (direction < 0) inputRef.stepDown();
    else inputRef.stepUp();
    emitValue();
  };

  return (
    <div class={twMerge("relative isolate", props.class)}>
      <input
        {...inputProps}
        ref={(node) => (inputRef = node)}
        type="number"
        inputmode="numeric"
        aria-invalid={props.invalid ? "true" : undefined}
        onInput={emitValue}
        class="h-9.5 w-full rounded-md border border-ctp-surface1 bg-ctp-base p-2 pr-20 pl-3 text-sm text-ctp-text transition-colors placeholder:text-ctp-overlay0 hover:border-ctp-surface2 focus:border-ctp-sky focus:ring-2 focus:ring-ctp-sky/40 focus:outline-none aria-invalid:border-ctp-red aria-invalid:focus:ring-ctp-red/40"
      />
      <div class="absolute top-1/2 right-1 flex -translate-y-1/2 items-center gap-1">
        <button
          type="button"
          aria-label="Decrement"
          class={iconTriggerClass}
          disabled={inputProps.disabled}
          onClick={() => step(-1)}
        >
          <span aria-hidden="true">−</span>
        </button>
        <button
          type="button"
          aria-label="Increment"
          class={iconTriggerClass}
          disabled={inputProps.disabled}
          onClick={() => step(1)}
        >
          <span aria-hidden="true">+</span>
        </button>
      </div>
    </div>
  );
};
