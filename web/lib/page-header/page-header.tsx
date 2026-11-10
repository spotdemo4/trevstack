import type { JSX } from "@solidjs/web";
import { type Component, Show } from "solid-js";
import { twMerge } from "tailwind-merge";

type PageHeaderProps = {
  title: string;
  /** The kind of RPC the page demonstrates, such as "Server streaming". */
  kind?: string;
  /** The RPC's declaration, shown as code. */
  rpc?: string;
  class?: string;
  children?: JSX.Element;
};

/** A page's title, with the RPC it demonstrates and a short description. */
export const PageHeader: Component<PageHeaderProps> = (props) => (
  <header class={twMerge("flex flex-col gap-2", props.class)}>
    <div class="flex flex-wrap items-center gap-x-3 gap-y-1">
      <h1 class="text-2xl font-bold">{props.title}</h1>
      <Show when={props.kind}>
        <span class="rounded-full bg-fill-subtle px-2.5 py-0.5 text-xs font-semibold text-foreground-muted">
          {props.kind}
        </span>
      </Show>
    </div>
    <Show when={props.rpc}>
      <code class="w-fit max-w-full overflow-x-auto rounded-md bg-sunken px-2 py-1 font-mono text-xs whitespace-nowrap text-foreground-muted">
        {props.rpc}
      </code>
    </Show>
    <Show when={props.children}>
      <p class="max-w-prose text-sm text-foreground-subtle">{props.children}</p>
    </Show>
  </header>
);
