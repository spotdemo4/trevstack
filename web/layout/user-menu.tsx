import { session } from "$lib/auth";
import { Button } from "$lib/button";
import { AuthClient } from "$lib/connect";
import { LoaderCircle, User } from "$lib/icon";
import { toaster } from "$lib/toast";
import { useNavigate } from "@solidjs/router";
import { Effect } from "effect";
import { type Component, Show, createSignal, createUniqueId } from "solid-js";

import styles from "./user-menu.module.css";

export const UserMenu: Component = () => {
  const navigate = useNavigate();
  const subject = () => session.claims()?.sub ?? "";
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
            session.clear();
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
    <Show when={session.claims()}>
      <Button.Icon
        type="button"
        aria-label={`Open account actions for ${subject()}`}
        aria-controls={popoverId}
        aria-expanded={isOpen() ? "true" : "false"}
        popovertarget={popoverId}
        class={`${styles.Trigger} ${isOpen() ? styles.TriggerOpen : ""}`}
      >
        <User />
        <span class="hidden max-w-32 truncate text-sm md:block">{subject()}</span>
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
        <p class="truncate px-3 py-2 text-sm text-ctp-subtext0" title={subject()}>
          {subject()}
        </p>
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
