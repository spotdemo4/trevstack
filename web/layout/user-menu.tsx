import { Button } from "$lib/button";
import { AuthClient } from "$lib/connect";
import { LoaderCircle, User } from "$lib/icon";
import { toaster } from "$lib/toast";
import { useLocation, useNavigate } from "@solidjs/router";
import { Effect } from "effect";
import { type Component, Show, createSignal, createUniqueId } from "solid-js";

import styles from "./user-menu.module.css";

export const UserMenu: Component = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const popoverId = createUniqueId();
  const [isOpen, setIsOpen] = createSignal(false);
  const [isLoggingOut, setIsLoggingOut] = createSignal(false);
  let popover: HTMLDivElement | undefined;

  const logout = () => {
    if (isLoggingOut()) return;
    setIsLoggingOut(true);

    void AuthClient.logout({})
      .pipe(
        Effect.match({
          onSuccess: () => {
            popover?.hidePopover();
            toaster.success({ title: "Signed out" });
            navigate("/auth", { replace: true });
          },
          onFailure: (error) => {
            toaster.error({
              title: "Failed to sign out",
              description: error.message,
            });
          },
        }),
        Effect.runPromise,
      )
      .finally(() => setIsLoggingOut(false));
  };

  return (
    <Show when={!isPublicPath(normalizePath(location.pathname))}>
      <Button.Icon
        type="button"
        aria-label="Open account actions"
        aria-controls={popoverId}
        aria-expanded={isOpen() ? "true" : "false"}
        popovertarget={popoverId}
        class={isOpen() ? styles.TriggerOpen : undefined}
      >
        <User />
      </Button.Icon>
      <div
        ref={(element) => (popover = element)}
        id={popoverId}
        popover="auto"
        role="group"
        aria-label="Account actions"
        class={styles.Panel}
        onToggle={(event) => setIsOpen(event.newState === "open")}
      >
        <button type="button" class={styles.Action} disabled={isLoggingOut()} onClick={logout}>
          <Show when={isLoggingOut()} fallback="Log out">
            <LoaderCircle class="animate-spin" size={18} />
            Signing out…
          </Show>
        </button>
      </div>
    </Show>
  );
};

function isPublicPath(path: string) {
  return path === "/auth" || path === "/403";
}

function normalizePath(path: string) {
  return path.toLowerCase().replace(/\/+$/, "") || "/";
}
