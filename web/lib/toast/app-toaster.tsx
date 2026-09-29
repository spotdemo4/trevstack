import { X } from "$lib/icon";
import { Toaster } from "@trev.zip/solid-toast";
import { Show, type Component } from "solid-js";

import { renderToastIcon } from "./icon";

import styles from "./toast.module.css";

export const AppToaster: Component = () => (
  <Toaster
    gap={12}
    classes={{
      toast: `${styles.toastRoot} flex items-start gap-3 overflow-hidden rounded-lg border p-3 text-ctp-text shadow-lg shadow-ctp-crust/35`,
      content: "flex min-w-0 flex-1 items-start gap-3",
      dismiss: `${styles.toastDismiss} mt-0.5 ml-1 inline-flex h-7 w-7 shrink-0 cursor-pointer items-center justify-center rounded-md text-ctp-subtext0 transition-colors hover:bg-ctp-surface0/75 hover:text-ctp-text focus-visible:ring-2 focus-visible:ring-ctp-sky/40`,
    }}
    dismissContent={<X size={14} />}
    renderToast={(toast, { remainingPercent, runAction }) => (
      <>
        <div
          class={styles.toastProgress}
          style={{ transform: `scaleX(${(remainingPercent() ?? 100) / 100})` }}
          aria-hidden="true"
        />
        <div
          class={`${styles.toastIcon} flex h-6 w-6 shrink-0 items-center justify-center`}
          aria-hidden="true"
        >
          {renderToastIcon(toast.type)}
        </div>
        <div class="min-w-0 flex-1 self-center">
          <div class="truncate text-sm leading-5 font-semibold">
            {toast.title ?? "Notification"}
          </div>
          <Show when={toast.description}>
            <div class="mt-1 text-sm leading-5 text-ctp-subtext1">{toast.description}</div>
          </Show>
          <Show when={toast.action}>
            {(action) => (
              <button
                type="button"
                class="mt-2 inline-flex h-7 cursor-pointer items-center rounded-md bg-ctp-sky px-3 text-sm font-semibold text-ctp-base transition-colors hover:bg-ctp-sapphire focus-visible:ring-2 focus-visible:ring-ctp-sky/50 focus-visible:outline-none"
                onClick={runAction}
              >
                {action().label}
              </button>
            )}
          </Show>
        </div>
      </>
    )}
  />
);
