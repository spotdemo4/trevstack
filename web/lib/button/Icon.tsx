import { type Component, type JSX, splitProps } from "solid-js";
import { Dynamic } from "solid-js/web";
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

const Icon: Component<IconProps> = (props) => {
  const [local, rest] = splitProps(props, ["as", "class", "children"]);
  const component = local.as ?? "button";

  return (
    <Dynamic component={component} class={twMerge(styles.Icon, local.class)} {...rest}>
      {local.children}
    </Dynamic>
  );
};

export default Icon;
