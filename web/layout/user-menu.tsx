import { session } from "$lib/auth";
import { Button } from "$lib/button";
import { User } from "$lib/icon";
import { toaster } from "$lib/toast";
import { useNavigate } from "@solidjs/router";
import { type Component, Show, createSignal, createUniqueId } from "solid-js";

import styles from "./user-menu.module.css";

type UserMenuProps = {
  onBeforeNavigate?: () => void;
};

export const UserMenu: Component<UserMenuProps> = (props) => {
  const navigate = useNavigate();
  const subject = () => session.claims()?.sub ?? "";
  const popoverId = createUniqueId();
  const anchorName = `--user-menu-${popoverId}`;
  const [isOpen, setIsOpen] = createSignal(false);
  let popover: HTMLDivElement | undefined;

  // The session token is only held by the browser, so forgetting it signs out.
  const logout = () => {
    popover?.hidePopover();
    props.onBeforeNavigate?.();
    session.clear();
    toaster.success({ title: "Signed out" });
    navigate("/auth", { replace: true });
  };

  return (
    <Show when={session.claims()}>
      <Button.Icon
        type="button"
        aria-label={`Open account actions for ${subject()}`}
        aria-controls={popoverId}
        aria-expanded={isOpen() ? "true" : "false"}
        popovertarget={popoverId}
        style={{ "anchor-name": anchorName }}
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
        style={{ "position-anchor": anchorName }}
        class={styles.Panel}
        onToggle={(event) => setIsOpen(event.newState === "open")}
      >
        <p class="truncate px-3 py-2 text-sm text-foreground-subtle" title={subject()}>
          {subject()}
        </p>
        <button type="button" class={styles.Action} onClick={logout}>
          Log out
        </button>
      </div>
    </Show>
  );
};
