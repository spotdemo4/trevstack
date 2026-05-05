import { Field } from "@ark-ui/solid/field";
import { NumberInput } from "@ark-ui/solid/number-input";
import { ChevronDown, ChevronUp } from "lucide-solid";
import { createMemo, For, Show } from "solid-js";

import { useFieldContext } from "./context";

export function NumberField(props: { label?: string }) {
  const field = useFieldContext<number>();
  const name = field().name;
  const errors = createMemo(() => [
    ...new Set(field().state.meta.errors.map((err) => err.message as string)),
  ]);

  return (
    <Field.Root
      invalid={!(field().state.meta.isValid || !field().state.meta.isBlurred)}
      class="flex flex-col gap-1.5"
    >
      <NumberInput.Root
        name={name}
        value={field().state.value != undefined ? String(field().state.value) : ""}
        class="flex flex-col gap-1.5"
        onValueChange={(c) => {
          // remove undefined numbers from form state
          if (Number.isNaN(c.valueAsNumber) && field().state.value !== undefined) {
            field().form.resetField(name);
          }

          // only handle change if the value is a valid number
          if (!Number.isNaN(c.valueAsNumber)) {
            field().handleChange(c.valueAsNumber);
          } else {
            // still validate
            field().form.validateField(name, "change");
          }
        }}
      >
        <Show when={props.label}>
          <NumberInput.Label class="text-sm font-medium text-ctp-subtext1 data-invalid:text-ctp-red">
            {props.label}
          </NumberInput.Label>
        </Show>
        <NumberInput.Control class="relative isolate">
          <NumberInput.Input
            onBlur={field().handleBlur}
            class="w-full rounded-md border border-ctp-surface1 bg-ctp-base px-3 py-2 text-sm text-ctp-text transition-colors placeholder:text-ctp-overlay0 hover:border-ctp-surface2 focus:border-ctp-sky focus:ring-2 focus:ring-ctp-sky/40 focus:outline-none data-invalid:border-ctp-red data-invalid:focus:ring-ctp-red/40"
          />
          <div class="absolute top-0 right-1 bottom-0 flex w-4 flex-col justify-center">
            <NumberInput.IncrementTrigger
              aria-label="Increment"
              class="cursor-pointer rounded transition-colors hover:bg-ctp-surface1"
            >
              <ChevronUp size={16} />
            </NumberInput.IncrementTrigger>
            <NumberInput.DecrementTrigger
              aria-label="Decrement"
              class="cursor-pointer rounded transition-colors hover:bg-ctp-surface1"
            >
              <ChevronDown size={16} />
            </NumberInput.DecrementTrigger>
          </div>
        </NumberInput.Control>
      </NumberInput.Root>
      <For each={errors()}>
        {(err) => <Field.ErrorText class="text-xs text-ctp-red">{err}</Field.ErrorText>}
      </For>
    </Field.Root>
  );
}
