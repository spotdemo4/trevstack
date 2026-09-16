import { X } from "$lib/icon";
import { For, Show, type Component } from "solid-js";

import { renderToastIcon } from "./icon";
import { toaster } from "./toaster";
import { resolveToastTone } from "./tone";

import styles from "./toast.module.css";

export const AppToaster: Component = () => {
  return (
    <div class="pointer-events-none fixed right-5 bottom-5 z-100 flex w-[min(92vw,24rem)] flex-col gap-3">
      <For each={toaster.toasts()}>
        {(toast) => {
          const tone = () => resolveToastTone(toast.type);

          return (
            <div
              role={toast.type === "error" ? "alert" : "status"}
              aria-atomic="true"
              data-state="open"
              class={`${styles.toastRoot} pointer-events-auto relative flex w-full items-start gap-3 overflow-hidden rounded-lg border p-3 text-ctp-text shadow-lg shadow-ctp-crust/35 ${tone().root}`}
            >
              <div class={`flex h-6 w-6 shrink-0 items-center justify-center ${tone().icon}`}>
                {renderToastIcon(toast.type)}
              </div>

              <div class="min-w-0 flex-1">
                <div class={`truncate text-sm leading-5 font-semibold ${tone().title}`}>
                  {toast.title ?? "Notification"}
                </div>
                <Show when={toast.description}>
                  <div class="mt-1 text-sm leading-5 text-ctp-subtext1">{toast.description}</div>
                </Show>
              </div>

              <Show when={toast.closable}>
                <button
                  type="button"
                  aria-label="Dismiss notification"
                  class="mt-0.5 ml-1 inline-flex h-7 w-7 shrink-0 cursor-pointer items-center justify-center rounded-md text-ctp-subtext0 transition-colors hover:bg-ctp-surface0/75 hover:text-ctp-text focus-visible:ring-2 focus-visible:ring-ctp-sky/40 focus-visible:outline-none"
                  onClick={() => toaster.dismiss(toast.id)}
                >
                  <X size={14} />
                </button>
              </Show>

              <div
                class={`pointer-events-none absolute inset-x-0 bottom-0 h-0.5 ${tone().progress}`}
              />
            </div>
          );
        }}
      </For>
    </div>
  );
};
