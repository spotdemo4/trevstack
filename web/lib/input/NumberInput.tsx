import {
  NumberInput as ArkNumberInput,
  type NumberInputRootProps,
} from "@ark-ui/solid/number-input";
import { ChevronDown, ChevronUp } from "lucide-solid";
import { type Component, type JSX, Show, splitProps } from "solid-js";
import { twMerge } from "tailwind-merge";

type NumberInputProps = NumberInputRootProps & {
  label?: string;
  onBlur?: JSX.FocusEventHandler<HTMLInputElement, FocusEvent>;
};

export const NumberInput: Component<NumberInputProps> = (props) => {
  const [local, rest] = splitProps(props, ["label", "onBlur", "class"]);

  return (
    <ArkNumberInput.Root {...rest} class={twMerge("flex flex-col gap-1.5", local.class)}>
      <Show when={local.label}>
        <ArkNumberInput.Label class="text-sm font-medium text-ctp-subtext1 data-invalid:text-ctp-red">
          {local.label}
        </ArkNumberInput.Label>
      </Show>
      <ArkNumberInput.Control class="relative isolate">
        <ArkNumberInput.Input
          onBlur={local.onBlur}
          class="w-full rounded-md border border-ctp-surface1 bg-ctp-base px-3 py-2 text-sm text-ctp-text transition-colors placeholder:text-ctp-overlay0 hover:border-ctp-surface2 focus:border-ctp-sky focus:ring-2 focus:ring-ctp-sky/40 focus:outline-none data-invalid:border-ctp-red data-invalid:focus:ring-ctp-red/40"
        />
        <div class="absolute top-0 right-1 bottom-0 flex w-4 flex-col justify-center">
          <ArkNumberInput.IncrementTrigger
            aria-label="Increment"
            class="cursor-pointer rounded transition-colors hover:bg-ctp-surface1"
          >
            <ChevronUp size={16} />
          </ArkNumberInput.IncrementTrigger>
          <ArkNumberInput.DecrementTrigger
            aria-label="Decrement"
            class="cursor-pointer rounded transition-colors hover:bg-ctp-surface1"
          >
            <ChevronDown size={16} />
          </ArkNumberInput.DecrementTrigger>
        </div>
      </ArkNumberInput.Control>
    </ArkNumberInput.Root>
  );
};
