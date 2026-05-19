import { createListCollection } from "@ark-ui/solid/collection";
import { Select as ArkSelect, type SelectRootProps } from "@ark-ui/solid/select";
import { ChevronDownIcon } from "lucide-solid";
import {
  type Component,
  createMemo,
  createSignal,
  For,
  onCleanup,
  Show,
  splitProps,
} from "solid-js";
import { Portal } from "solid-js/web";
import { twMerge } from "tailwind-merge";

type SelectItem = {
  label: string;
  value: string;
  disabled?: boolean;
};

type SelectInputProps = Omit<SelectRootProps<SelectItem>, "collection"> & {
  class?: string;
  label?: string;
  items: SelectItem[];
  placeholder?: string;
};

function createMediaQuery(query: string) {
  const mql = window.matchMedia(query);
  const [matches, setMatches] = createSignal(mql.matches);
  const handler = (e: MediaQueryListEvent) => setMatches(e.matches);
  mql.addEventListener("change", handler);
  onCleanup(() => mql.removeEventListener("change", handler));
  return matches;
}

export const SelectInput: Component<SelectInputProps> = (props) => {
  const isCoarsePointer = createMediaQuery("(pointer: coarse)");

  return (
    <Show when={isCoarsePointer()} fallback={<DesktopSelectInput {...props} />}>
      <NativeSelectInput {...props} />
    </Show>
  );
};

const NativeSelectInput: Component<SelectInputProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "label", "items", "placeholder"]);
  const isMulti = () => rest.multiple ?? false;
  const singleValue = () => rest.value?.[0] ?? "";

  return (
    <div class="flex flex-col gap-1.5">
      <Show when={local.label}>
        <label class="text-sm font-medium text-ctp-subtext1">{local.label}</label>
      </Show>
      <div onFocusOut={rest.onBlur} class={twMerge("relative", local.class)}>
        <select
          name={rest.name}
          multiple={rest.multiple}
          value={isMulti() ? undefined : singleValue()}
          disabled={rest.disabled}
          required={rest.required}
          class={twMerge(
            "w-full rounded-md border border-ctp-surface1 bg-ctp-base px-3 py-2 text-sm text-ctp-text transition-colors hover:border-ctp-surface2 focus:border-ctp-sky focus:ring-2 focus:ring-ctp-sky/40 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50",
            isMulti() ? "" : "appearance-none pr-8",
          )}
          onChange={(e) => {
            if (isMulti()) {
              const values = Array.from(e.currentTarget.selectedOptions)
                .map((o) => o.value)
                .filter((v) => v !== "");
              const items = values
                .map((v) => local.items.find((i) => i.value === v))
                .filter((i): i is SelectItem => i !== undefined);
              rest.onValueChange?.({ value: values, items });
            } else {
              const v = e.currentTarget.value;
              const item = local.items.find((i) => i.value === v);
              rest.onValueChange?.({
                value: v ? [v] : [],
                items: item ? [item] : [],
              });
            }
          }}
        >
          <Show when={!isMulti()}>
            <option value="" disabled>
              {local.placeholder ?? "Select an option"}
            </option>
          </Show>
          <For each={local.items}>
            {(item) => (
              <option
                value={item.value}
                disabled={item.disabled}
                selected={isMulti() ? rest.value?.includes(item.value) : undefined}
              >
                {item.label}
              </option>
            )}
          </For>
        </select>
        <Show when={!isMulti()}>
          <ChevronDownIcon
            size={16}
            class="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-ctp-subtext0"
          />
        </Show>
      </div>
    </div>
  );
};

const DesktopSelectInput: Component<SelectInputProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "label", "items", "placeholder"]);

  const collection = createMemo(() =>
    createListCollection<SelectItem>({
      items: local.items,
      itemToValue: (item) => item.value,
      itemToString: (item) => item.label,
      isItemDisabled: (item) => item.disabled ?? false,
    }),
  );

  return (
    <ArkSelect.Root collection={collection()} positioning={{ sameWidth: true }} {...rest}>
      <Show when={local.label}>
        <ArkSelect.Label class="text-sm font-medium text-ctp-subtext1 data-invalid:text-ctp-red">
          {local.label}
        </ArkSelect.Label>
      </Show>
      <ArkSelect.Control class={twMerge("relative", local.class)}>
        <ArkSelect.Trigger class="flex w-full items-center justify-between gap-2 rounded-md border border-ctp-surface1 bg-ctp-base px-3 py-2 text-sm text-ctp-text transition-colors hover:border-ctp-surface2 focus:border-ctp-sky focus:ring-2 focus:ring-ctp-sky/40 focus:outline-none data-invalid:border-ctp-red data-invalid:focus:ring-ctp-red/40">
          <ArkSelect.ValueText placeholder={local.placeholder ?? "Select an option"} />
          <ChevronDownIcon size={16} class="text-ctp-subtext0" />
        </ArkSelect.Trigger>
      </ArkSelect.Control>
      <Portal>
        <ArkSelect.Positioner>
          <ArkSelect.Content class="z-50 max-h-64 overflow-y-auto rounded-md border border-ctp-surface1 bg-ctp-base p-1 text-ctp-text shadow-lg">
            <For each={local.items}>
              {(item) => (
                <ArkSelect.Item
                  item={item}
                  class="flex cursor-pointer items-center rounded px-2 py-1.5 text-sm transition-colors hover:bg-ctp-surface1 data-disabled:cursor-not-allowed data-disabled:opacity-50 data-highlighted:bg-ctp-surface1"
                >
                  <ArkSelect.ItemText>{item.label}</ArkSelect.ItemText>
                </ArkSelect.Item>
              )}
            </For>
          </ArkSelect.Content>
        </ArkSelect.Positioner>
      </Portal>
      <ArkSelect.HiddenSelect />
    </ArkSelect.Root>
  );
};
