import { createListCollection } from "@ark-ui/solid/collection";
import { Select as ArkSelect, type SelectRootProps } from "@ark-ui/solid/select";
import { ChevronDownIcon } from "lucide-solid";
import { type Component, createMemo, For, Show, splitProps } from "solid-js";
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

export const SelectInput: Component<SelectInputProps> = (props) => {
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
