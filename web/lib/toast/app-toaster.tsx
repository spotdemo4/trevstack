import { X } from "$lib/icon";
import { Toaster } from "@trev.zip/solid-toast";
import { Show, type Component } from "solid-js";

import { renderToastIcon } from "./icon";

import styles from "./toast.module.css";

export const AppToaster: Component = () => (
  <Toaster
    gap={12}
    classes={{
      toast: `${styles.toastRoot} flex items-start gap-3 overflow-hidden rounded-lg border p-3 text-foreground shadow-lg shadow-shadow/35`,
      content: "flex min-w-0 flex-1 items-start gap-3",
      dismiss: `${styles.toastDismiss} mt-0.5 ml-1 inline-flex h-7 w-7 shrink-0 cursor-pointer items-center justify-center rounded-md text-foreground-subtle transition-colors hover:bg-fill-subtle/75 hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring/40`,
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
            <div class="mt-1 text-sm leading-5 text-foreground-muted">{toast.description}</div>
          </Show>
          <Show when={toast.action}>
            {(action) => (
              <button
                type="button"
                class="mt-2 inline-flex h-7 cursor-pointer items-center rounded-md bg-primary px-3 text-sm font-semibold text-primary-foreground transition-colors hover:bg-primary-hover focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none"
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
