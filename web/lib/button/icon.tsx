import type { JSX } from "@solidjs/web";
import { Dynamic } from "@solidjs/web";
import type { Component } from "solid-js";
import { omit } from "solid-js";
import { twMerge } from "tailwind-merge";

import styles from "./button.module.css";

type IconAsButtonProps = JSX.ButtonHTMLAttributes<HTMLButtonElement> & {
  as?: "button";
};

type IconAsAnchorProps = JSX.AnchorHTMLAttributes<HTMLAnchorElement> & {
  as: "a";
};

type IconProps = (IconAsButtonProps | IconAsAnchorProps) & {
  children?: JSX.Element;
  class?: string;
};

export const Icon: Component<IconProps> = (props) => {
  const component = props.as ?? "button";
  const rest = omit(props, "as", "class", "children");

  return (
    <Dynamic component={component} class={twMerge(styles.Icon, props.class)} {...rest}>
      {props.children}
    </Dynamic>
  );
};
