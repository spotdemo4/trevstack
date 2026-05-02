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
        value={field().state.value ? String(field().state.value) : ""}
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
          <NumberInput.Label class="text-ctp-subtext1 data-invalid:text-ctp-red text-sm font-medium">
            {props.label}
          </NumberInput.Label>
        </Show>
        <NumberInput.Control class="relative isolate">
          <NumberInput.Input
            onBlur={field().handleBlur}
            class="border-ctp-surface1 bg-ctp-base text-ctp-text placeholder:text-ctp-overlay0 hover:border-ctp-surface2 focus:border-ctp-sky focus:ring-ctp-sky/40 data-invalid:border-ctp-red data-invalid:focus:ring-ctp-red/40 w-full rounded-md border px-3 py-2 text-sm transition-colors focus:ring-2 focus:outline-none"
          />
          <div class="absolute top-0 right-1 bottom-0 flex w-4 flex-col justify-center">
            <NumberInput.IncrementTrigger
              aria-label="Increment"
              class="hover:bg-ctp-surface1 cursor-pointer rounded transition-colors"
            >
              <ChevronUp size={16} />
            </NumberInput.IncrementTrigger>
            <NumberInput.DecrementTrigger
              aria-label="Decrement"
              class="hover:bg-ctp-surface1 cursor-pointer rounded transition-colors"
            >
              <ChevronDown size={16} />
            </NumberInput.DecrementTrigger>
          </div>
        </NumberInput.Control>
      </NumberInput.Root>
      <For each={errors()}>
        {(err) => <Field.ErrorText class="text-ctp-red text-xs">{err}</Field.ErrorText>}
      </For>
    </Field.Root>
  );
}
