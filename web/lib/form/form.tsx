import type { JSX } from "@solidjs/web";
import { For, type Component } from "solid-js";
import { twMerge } from "tailwind-merge";

import type { FormController } from "./hook";

type FormProps = {
  children?: JSX.Element;
  class?: string;
  form: FormController<any>;
};

export const Form: Component<FormProps> = (props) => (
  <form
    onSubmit={async (event) => {
      event.preventDefault();
      event.stopPropagation();

      if (document.activeElement instanceof HTMLElement) document.activeElement.blur();
      await props.form.handleSubmit();
    }}
    class={twMerge("flex w-full flex-col gap-4", props.class)}
  >
    {props.children}
    <For each={props.form.errors()}>
      {(error) => <span class="text-xs text-ctp-red">{error}</span>}
    </For>
  </form>
);
