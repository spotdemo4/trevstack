import { Ellipsis } from "$lib/icon";
import { Show, type Component } from "solid-js";
import { twMerge } from "tailwind-merge";

import type { FormController } from "./hook";

type SubmitButtonProps = {
  class?: string;
  form: FormController<any>;
  label?: string;
};

export const SubmitButton: Component<SubmitButtonProps> = (props) => (
  <button
    type="submit"
    disabled={!props.form.canSubmit()}
    class={twMerge(
      "inline-flex h-9.5 items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground shadow-sm transition-colors hover:cursor-pointer hover:bg-primary-hover focus:ring-2 focus:ring-ring/50 focus:ring-offset-2 focus:ring-offset-surface focus:outline-none disabled:cursor-not-allowed disabled:bg-fill disabled:text-foreground-faint disabled:shadow-none",
      props.class,
    )}
  >
    <Show when={props.form.isSubmitting()} fallback={props.label ?? "Submit"}>
      <Ellipsis class="animate-pulse" />
    </Show>
  </button>
);
