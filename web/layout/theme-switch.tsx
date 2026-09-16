import { Button } from "$lib/button";
import { Swap } from "@ark-ui/solid/swap";
import Moon from "lucide-solid/icons/moon";
import Sun from "lucide-solid/icons/sun";
import { createSignal } from "solid-js";

import styles from "./theme-switch.module.css";

function toggleTheme(dark: boolean) {
  if (dark) {
    document.querySelector("#theme-color")?.setAttribute("content", "#89dceb");
    document.documentElement.classList.add("dark");
    document.documentElement.classList.add("mocha");
    document.documentElement.classList.remove("light");
    document.documentElement.classList.remove("latte");
    localStorage.theme = "dark";
  } else {
    document.querySelector("#theme-color")?.setAttribute("content", "#04a5e5");
    document.documentElement.classList.add("light");
    document.documentElement.classList.add("latte");
    document.documentElement.classList.remove("dark");
    document.documentElement.classList.remove("mocha");
    localStorage.theme = "light";
  }
}

export const ThemeSwitch = () => {
  const [dark, setDark] = createSignal(
    localStorage.theme === "dark" ||
      (!("theme" in localStorage) && window.matchMedia("(prefers-color-scheme: dark)").matches),
  );

  return (
    <Button.Icon
      type="button"
      onClick={() => {
        const next = !dark();
        setDark(next);
        toggleTheme(next);
      }}
    >
      <Swap.Root swap={dark()} class={styles.Root}>
        <Swap.Indicator type="on" class={styles.FadeIndicator}>
          <Sun />
        </Swap.Indicator>
        <Swap.Indicator type="off" class={styles.FadeIndicator}>
          <Moon />
        </Swap.Indicator>
      </Swap.Root>
    </Button.Icon>
  );
};
