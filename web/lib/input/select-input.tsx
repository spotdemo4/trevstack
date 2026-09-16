import { ChevronDown } from "$lib/icon";
import type { JSX } from "@solidjs/web";
import { For, Show, type Component, createUniqueId, omit } from "solid-js";
import { twMerge } from "tailwind-merge";

type SelectItem = {
  label: string;
  value: string;
  disabled?: boolean;
};

type SelectValueChangeDetails = {
  value: string[];
  items: SelectItem[];
};

type SelectInputProps = Omit<
  JSX.SelectHTMLAttributes<HTMLSelectElement>,
  "children" | "onChange" | "value"
> & {
  class?: string;
  invalid?: boolean;
  items: SelectItem[];
  label?: string;
  placeholder?: string;
  value?: string[];
  onValueChange?: (details: SelectValueChangeDetails) => void;
};

export const SelectInput: Component<SelectInputProps> = (props) => {
  const selectProps = omit(
    props,
    "class",
    "invalid",
    "items",
    "label",
    "placeholder",
    "value",
    "onValueChange",
    "id",
  );
  const generatedId = createUniqueId();
  const selectId = () => props.id ?? generatedId;
  const isMulti = () => selectProps.multiple ?? false;
  const singleValue = () => props.value?.[0] ?? "";

  return (
    <div class="flex flex-col gap-1.5">
      <Show when={props.label}>
        <label for={selectId()} class="text-sm font-medium text-ctp-subtext1">
          {props.label}
        </label>
      </Show>
      <div class={twMerge("relative", props.class)}>
        <select
          {...selectProps}
          id={selectId()}
          value={isMulti() ? undefined : singleValue()}
          aria-invalid={props.invalid ? "true" : undefined}
          class={twMerge(
            "w-full rounded-md border border-ctp-surface1 bg-ctp-base px-3 py-2 text-sm text-ctp-text transition-colors hover:border-ctp-surface2 focus:border-ctp-sky focus:ring-2 focus:ring-ctp-sky/40 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-ctp-red aria-invalid:focus:ring-ctp-red/40",
            isMulti() ? "" : "appearance-none pr-8",
          )}
          onChange={(event) => {
            if (isMulti()) {
              const values = Array.from(event.currentTarget.selectedOptions)
                .map((option) => option.value)
                .filter(Boolean);
              const items = values
                .map((value) => props.items.find((item) => item.value === value))
                .filter((item): item is SelectItem => item !== undefined);
              props.onValueChange?.({ value: values, items });
              return;
            }

            const value = event.currentTarget.value;
            const item = props.items.find((candidate) => candidate.value === value);
            props.onValueChange?.({
              value: value ? [value] : [],
              items: item ? [item] : [],
            });
          }}
        >
          <Show when={!isMulti()}>
            <option value="" disabled={selectProps.required}>
              {props.placeholder ?? "Select an option"}
            </option>
          </Show>
          <For each={props.items}>
            {(item) => (
              <option
                value={item.value}
                disabled={item.disabled}
                selected={isMulti() ? props.value?.includes(item.value) : undefined}
              >
                {item.label}
              </option>
            )}
          </For>
        </select>
        <Show when={!isMulti()}>
          <ChevronDown
            size={16}
            class="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-ctp-subtext0"
          />
        </Show>
      </div>
    </div>
  );
};
