import { A } from "@solidjs/router";
import { type ComponentProps, type JSX, splitProps } from "solid-js";
import { Dynamic } from "solid-js/web";
import { twMerge } from "tailwind-merge";

type LayoutNavLinkAsRouterProps = ComponentProps<typeof A> & {
  as?: typeof A;
};

type LayoutNavLinkAsAnchorProps = JSX.AnchorHTMLAttributes<HTMLAnchorElement> & {
  as: "a";
  activeClass?: never;
  end?: never;
  inactiveClass?: never;
};

type LayoutNavLinkProps = LayoutNavLinkAsRouterProps | LayoutNavLinkAsAnchorProps;

const baseClass =
  "inline-flex px-1 text-sm text-ctp-subtext0 transition-colors hover:text-ctp-text";
const activeClass = "text-ctp-text";

const LayoutNavLink = (props: LayoutNavLinkProps) => {
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

export default LayoutNavLink;
