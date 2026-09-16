import type { JSX } from "@solidjs/web";
import type { Component } from "solid-js";
import { twMerge } from "tailwind-merge";

type CardProps = {
  children?: JSX.Element;
  class?: string;
};

export const Card: Component<CardProps> = (props) => {
  return (
    <div
      class={twMerge(
        "rounded-xl border border-ctp-surface0 bg-ctp-mantle p-6 shadow-lg shadow-ctp-crust/40",
        props.class,
      )}
    >
      {props.children}
    </div>
  );
};
