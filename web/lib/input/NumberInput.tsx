import {
  NumberInput as ArkNumberInput,
  type NumberInputRootProps,
} from "@ark-ui/solid/number-input";
import { MinusIcon, PlusIcon } from "lucide-solid";
import { type Component, type JSX, Show, splitProps } from "solid-js";
import { twMerge } from "tailwind-merge";

type NumberInputProps = NumberInputRootProps & {
  label?: string;
  onBlur?: JSX.FocusEventHandler<HTMLInputElement, FocusEvent>;
};

const iconTriggerClass =
  "inline-flex h-7 w-7 items-center justify-center rounded-md border border-ctp-surface1 bg-ctp-surface0 text-ctp-subtext0 shadow-sm transition-colors hover:cursor-pointer hover:border-ctp-surface2 hover:bg-ctp-surface1 hover:text-ctp-text focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ctp-sky/40 active:bg-ctp-surface2";

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
          class="w-full rounded-md border border-ctp-surface1 bg-ctp-base p-2 pr-20 pl-3 text-sm text-ctp-text transition-colors placeholder:text-ctp-overlay0 hover:border-ctp-surface2 focus:border-ctp-sky focus:ring-2 focus:ring-ctp-sky/40 focus:outline-none data-invalid:border-ctp-red data-invalid:focus:ring-ctp-red/40"
        />
        <div class="absolute top-1/2 right-1 flex -translate-y-1/2 items-center gap-1">
          <ArkNumberInput.DecrementTrigger aria-label="Decrement" class={iconTriggerClass}>
            <MinusIcon size={16} />
          </ArkNumberInput.DecrementTrigger>
          <ArkNumberInput.IncrementTrigger aria-label="Increment" class={iconTriggerClass}>
            <PlusIcon size={16} />
          </ArkNumberInput.IncrementTrigger>
        </div>
      </ArkNumberInput.Control>
    </ArkNumberInput.Root>
  );
};
