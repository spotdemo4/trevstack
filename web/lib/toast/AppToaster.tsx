import { Toast, Toaster } from "@ark-ui/solid/toast";
import { X } from "lucide-solid";
import type { Component } from "solid-js";
import { Show } from "solid-js";
import { Portal } from "solid-js/web";

import { renderToastIcon } from "./icon";
import { toaster } from "./toaster";
import { resolveToastTone } from "./tone";

export const AppToaster: Component = () => {
  return (
    <Portal>
      <Toaster toaster={toaster}>
        {(toast) => {
          const tone = () => resolveToastTone(toast().type);

          return (
            <Toast.Root
              class={`text-ctp-text shadow-ctp-crust/35 data-[state=closed]:animate-toast-out data-[state=open]:animate-toast-in pointer-events-auto relative flex w-[min(92vw,24rem)] items-start gap-3 overflow-hidden rounded-lg border p-3 shadow-lg ${tone().root}`}
            >
              <div class={`flex h-6 w-6 shrink-0 items-center justify-center ${tone().icon}`}>
                {renderToastIcon(toast().type)}
              </div>

              <div class="min-w-0 flex-1">
                <Toast.Title class={`truncate text-sm leading-5 font-semibold ${tone().title}`}>
                  {toast().title ?? "Notification"}
                </Toast.Title>
                <Show when={toast().description}>
                  <Toast.Description class="text-ctp-subtext1 mt-1 text-sm leading-5">
                    {toast().description}
                  </Toast.Description>
                </Show>
                <Show when={toast().action?.label}>
                  <Toast.ActionTrigger class="border-ctp-surface1 bg-ctp-base hover:bg-ctp-surface0 focus-visible:ring-ctp-sky/40 mt-3 inline-flex items-center rounded-md border px-2.5 py-1.5 text-xs font-medium transition-colors focus-visible:ring-2 focus-visible:outline-none">
                    {toast().action?.label}
                  </Toast.ActionTrigger>
                </Show>
              </div>

              <Show when={toast().closable}>
                <Toast.CloseTrigger class="text-ctp-subtext0 hover:bg-ctp-surface0/75 hover:text-ctp-text focus-visible:ring-ctp-sky/40 mt-0.5 ml-1 inline-flex h-7 w-7 shrink-0 cursor-pointer items-center justify-center rounded-md transition-colors focus-visible:ring-2 focus-visible:outline-none">
                  <X size={14} />
                </Toast.CloseTrigger>
              </Show>

              <div
                class={`pointer-events-none absolute inset-x-0 bottom-0 h-0.5 ${tone().progress}`}
              />
            </Toast.Root>
          );
        }}
      </Toaster>
    </Portal>
  );
};
