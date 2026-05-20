import { A } from "@solidjs/router";
import { type ComponentProps, type JSX, splitProps } from "solid-js";
import { Dynamic } from "solid-js/web";
import { twMerge } from "tailwind-merge";

type NavLinkAsRouterProps = ComponentProps<typeof A> & {
  as?: typeof A;
};

type NavLinkAsAnchorProps = JSX.AnchorHTMLAttributes<HTMLAnchorElement> & {
  as: "a";
  activeClass?: never;
  end?: never;
  inactiveClass?: never;
};

type NavLinkProps = NavLinkAsRouterProps | NavLinkAsAnchorProps;

const baseClass =
  "inline-flex px-1 text-sm text-ctp-subtext0 transition-colors hover:text-ctp-text";
const activeClass = "text-ctp-text";

export const NavLink = (props: NavLinkProps) => {
  if (props.as === "a") {
    const [local, rest] = splitProps(props, ["as", "class"]);
    return <Dynamic component="a" {...rest} class={twMerge(baseClass, local.class)} />;
  }

  const [local, rest] = splitProps(props, ["class", "activeClass"]);

  return (
    <A
      {...rest}
      class={twMerge(baseClass, local.class)}
      activeClass={twMerge(activeClass, local.activeClass)}
    />
  );
};
