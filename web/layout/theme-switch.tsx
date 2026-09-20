import { Button } from "$lib/button";
import { Moon, Sun } from "$lib/icon";
import { createSignal } from "solid-js";

import styles from "./theme-switch.module.css";

function toggleTheme(dark: boolean) {
  if (dark) {
    document.querySelector("#theme-color")?.setAttribute("content", "#89dceb");
    document.documentElement.classList.add("dark");
    document.documentElement.classList.add("mocha");
    document.documentElement.classList.remove("light");
    document.documentElement.classList.remove("latte");
  } else {
    document.querySelector("#theme-color")?.setAttribute("content", "#04a5e5");
    document.documentElement.classList.add("light");
    document.documentElement.classList.add("latte");
    document.documentElement.classList.remove("dark");
    document.documentElement.classList.remove("mocha");
  }

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
