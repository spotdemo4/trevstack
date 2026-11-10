import type { LiveStream } from "$lib/effect";
import { errorDescription } from "$lib/effect";
import { type Component, Match, Show, Switch } from "solid-js";
import { twMerge } from "tailwind-merge";

/** The state of a live stream, with a button that restarts it once it stops. */
export const LiveStatus: Component<{ stream: LiveStream; class?: string }> = (props) => (
  <div
    role="status"
    class={twMerge("flex min-h-8 flex-wrap items-center gap-2 text-sm", props.class)}
  >
    <Switch>
      <Match when={props.stream.status() === "live"}>
        <span class="size-2 animate-pulse rounded-full bg-success" aria-hidden="true" />
        <span class="text-foreground-muted">Live</span>
      </Match>
      <Match when={props.stream.status() === "ended"}>
        <span class="size-2 rounded-full bg-foreground-faint" aria-hidden="true" />
        <span class="text-foreground-muted">Ended</span>
      </Match>
      <Match when={props.stream.status() === "failed"}>
        <span class="size-2 rounded-full bg-danger" aria-hidden="true" />
        <span class="text-danger">{errorDescription(props.stream.error())}</span>
        <Show when={props.stream.retrying()}>
          <span class="text-foreground-subtle">Reconnecting...</span>
        </Show>
      </Match>
    </Switch>
    <Show when={props.stream.status() !== "live"}>
      <button
        type="button"
        class="rounded-md bg-fill px-2.5 py-1 text-xs font-semibold text-foreground transition-colors hover:cursor-pointer hover:bg-fill-strong"
        onClick={() => props.stream.restart()}
      >
        Reconnect
      </button>
    </Show>
  </div>
);
