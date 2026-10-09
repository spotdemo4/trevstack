import { Button } from "$lib/button";
import { Moon, Sun } from "$lib/icon";
import { createSignal } from "solid-js";

import styles from "./theme-switch.module.css";

function toggleTheme(dark: boolean) {
  document.documentElement.classList.toggle("dark", dark);
  document.documentElement.classList.toggle("light", !dark);
  document
    .querySelector("#theme-color")
    ?.setAttribute(
      "content",
      getComputedStyle(document.documentElement).getPropertyValue("--color-primary").trim(),
    );

  try {
    localStorage.setItem("theme", dark ? "dark" : "light");
  } catch {
    // Theme changes still work when storage is unavailable.
  }
}

export const ThemeSwitch = () => {
  const [dark, setDark] = createSignal(document.documentElement.classList.contains("dark"));

  return (
    <Button.Icon
      type="button"
      aria-label={dark() ? "Switch to light theme" : "Switch to dark theme"}
      aria-pressed={dark() ? "true" : "false"}
      onClick={() => {
        const next = !dark();
        setDark(next);
        toggleTheme(next);
      }}
    >
      <span class={styles.Root}>
        <span
          aria-hidden="true"
          data-state={dark() ? "open" : "closed"}
          class={styles.FadeIndicator}
        >
          <Sun />
        </span>
        <span
          aria-hidden="true"
          data-state={dark() ? "closed" : "open"}
          class={styles.FadeIndicator}
        >
          <Moon />
        </span>
      </span>
    </Button.Icon>
  );
};
