import type { JSX } from "solid-js";
import { twMerge } from "tailwind-merge";

export function Card(props: { children?: JSX.Element; class?: string }) {
  return (
    <div
      class={twMerge(
        "rounded-xl border border-ctp-surface0 bg-ctp-mantle p-6 shadow-ctp-crust/40 shadow-lg",
        props.class,
      )}
    >
      {props.children}
    </div>
  );
}
