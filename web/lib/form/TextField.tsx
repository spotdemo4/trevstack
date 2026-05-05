import { Field } from "@ark-ui/solid/field";
import { createMemo, For, Show } from "solid-js";

import { useFieldContext } from "./context";

export function TextField(props: { label?: string }) {
  const field = useFieldContext<string>();
  const name = field().name;
  const errors = createMemo(() => [
    ...new Set(field().state.meta.errors.map((err) => err.message as string)),
  ]);

  return (
    <Field.Root
      invalid={!(field().state.meta.isValid || !field().state.meta.isBlurred)}
      class="flex flex-col gap-1.5"
    >
      <Show when={props.label}>
        <Field.Label class="text-sm font-medium text-ctp-subtext1 data-invalid:text-ctp-red">
          {props.label}
        </Field.Label>
      </Show>
      <Field.Input
        name={name}
        value={field().state.value ?? ""}
        class="rounded-md border border-ctp-surface1 bg-ctp-base px-3 py-2 text-sm text-ctp-text transition-colors placeholder:text-ctp-overlay0 hover:border-ctp-surface2 focus:border-ctp-sky focus:ring-2 focus:ring-ctp-sky/40 focus:outline-none data-invalid:border-ctp-red data-invalid:focus:ring-ctp-red/40"
        onInput={(e) => field().handleChange(e.target.value)}
        onBlur={field().handleBlur}
      />
      <For each={errors()}>
        {(err) => <Field.ErrorText class="text-xs text-ctp-red">{err}</Field.ErrorText>}
      </For>
    </Field.Root>
  );
}
